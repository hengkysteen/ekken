package node

import (
	"fmt"
	"sync"
)

type EventListener struct {
	ID         string
	Action     Action
	WorkflowID string
	Payload    chan any
}

type EventListenerManager struct {
	mu        sync.RWMutex
	listeners map[string]*EventListener
}

func NewEventListenerManager() *EventListenerManager {
	return &EventListenerManager{
		listeners: make(map[string]*EventListener),
	}
}

func (m *EventListenerManager) Register(id string, action Action, workflowID string) (*EventListener, func(), error) {
	if id == "" {
		return nil, nil, fmt.Errorf("listener id is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.listeners[id]; exists {
		return nil, nil, fmt.Errorf("event listener '%s' is already active", id)
	}

	listener := &EventListener{
		ID:         id,
		Action:     action,
		WorkflowID: workflowID,
		Payload:    make(chan any, 1),
	}
	m.listeners[id] = listener

	cleanup := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if current, ok := m.listeners[id]; ok && current == listener {
			delete(m.listeners, id)
		}
	}

	return listener, cleanup, nil
}

func (m *EventListenerManager) Get(id string) (*EventListener, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	listener, ok := m.listeners[id]
	return listener, ok
}

func (m *EventListenerManager) Dispatch(id string, payload any) error {
	listener, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("event listener '%s' is not active", id)
	}

	select {
	case listener.Payload <- payload:
		return nil
	default:
		return fmt.Errorf("event listener '%s' is busy", id)
	}
}

var GlobalEventListeners = NewEventListenerManager()
