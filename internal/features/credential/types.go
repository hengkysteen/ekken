package credential

// Credential stores an encrypted key-value secret usable by nodes and assistants.
type Credential struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`            // Display name
	Key       string   `json:"key"`             // Credential key identifier (e.g. OPENAI_API_KEY)
	Value     string   `json:"value,omitempty"` // Decrypted secret; omitted in list responses
	Tags      []string `json:"tags"`            // Categorization tags
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}
