package credential

import (
	"fmt"
	"testing"

)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	data map[string]string
}

func (m *MockDatabase) GetCredentialByKey(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("not found")
}

// Implement other methods as no-ops since they aren't used in this specific test
func (m *MockDatabase) ListCredentials() ([]CredentialItem, error) { return nil, nil }
func (m *MockDatabase) GetCredential(id string) (CredentialItem, error) {
	return CredentialItem{}, nil
}
func (m *MockDatabase) SaveCredential(item CredentialItem) error                       { return nil }
func (m *MockDatabase) UpdateCredential(id, name, key, value string, tags []string) error { return nil }
func (m *MockDatabase) DeleteCredential(id string) error                                  { return nil }

func TestResolveConfig(t *testing.T) {
	mockDB := &MockDatabase{
		data: map[string]string{
			"cred.MY_API_KEY": "sk-secret-token",
		},
	}

	service := New(mockDB)

	tests := []struct {
		name    string
		config  map[string]string
		wantVal string
		wantErr bool
	}{
		{
			name: "Valid reference format with prefix",
			config: map[string]string{
				"api_key": "{{ cred.MY_API_KEY }}",
			},
			wantVal: "sk-secret-token",
			wantErr: false,
		},
		{
			name: "Plain text value",
			config: map[string]string{
				"api_key": "plain-text-key",
			},
			wantVal: "plain-text-key",
			wantErr: false,
		},
		{
			name: "No prefix treated as plain text",
			config: map[string]string{
				"api_key": "{{ MY_API_KEY }}",
			},
			wantVal: "{{ MY_API_KEY }}",
			wantErr: false,
		},
		{
			name: "Valid prefix but reference not found",
			config: map[string]string{
				"api_key": "{{ cred.NOT_FOUND_KEY }}",
			},
			wantErr: true,
		},

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ResolveConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got["api_key"] != tt.wantVal {
				t.Errorf("ResolveConfig() got = %v, want %v", got["api_key"], tt.wantVal)
			}
		})
	}
}
