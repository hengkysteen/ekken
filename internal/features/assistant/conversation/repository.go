package conversation

import (
	"time"

	"ekken/internal/db"
)

func init() {
	db.RegisterMigration(`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	db.RegisterMigration(`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			provider TEXT,
			model TEXT,
			agent TEXT,
			thinking TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		)`)
	db.RegisterMigration(`CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)`)
}

type Repository struct {
	db *db.DB
}

func NewRepository(database *db.DB) *Repository {
	database.AddColumnIfNotExists("messages", "agent", "TEXT")
	database.AddColumnIfNotExists("messages", "is_system", "BOOLEAN DEFAULT FALSE")
	return &Repository{db: database}
}

type ConversationItem struct {
	ID        string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MessageItem struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	Thinking       string
	Provider       string
	Model          string
	Agent          string
	IsSystem       bool
	CreatedAt      time.Time
}

func (r *Repository) CreateConversation(id, title string) error {
	_, err := r.db.Conn().Exec("INSERT INTO conversations (id, title) VALUES (?, ?)", id, title)
	return err
}

func (r *Repository) ListConversations() ([]ConversationItem, error) {
	rows, err := r.db.Conn().Query("SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ConversationItem
	for rows.Next() {
		var item ConversationItem
		if err := rows.Scan(&item.ID, &item.Title, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetConversation(id string) (ConversationItem, error) {
	var item ConversationItem
	err := r.db.Conn().QueryRow("SELECT id, title, created_at, updated_at FROM conversations WHERE id = ?", id).
		Scan(&item.ID, &item.Title, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) UpdateConversationTitle(id, title string) error {
	_, err := r.db.Conn().Exec("UPDATE conversations SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", title, id)
	return err
}

func (r *Repository) DeleteConversation(id string) error {
	tx, err := r.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM messages WHERE conversation_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM conversations WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteAllConversations() error {
	tx, err := r.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM messages"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM conversations"); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) AddMessage(id, conversationID, role, content, thinking, provider, model, agent string, isSystem bool) error {
	tx, err := r.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO messages (id, conversation_id, role, content, thinking, provider, model, agent, is_system) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", id, conversationID, role, content, thinking, provider, model, agent, isSystem)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", conversationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetMessages(conversationID string) ([]MessageItem, error) {
	rows, err := r.db.Conn().Query("SELECT id, conversation_id, role, content, thinking, provider, model, agent, is_system, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at ASC", conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MessageItem
	for rows.Next() {
		var item MessageItem
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.Role, &item.Content, &item.Thinking, &item.Provider, &item.Model, &item.Agent, &item.IsSystem, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

