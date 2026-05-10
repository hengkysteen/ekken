package mynode

import (
	"testing"
)

type MockMyNodesDB struct {
	ListFunc   func() ([]MyNodeData, error)
	SaveFunc   func(item MyNodeData) error
	DeleteFunc func(id string) error
	UpdateFunc func(id string, name string, data string) error
}

func (m *MockMyNodesDB) ListMyNodesItems() ([]MyNodeData, error) { return m.ListFunc() }
func (m *MockMyNodesDB) SaveMyNodesItem(item MyNodeData) error   { return m.SaveFunc(item) }
func (m *MockMyNodesDB) DeleteMyNodesItem(id string) error       { return m.DeleteFunc(id) }
func (m *MockMyNodesDB) UpdateMyNodesItem(id string, name string, data string) error {
	return m.UpdateFunc(id, name, data)
}

func TestMyNodesService_Delete(t *testing.T) {
	called := false
	mock := &MockMyNodesDB{
		DeleteFunc: func(id string) error {
			called = true
			if id != "lib_123" {
				t.Errorf("expected lib_123, got %s", id)
			}
			return nil
		},
	}

	service := NewMyNodesService(mock)
	err := service.Delete("lib_123")

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if !called {
		t.Errorf("expected DeleteMyNodesItem to be called")
	}
}

func TestMyNodesService_List(t *testing.T) {
	mock := &MockMyNodesDB{
		ListFunc: func() ([]MyNodeData, error) {
			return []MyNodeData{
				{ID: "1", Name: "Node 1", Data: `{"name":"Node 1"}`},
			}, nil
		},
	}

	service := NewMyNodesService(mock)
	items, err := service.List()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "Node 1" {
		t.Errorf("expected Node 1, got %s", items[0].Name)
	}
}
