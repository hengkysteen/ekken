package node

import (
	"fmt"
	"os"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	variables := map[string]interface{}{
		"API_KEY": "secret_123",
		"count":   10,
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "no space",
			template: "Bearer {{API_KEY}}",
			want:     "Bearer secret_123",
		},
		{
			name:     "with spaces",
			template: "Bearer {{ API_KEY }}",
			want:     "Bearer secret_123",
		},
		{
			name:     "multiple spaces",
			template: "Bearer {{   API_KEY   }}",
			want:     "Bearer secret_123",
		},
		{
			name:     "multiple variables",
			template: "{{ API_KEY }} - {{ count }}",
			want:     "secret_123 - 10",
		},
		{
			name:     "arithmetic with spaces",
			template: "Value: {{ count + 5 }}",
			want:     "Value: 15",
		},
		{
			name:     "credential resolution",
			template: "Key: {{ cred.MISTRAL_API_KEY }}",
			want:     "Key: mistral_secret_key",
		},
		{
			name:     "credential resolution no prefix (fail)",
			template: "Key: {{ MISTRAL_API_KEY }}",
			want:     "Key: {{ MISTRAL_API_KEY }}",
		},

	}

	// Mock CredentialResolver
	oldResolver := CredentialResolver
	CredentialResolver = func(key string) (string, error) {
		if key == "cred.MISTRAL_API_KEY" {
			return "mistral_secret_key", nil
		}
		return "", fmt.Errorf("not found")
	}
	defer func() { CredentialResolver = oldResolver }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTemplate(tt.template, variables); got != tt.want {
				t.Errorf("ParseTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTemplate_Env(t *testing.T) {
	os.Setenv("TEST_VAR", "env_value")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "env no space",
			template: "{{env.TEST_VAR}}",
			want:     "env_value",
		},
		{
			name:     "env with spaces",
			template: "{{ env.TEST_VAR }}",
			want:     "env_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTemplate(tt.template, nil); got != tt.want {
				t.Errorf("ParseTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
