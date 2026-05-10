package credential

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"ekken/internal/db"
)

func init() {
	db.RegisterMigration(`CREATE TABLE IF NOT EXISTS credentials (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL UNIQUE,
			key        TEXT NOT NULL,
			value      TEXT NOT NULL,
			tags       TEXT NOT NULL DEFAULT '[]',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	db.RegisterMigration(`CREATE INDEX IF NOT EXISTS idx_credentials_name ON credentials(name)`)
}

// Repository handles database interactions and encryption for credentials.
type Repository struct {
	db     *db.DB
	encKey []byte
}

func NewRepository(database *db.DB) (*Repository, error) {
	r := &Repository{db: database}
	if err := r.initEncryptionKey(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repository) initEncryptionKey() error {
	var hexKey string
	err := r.db.Conn().QueryRow(`SELECT value FROM app_secrets WHERE key = 'encryption_key'`).Scan(&hexKey)
	if err == sql.ErrNoRows {
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return fmt.Errorf("generate encryption key: %w", err)
		}
		hexKey = hex.EncodeToString(raw)
		if _, err := r.db.Conn().Exec(`INSERT INTO app_secrets (key, value) VALUES ('encryption_key', ?)`, hexKey); err != nil {
			return fmt.Errorf("store encryption key: %w", err)
		}
		r.encKey = raw
		return nil
	}
	if err != nil {
		return fmt.Errorf("load encryption key: %w", err)
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return fmt.Errorf("decode encryption key: %w", err)
	}
	r.encKey = key
	return nil
}

func (r *Repository) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(r.encKey)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (r *Repository) decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(r.encKey)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}

// CredentialItem is the DB-level representation.
type CredentialItem struct {
	ID        string
	Name      string
	Key       string
	Value     string
	Tags      []string
	CreatedAt string
	UpdatedAt string
}

func (r *Repository) SaveCredential(item CredentialItem) error {
	encrypted, err := r.encrypt(item.Value)
	if err != nil {
		return err
	}
	tagsJSON, _ := json.Marshal(item.Tags)
	now := time.Now().Format(time.RFC3339)
	_, err = r.db.Conn().Exec(
		`INSERT INTO credentials (id, name, key, value, tags, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Name, item.Key, encrypted, string(tagsJSON), now, now,
	)
	return err
}

func (r *Repository) GetCredential(id string) (CredentialItem, error) {
	var item CredentialItem
	var encValue, tagsJSON string
	err := r.db.Conn().QueryRow(
		`SELECT id, name, key, value, tags, created_at, updated_at FROM credentials WHERE id = ?`, id,
	).Scan(&item.ID, &item.Name, &item.Key, &encValue, &tagsJSON, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return item, err
	}
	decrypted, err := r.decrypt(encValue)
	if err != nil {
		return item, err
	}
	item.Value = decrypted
	json.Unmarshal([]byte(tagsJSON), &item.Tags)
	return item, nil
}

func (r *Repository) ListCredentials() ([]CredentialItem, error) {
	rows, err := r.db.Conn().Query(`SELECT id, name, key, tags, created_at, updated_at FROM credentials ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CredentialItem
	for rows.Next() {
		var item CredentialItem
		var tagsJSON string
		if err := rows.Scan(&item.ID, &item.Name, &item.Key, &tagsJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &item.Tags)
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) UpdateCredential(id, name, key, value string, tags []string) error {
	encrypted, err := r.encrypt(value)
	if err != nil {
		return err
	}
	tagsJSON, _ := json.Marshal(tags)
	_, err = r.db.Conn().Exec(
		`UPDATE credentials SET name = ?, key = ?, value = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, key, encrypted, string(tagsJSON), id,
	)
	return err
}

func (r *Repository) DeleteCredential(id string) error {
	_, err := r.db.Conn().Exec(`DELETE FROM credentials WHERE id = ?`, id)
	return err
}

func (r *Repository) GetCredentialByKey(key string) (string, error) {
	var encValue string
	err := r.db.Conn().QueryRow(`SELECT value FROM credentials WHERE key = ?`, key).Scan(&encValue)
	if err != nil {
		return "", err
	}
	return r.decrypt(encValue)
}
