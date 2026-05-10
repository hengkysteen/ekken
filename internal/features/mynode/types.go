package mynode

import "ekken/internal/features/workflow/node"

type MyNodeData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Data      string `json:"data"`
	CreatedAt string `json:"created_at"`
}

type MyNodesService struct {
	db MyNodesDatabase
}

type MyNodesItem struct {
	node.Node
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type MyNodesServicer interface {
	List() ([]MyNodesItem, error)
	Save(req MyNodesItem) (MyNodesItem, error)
	Delete(id string) error
	Update(id string, req MyNodesItem) (MyNodesItem, error)
}

type MyNodesDatabase interface {
	ListMyNodesItems() ([]MyNodeData, error)
	SaveMyNodesItem(item MyNodeData) error
	DeleteMyNodesItem(id string) error
	UpdateMyNodesItem(id string, name string, data string) error
}
