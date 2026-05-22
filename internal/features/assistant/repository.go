package assistant

import (
	"encoding/json"

	"ekken/internal/db"
)

func init() {
	db.RegisterMigration(`CREATE TABLE IF NOT EXISTS assistant (
			provider_id TEXT PRIMARY KEY,
			config JSON,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
}

type Repository struct {
	db *db.DB
}

func NewRepository(database *db.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) SaveAssistantProvider(providerID string, config map[string]string) error {
	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = r.db.Conn().Exec(
		`INSERT INTO assistant (provider_id, config, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(provider_id) DO UPDATE SET config = excluded.config, updated_at = CURRENT_TIMESTAMP`,
		providerID, string(configData),
	)
	return err
}

type AssistantProviderRecord struct {
	ProviderID string
	Config     map[string]string
}

func (r *Repository) GetAssistantProviders() ([]AssistantProviderRecord, error) {
	rows, err := r.db.Conn().Query("SELECT provider_id, config FROM assistant")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []AssistantProviderRecord
	for rows.Next() {
		var providerID, configData string
		if err := rows.Scan(&providerID, &configData); err != nil {
			return nil, err
		}
		var config map[string]string
		json.Unmarshal([]byte(configData), &config)
		providers = append(providers, AssistantProviderRecord{
			ProviderID: providerID,
			Config:     config,
		})
	}
	return providers, rows.Err()
}

func (r *Repository) DeleteAssistantProvider(providerID string) error {
	_, err := r.db.Conn().Exec("DELETE FROM assistant WHERE provider_id = ?", providerID)
	return err
}
