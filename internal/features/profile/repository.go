package profile

import (
	"database/sql"
	"time"

	"ekken/internal/db"
)

const profileSchema = `CREATE TABLE IF NOT EXISTS profile (
	id                   INTEGER PRIMARY KEY CHECK (id = 1),
	name                 TEXT NOT NULL DEFAULT '',
	pin_enabled          BOOLEAN NOT NULL DEFAULT false,
	pin_hash             TEXT NOT NULL DEFAULT '',
	updated_at           TEXT NOT NULL,
	pin_updated_at       TEXT NOT NULL DEFAULT '',
	security_question    TEXT NOT NULL DEFAULT '',
	security_answer_hash TEXT NOT NULL DEFAULT ''
)`

func init() {
	db.RegisterMigration(profileSchema)
}

type Repository struct {
	db *db.DB
}

func NewRepository(database *db.DB) *Repository {
	r := &Repository{db: database}
	// Migrate existing tables if they exist
	_ = database.AddColumnIfNotExists("profile", "security_question", "TEXT NOT NULL DEFAULT ''")
	_ = database.AddColumnIfNotExists("profile", "security_answer_hash", "TEXT NOT NULL DEFAULT ''")
	return r
}

func defaultProfileItem() ProfileItem {
	return ProfileItem{
		Name:       "",
		PINEnabled: false,
		UpdatedAt:  "",
	}
}

func (r *Repository) GetProfile() (ProfileItem, error) {
	item := defaultProfileItem()
	err := r.db.Conn().QueryRow(
		`SELECT name, pin_enabled, pin_hash, updated_at, pin_updated_at, security_question, security_answer_hash 
		 FROM profile WHERE id = 1`,
	).Scan(&item.Name, &item.PINEnabled, &item.PINHash, &item.UpdatedAt, &item.PINUpdatedAt, &item.SecurityQuestion, &item.SecurityAnswerHash)
	
	if err == sql.ErrNoRows {
		return item, nil
	}
	if err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) SaveProfile(item ProfileItem) error {
	if item.UpdatedAt == "" {
		item.UpdatedAt = time.Now().Format(time.RFC3339)
	}

	_, err := r.db.Conn().Exec(
		`INSERT INTO profile (id, name, pin_enabled, pin_hash, updated_at, pin_updated_at, security_question, security_answer_hash)
		 VALUES (1, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   name = excluded.name,
		   pin_enabled = excluded.pin_enabled,
		   pin_hash = excluded.pin_hash,
		   updated_at = excluded.updated_at,
		   pin_updated_at = excluded.pin_updated_at,
		   security_question = excluded.security_question,
		   security_answer_hash = excluded.security_answer_hash`,
		item.Name,
		item.PINEnabled,
		item.PINHash,
		item.UpdatedAt,
		item.PINUpdatedAt,
		item.SecurityQuestion,
		item.SecurityAnswerHash,
	)
	return err
}
