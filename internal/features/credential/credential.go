package credential

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Servicer defines the business logic interface for credential management.
type Servicer interface {
	List() ([]Credential, error)
	Get(id string) (Credential, error)
	Create(req Credential) (Credential, error)
	Update(id string, req Credential) (Credential, error)
	Delete(id string) error
	GetValueByKey(key string) (string, error)
	ResolveConfig(config map[string]string) (map[string]string, error)
}

// Database is the storage interface required by Service.
type Database interface {
	ListCredentials() ([]CredentialItem, error)
	GetCredential(id string) (CredentialItem, error)
	GetCredentialByKey(key string) (string, error)
	SaveCredential(item CredentialItem) error
	UpdateCredential(id, name, key, value string, tags []string) error
	DeleteCredential(id string) error
}

// Service implements Servicer.
type Service struct {
	db Database
}

// New creates a new Service backed by the given database.
func New(database Database) *Service {
	return &Service{db: database}
}

// List returns all credentials without the sensitive value field.
func (s *Service) List() ([]Credential, error) {
	items, err := s.db.ListCredentials()
	if err != nil {
		return nil, err
	}
	result := make([]Credential, 0, len(items))
	for _, item := range items {
		result = append(result, Credential{
			ID:        item.ID,
			Name:      item.Name,
			Key:       item.Key,
			Tags:      item.Tags,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return result, nil
}

// Get retrieves a single credential by ID, including the decrypted value.
func (s *Service) Get(id string) (Credential, error) {
	item, err := s.db.GetCredential(id)
	if err != nil {
		return Credential{}, err
	}
	return Credential{
		ID:        item.ID,
		Name:      item.Name,
		Key:       item.Key,
		Value:     item.Value,
		Tags:      item.Tags,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

// Create validates and stores a new credential with the value encrypted.
func (s *Service) Create(req Credential) (Credential, error) {
	if strings.TrimSpace(req.Name) == "" {
		return Credential{}, fmt.Errorf("credential name is required")
	}
	if strings.TrimSpace(req.Key) == "" {
		return Credential{}, fmt.Errorf("credential key is required")
	}
	if !strings.HasPrefix(strings.ToLower(req.Key), "cred.") {
		req.Key = "cred." + strings.TrimSpace(req.Key)
	}
	if strings.TrimSpace(req.Value) == "" {
		return Credential{}, fmt.Errorf("credential value is required")
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	id := fmt.Sprintf("cred_%d", time.Now().UnixNano())
	item := CredentialItem{
		ID:    id,
		Name:  req.Name,
		Key:   req.Key,
		Value: req.Value,
		Tags:  req.Tags,
	}
	if err := s.db.SaveCredential(item); err != nil {
		return Credential{}, err
	}
	req.ID = id
	return req, nil
}


// Update replaces an existing credential's fields (value is re-encrypted).
func (s *Service) Update(id string, req Credential) (Credential, error) {
	if strings.TrimSpace(req.Name) == "" {
		return Credential{}, fmt.Errorf("credential name is required")
	}
	if strings.TrimSpace(req.Key) == "" {
		return Credential{}, fmt.Errorf("credential key is required")
	}
	if !strings.HasPrefix(strings.ToLower(req.Key), "cred.") {
		req.Key = "cred." + strings.TrimSpace(req.Key)
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if err := s.db.UpdateCredential(id, req.Name, req.Key, req.Value, req.Tags); err != nil {
		return Credential{}, err
	}
	req.ID = id
	return req, nil
}


// Delete removes a credential by ID.
func (s *Service) Delete(id string) error {
	return s.db.DeleteCredential(id)
}

// GetValueByKey retrieves the raw value of a credential by its unique key name.
func (s *Service) GetValueByKey(key string) (string, error) {
	return s.db.GetCredentialByKey(key)
}

// ResolveConfig scans for {{ cred.KEY }} patterns and replaces them with real credential values from database.
// It will throw an error if the format is strictly not {{ cred.reference }} or if the key is not found.
func (s *Service) ResolveConfig(config map[string]string) (map[string]string, error) {
	resolved := make(map[string]string)

	// Regex that looks for {{ cred.reference }} anywhere in the string
	re := regexp.MustCompile(`\{\{\s*(cred\..+?)\s*\}\}`)

	for k, v := range config {
		matches := re.FindStringSubmatch(v)
		if len(matches) < 2 {
			// Not a placeholder, use raw value
			resolved[k] = v
			continue
		}

		keyName := strings.TrimSpace(matches[1])
		val, err := s.GetValueByKey(keyName)
		if err != nil {
			// If key not found, we return error to be safe.
			return nil, fmt.Errorf("credential reference '%s' not found for key '%s'", keyName, k)
		}
		resolved[k] = val
	}

	return resolved, nil
}

