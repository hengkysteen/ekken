package mynode

import (
	"encoding/json"
	"fmt"
	"time"
)

func NewMyNodesService(database MyNodesDatabase) *MyNodesService {
	return &MyNodesService{db: database}
}

func (s *MyNodesService) List() ([]MyNodesItem, error) {
	items, err := s.db.ListMyNodesItems()
	if err != nil {
		return nil, err
	}

	result := make([]MyNodesItem, 0, len(items))
	for _, item := range items {
		var parsed MyNodesItem
		if err := json.Unmarshal([]byte(item.Data), &parsed); err != nil {
			continue
		}
		// Set the ID from the database record
		parsed.ID = item.ID
		parsed.CreatedAt = item.CreatedAt

		result = append(result, parsed)
	}
	return result, nil
}

func (s *MyNodesService) Save(req MyNodesItem) (MyNodesItem, error) {
	id := fmt.Sprintf("lib_%d", time.Now().UnixNano())
	req.ID = id // Ensure ID is in the data

	dataJSON, err := json.Marshal(req)
	if err != nil {
		return MyNodesItem{}, fmt.Errorf("marshal data: %w", err)
	}

	item := MyNodeData{
		ID:   id,
		Name: req.Name,
		Data: string(dataJSON),
	}

	if err := s.db.SaveMyNodesItem(item); err != nil {
		return MyNodesItem{}, err
	}

	return req, nil
}

func (s *MyNodesService) Delete(id string) error {
	return s.db.DeleteMyNodesItem(id)
}

func (s *MyNodesService) Update(id string, req MyNodesItem) (MyNodesItem, error) {
	req.ID = id
	dataJSON, err := json.Marshal(req)
	if err != nil {
		return MyNodesItem{}, fmt.Errorf("marshal data: %w", err)
	}

	if err := s.db.UpdateMyNodesItem(id, req.Name, string(dataJSON)); err != nil {
		return MyNodesItem{}, err
	}

	return req, nil
}
