package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrJobAlreadyRunning = errors.New("assistant chat is already running")

type JobStatus string

const (
	JobStatusRunning  JobStatus = "running"
	JobStatusDone     JobStatus = "done"
	JobStatusError    JobStatus = "error"
	JobStatusCanceled JobStatus = "canceled"
)

type JobSnapshot struct {
	ConversationID string    `json:"conversation_id"`
	Status         JobStatus `json:"status"`
	Error          string    `json:"error,omitempty"`
	Running        bool      `json:"running"`
}

type JobManager struct {
	mu   sync.RWMutex
	jobs map[string]*AssistantJob
}

func NewJobManager() *JobManager {
	return &JobManager{jobs: make(map[string]*AssistantJob)}
}

func (m *JobManager) Start(convID string, run func(context.Context, StreamSink) error) (*AssistantJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cleanupLocked()

	if existing, ok := m.jobs[convID]; ok && existing.IsRunning() {
		return existing, ErrJobAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := newAssistantJob(convID, cancel)
	m.jobs[convID] = job

	go func() {
		err := run(ctx, job.sink())
		job.finish(ctx, err)
	}()

	return job, nil
}

func (m *JobManager) Get(convID string) (*AssistantJob, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[convID]
	return job, ok
}

func (m *JobManager) Snapshot(convID string) (JobSnapshot, bool) {
	job, ok := m.Get(convID)
	if !ok {
		return JobSnapshot{}, false
	}
	return job.Snapshot(), true
}

func (m *JobManager) Running() []JobSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make([]JobSnapshot, 0)
	for _, job := range m.jobs {
		snapshot := job.Snapshot()
		if snapshot.Running {
			snapshots = append(snapshots, snapshot)
		}
	}
	return snapshots
}

func (m *JobManager) Stop(convID string) error {
	job, ok := m.Get(convID)
	if !ok || !job.IsRunning() {
		return fmt.Errorf("no active chat found for conversation: %s", convID)
	}
	job.cancel()
	return nil
}

func (m *JobManager) cleanupLocked() {
	cutoff := time.Now().Add(-30 * time.Minute)
	for id, job := range m.jobs {
		if job.isExpired(cutoff) {
			delete(m.jobs, id)
		}
	}
}

type AssistantJob struct {
	convID string
	cancel context.CancelFunc

	mu         sync.Mutex
	status     JobStatus
	err        string
	doneAt     time.Time
	events     [][]byte
	subs       map[int]chan []byte
	nextSubID  int
	totalBytes int
}

func newAssistantJob(convID string, cancel context.CancelFunc) *AssistantJob {
	return &AssistantJob{
		convID: convID,
		cancel: cancel,
		status: JobStatusRunning,
		subs:   make(map[int]chan []byte),
	}
}

func (j *AssistantJob) sink() StreamSink {
	return &JobStreamSink{job: j}
}

func (j *AssistantJob) Subscribe() (<-chan []byte, func()) {
	j.mu.Lock()
	defer j.mu.Unlock()

	capacity := len(j.events) + 256
	if capacity < 256 {
		capacity = 256
	}
	ch := make(chan []byte, capacity)

	for _, event := range j.events {
		ch <- cloneBytes(event)
	}

	if j.status != JobStatusRunning {
		close(ch)
		return ch, func() {}
	}

	subID := j.nextSubID
	j.nextSubID++
	j.subs[subID] = ch

	return ch, func() {
		j.mu.Lock()
		defer j.mu.Unlock()
		if sub, ok := j.subs[subID]; ok {
			delete(j.subs, subID)
			close(sub)
		}
	}
}

func (j *AssistantJob) Snapshot() JobSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()
	return JobSnapshot{
		ConversationID: j.convID,
		Status:         j.status,
		Error:          j.err,
		Running:        j.status == JobStatusRunning,
	}
}

func (j *AssistantJob) IsRunning() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.status == JobStatusRunning
}

func (j *AssistantJob) isExpired(cutoff time.Time) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.status != JobStatusRunning && !j.doneAt.IsZero() && j.doneAt.Before(cutoff)
}

func (j *AssistantJob) finish(ctx context.Context, err error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			j.status = JobStatusCanceled
		} else {
			j.status = JobStatusError
			j.err = err.Error()
		}
	} else {
		j.status = JobStatusDone
	}
	j.doneAt = time.Now()

	for id, ch := range j.subs {
		close(ch)
		delete(j.subs, id)
	}
}

func (j *AssistantJob) appendEvent(p []byte) {
	event := cloneBytes(p)

	j.mu.Lock()
	defer j.mu.Unlock()

	j.events = append(j.events, event)
	j.totalBytes += len(event)
	for j.totalBytes > 2*1024*1024 && len(j.events) > 1 {
		j.totalBytes -= len(j.events[0])
		j.events = j.events[1:]
	}

	for _, ch := range j.subs {
		select {
		case ch <- cloneBytes(event):
		default:
		}
	}
}

type JobStreamSink struct {
	job *AssistantJob
}

func (s *JobStreamSink) Prepare(convID, model, provider string) error {
	return s.Send(ChatResponse{
		ConversationID: convID,
		Model:          model,
		ProviderName:   provider,
		Message:        MessageContent{Role: "assistant"},
		Done:           false,
	})
}

func (s *JobStreamSink) Send(data ChatResponse) error {
	if s == nil || s.job == nil {
		return nil
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s.job.appendEvent([]byte(fmt.Sprintf("data: %s\n\n", payload)))
	return nil
}

func (s *JobStreamSink) Done(convID, model string) error {
	if err := s.Send(ChatResponse{ConversationID: convID, Model: model, Done: true}); err != nil {
		return err
	}
	s.job.appendEvent([]byte("data: [DONE]\n\n"))
	return nil
}

func cloneBytes(p []byte) []byte {
	cp := make([]byte, len(p))
	copy(cp, p)
	return cp
}
