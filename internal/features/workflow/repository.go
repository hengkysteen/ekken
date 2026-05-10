package workflow

import (
	"database/sql"
	"fmt"
	"time"

	"ekken/internal/db"
)

func init() {
	db.RegisterMigration(`CREATE TABLE IF NOT EXISTS workflows (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			data TEXT NOT NULL,
			status TEXT DEFAULT 'idle',
			iteration INTEGER DEFAULT 0,
			last_run_at DATETIME,
			created_by TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	db.RegisterMigration(`CREATE INDEX IF NOT EXISTS idx_workflows_name ON workflows(name)`)
	db.RegisterMigration(`CREATE INDEX IF NOT EXISTS idx_workflows_updated ON workflows(updated_at)`)
}

// Repository handles database interactions for the workflow feature.
type Repository struct {
	db *db.DB
}

func NewRepository(database *db.DB) *Repository {
	return &Repository{db: database}
}

// SaveWorkflow stores the raw JSON data of a workflow into the database.
func (r *Repository) SaveWorkflow(id, name, data, createdBy string) error {
	_, err := r.db.Conn().Exec(
		`INSERT INTO workflows (id, name, data, created_by, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(id) DO UPDATE SET data = excluded.data, name = excluded.name, updated_at = CURRENT_TIMESTAMP`,
		id, name, data, createdBy,
	)
	if err != nil {
		return fmt.Errorf("save workflow: %w", err)
	}
	return nil
}

// GetWorkflow retrieves the raw JSON data of a workflow.
func (r *Repository) GetWorkflow(id string) (string, error) {
	var data string
	err := r.db.Conn().QueryRow(`SELECT data FROM workflows WHERE id = ?`, id).Scan(&data)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("workflow not found: %s", id)
	}
	if err != nil {
		return "", fmt.Errorf("get workflow: %w", err)
	}
	return data, nil
}

type WorkflowSummary struct {
	ID        string
	Name      string
	Data      string
	Status    string
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastRunAt *time.Time
}

func (r *Repository) ListWorkflows() ([]WorkflowSummary, error) {
	rows, err := r.db.Conn().Query(`SELECT id, name, data, status, created_by, created_at, updated_at, last_run_at FROM workflows ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	defer rows.Close()

	var results []WorkflowSummary
	for rows.Next() {
		var s WorkflowSummary
		var lastRunAt sql.NullTime
		var createdBy sql.NullString
		if err := rows.Scan(&s.ID, &s.Name, &s.Data, &s.Status, &createdBy, &s.CreatedAt, &s.UpdatedAt, &lastRunAt); err != nil {
			return nil, fmt.Errorf("scan workflow: %w", err)
		}
		if lastRunAt.Valid {
			s.LastRunAt = &lastRunAt.Time
		}
		if createdBy.Valid {
			s.CreatedBy = createdBy.String
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *Repository) DeleteWorkflow(id string) error {
	result, err := r.db.Conn().Exec(`DELETE FROM workflows WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete workflow: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("workflow not found: %s", id)
	}
	return nil
}

func (r *Repository) DeleteAllWorkflows() error {
	_, err := r.db.Conn().Exec(`DELETE FROM workflows`)
	return err
}

func (r *Repository) UpdateStatus(id, status string, iteration int) error {
	_, err := r.db.Conn().Exec(
		`UPDATE workflows SET status = ?, iteration = ?, last_run_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, iteration, id,
	)
	return err
}

func (r *Repository) UpdateLastRun(id string, lastRunAt time.Time) error {
	_, err := r.db.Conn().Exec(
		`UPDATE workflows SET last_run_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		lastRunAt, id,
	)
	return err
}

type WorkflowStatusInfo struct {
	Status    string
	Iteration int
	LastRunAt *time.Time
}

func (r *Repository) GetStatus(id string) (WorkflowStatusInfo, error) {
	var status string
	var iteration int
	var lastRunAt sql.NullTime

	err := r.db.Conn().QueryRow(
		`SELECT status, iteration, last_run_at FROM workflows WHERE id = ?`, id,
	).Scan(&status, &iteration, &lastRunAt)
	if err == sql.ErrNoRows {
		return WorkflowStatusInfo{}, fmt.Errorf("workflow not found: %s", id)
	}
	if err != nil {
		return WorkflowStatusInfo{}, fmt.Errorf("get status: %w", err)
	}

	info := WorkflowStatusInfo{
		Status:    status,
		Iteration: iteration,
	}
	if lastRunAt.Valid {
		info.LastRunAt = &lastRunAt.Time
	}
	return info, nil
}
