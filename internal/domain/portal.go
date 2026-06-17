package domain

// PortalSession holds the auth state for an end-user logged into the self-service
// portal via their subscription token.
type PortalSession struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"` // JWT issued for portal access
}
