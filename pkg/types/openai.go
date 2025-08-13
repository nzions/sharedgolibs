package types

// Version represents the package version for types
const Version = "v0.1.0"

// OpenAIAPIKeyListResponse represents the response structure for listing OpenAI API keys
type OpenAIAPIKeyListResponse struct {
	Object  string         `json:"object"`
	Data    []OpenAIAPIKey `json:"data"`
	FirstID string         `json:"first_id"`
	LastID  string         `json:"last_id"`
	HasMore bool           `json:"has_more"`
}

// OpenAIAPIKey represents an individual OpenAI API key
type OpenAIAPIKey struct {
	Object        string               `json:"object"`
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	RedactedValue string               `json:"redacted_value"`
	CreatedAt     int64                `json:"created_at"`
	LastUsedAt    int64                `json:"last_used_at"`
	Owner         OpenAIServiceAccount `json:"owner"`
}

// OpenAIServiceAccount represents the owner of an API key
type OpenAIServiceAccount struct {
	Type      string `json:"type"`
	Object    string `json:"object"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Role      string `json:"role"`
}
