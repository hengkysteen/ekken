package mynode

import (
	"database/sql"
	"fmt"
	"time"

	"ekken/internal/db"
)

func init() {
	db.RegisterMigration(`CREATE TABLE IF NOT EXISTS my_nodes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
}

type MyNodesRepository struct {
	db *db.DB
}

func NewMyNodesRepository(database *db.DB) *MyNodesRepository {
	return &MyNodesRepository{db: database}
}

func (r *MyNodesRepository) SaveMyNodesItem(item MyNodeData) error {
	_, err := r.db.Conn().Exec(
		`INSERT INTO my_nodes (id, name, data, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET name = excluded.name, data = excluded.data`,
		item.ID, item.Name, item.Data, time.Now().Format(time.RFC3339),
	)
	return err
}

func (r *MyNodesRepository) GetMyNodesItem(id string) (MyNodeData, error) {
	var item MyNodeData
	err := r.db.Conn().QueryRow(
		`SELECT id, name, data, created_at FROM my_nodes WHERE id = ?`, id,
	).Scan(&item.ID, &item.Name, &item.Data, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return MyNodeData{}, fmt.Errorf("my nodes item not found: %s", id)
	}
	return item, err
}

func (r *MyNodesRepository) ListMyNodesItems() ([]MyNodeData, error) {
	rows, err := r.db.Conn().Query(`SELECT id, name, data, created_at FROM my_nodes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MyNodeData
	for rows.Next() {
		var item MyNodeData
		if err := rows.Scan(&item.ID, &item.Name, &item.Data, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MyNodesRepository) DeleteMyNodesItem(id string) error {
	_, err := r.db.Conn().Exec(`DELETE FROM my_nodes WHERE id = ?`, id)
	return err
}

func (r *MyNodesRepository) UpdateMyNodesItem(id string, name string, data string) error {
	_, err := r.db.Conn().Exec(
		`UPDATE my_nodes SET name = ?, data = ? WHERE id = ?`,
		name, data, id,
	)
	return err
}
