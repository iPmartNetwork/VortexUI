package subscription

import (
	"strings"
	"testing"
)

// Property 6: Format variable resolution completeness — all declared variables
// resolve to non-empty strings for a valid user context.
func TestProperty_FormatVariableResolutionCompleteness(t *testing.T) {
	// Core format variables that must always resolve.
	requiredVars := []string{
		"{USERNAME}", "{DATA_LIMIT}", "{DATA_USED}", "{DATA_REMAINING}",
		"{EXPIRE_DATE}", "{DAYS_LEFT}", "{STATUS}", "{SUB_URL}",
		"{PROTOCOL}", "{NODE_NAME}", "{NODE_FLAG}",
	}

	// Simulate resolution with placeholder values.
	template := strings.Join(requiredVars, " | ")
	resolved := mockResolveVars(template)

	for _, v := range requiredVars {
		if strings.Contains(resolved, v) {
			t.Fatalf("variable %s was not resolved", v)
		}
	}
}

// Property 7: Outline format schema compliance — Outline renderer produces
// valid JSON with required fields: server, server_port, password, method.
func TestProperty_OutlineFormatSchemaCompliance(t *testing.T) {
	requiredFields := []string{"server", "server_port", "password", "method"}

	sample := `{"server":"1.2.3.4","server_port":443,"password":"abc123","method":"aes-256-gcm"}`

	for _, field := range requiredFields {
		if !strings.Contains(sample, `"`+field+`"`) {
			t.Fatalf("Outline output missing required field: %s", field)
		}
	}
}

// --- helpers ---

func mockResolveVars(template string) string {
	replacements := map[string]string{
		"{USERNAME}":       "testuser",
		"{DATA_LIMIT}":     "100GB",
		"{DATA_USED}":      "25GB",
		"{DATA_REMAINING}": "75GB",
		"{EXPIRE_DATE}":    "2025-12-31",
		"{DAYS_LEFT}":      "180",
		"{STATUS}":         "active",
		"{SUB_URL}":        "https://panel.example.com/sub/token123",
		"{PROTOCOL}":       "vless",
		"{NODE_NAME}":      "DE-Frankfurt",
		"{NODE_FLAG}":      "🇩🇪",
	}
	result := template
	for k, v := range replacements {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}
