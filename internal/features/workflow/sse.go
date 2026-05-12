package workflow

import (
	"fmt"
	"sync"
	"sync/atomic"

	"ekken/internal/logger"
)

type SSEServicer interface {
	Subscribe(id string) (string, <-chan SSEMessage)
	Unsubscribe(id, subID string)
	Send(id string, msg SSEMessage)
	Create(id string)
	Finish(id string)
	SubscribeGlobal() (string, <-chan SSEMessage)
	UnsubscribeGlobal(subID string)
	SendGlobal(msg SSEMessage)
}

type WorkflowEventStream struct {
	mu        sync.RWMutex
	workflows map[string]*workflowStream
	global    *GlobalEventStream
}

type workflowStream struct {
	mu   sync.RWMutex
	subs map[string]chan SSEMessage
}

func NewWorkflowEventStream() *WorkflowEventStream {
	return &WorkflowEventStream{
		workflows: make(map[string]*workflowStream),
		global:    NewGlobalEventStream(),
	}
}

func (s *WorkflowEventStream) Create(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.workflows[id]; !exists {
		s.workflows[id] = &workflowStream{subs: make(map[string]chan SSEMessage)}
	}
}

func (s *WorkflowEventStream) Subscribe(id string) (string, <-chan SSEMessage) {
	s.mu.RLock()
	ws, ok := s.workflows[id]
	s.mu.RUnlock()

	if !ok {
		s.mu.Lock()
		ws, ok = s.workflows[id]
		if !ok {
			ws = &workflowStream{subs: make(map[string]chan SSEMessage)}
			s.workflows[id] = ws
		}
		s.mu.Unlock()
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	subID := fmt.Sprintf("sub-%s-%d", id, globalSubCounter.Add(1))
	ch := make(chan SSEMessage, 2048)
	ws.subs[subID] = ch

	return subID, ch
}

func (s *WorkflowEventStream) Unsubscribe(id, subID string) {
	s.mu.RLock()
	ws, ok := s.workflows[id]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ws.mu.Lock()
	delete(ws.subs, subID)
	ws.mu.Unlock()
}

func (s *WorkflowEventStream) Send(id string, msg SSEMessage) {
	s.mu.RLock()
	ws, ok := s.workflows[id]
	s.mu.RUnlock()
	if !ok {
		return
	}

	ws.mu.RLock()
	defer ws.mu.RUnlock()

	for subID, ch := range ws.subs {
		select {
		case ch <- msg:
		default:
			logger.Error("Workflow SSE buffer full, message dropped", "id", id, "sub", subID, "type", msg.Type)
		}
		_ = subID
	}
}

func (s *WorkflowEventStream) Finish(id string) {
	s.mu.Lock()
	ws, ok := s.workflows[id]
	if ok {
		delete(s.workflows, id)
	}
	s.mu.Unlock()
	if !ok {
		return
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()
	for _, ch := range ws.subs {
		close(ch)
	}
}

func (s *WorkflowEventStream) SubscribeGlobal() (string, <-chan SSEMessage) {
	return s.global.Subscribe()
}

func (s *WorkflowEventStream) UnsubscribeGlobal(subID string) {
	s.global.Unsubscribe(subID)
}

func (s *WorkflowEventStream) SendGlobal(msg SSEMessage) {
	s.global.Send(msg)
}

var globalSubCounter atomic.Int64

type GlobalEventStream struct {
	mu   sync.RWMutex
	subs map[string]chan SSEMessage
}

func NewGlobalEventStream() *GlobalEventStream {
	return &GlobalEventStream{
		subs: make(map[string]chan SSEMessage),
	}
}

func (s *GlobalEventStream) Subscribe() (string, <-chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subID := fmt.Sprintf("global-%d", globalSubCounter.Add(1))
	ch := make(chan SSEMessage, 4096)
	s.subs[subID] = ch

	return subID, ch
}

func (s *GlobalEventStream) Unsubscribe(subID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subs, subID)
}

func (s *GlobalEventStream) Send(msg SSEMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subs {
		select {
		case ch <- msg:
		default:
			logger.Error("Global SSE buffer full, message dropped", "type", msg.Type)
		}
	}
}
