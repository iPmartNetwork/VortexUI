package domain

// DeepLinkConfig stores deep link and QR configuration.
type DeepLinkConfig struct {
	Enabled      bool   `json:"enabled"`
	Scheme       string `json:"scheme"`        // e.g. "vortex"
	BaseURL      string `json:"base_url"`      // e.g. "https://panel.example.com"
	AppStoreURL  string `json:"app_store_url"` // fallback for iOS
	PlayStoreURL string `json:"play_store_url"` // fallback for Android
	QRLogoURL    string `json:"qr_logo_url"`   // custom logo in center of QR
}

// DefaultDeepLinkConfig returns sensible defaults.
func DefaultDeepLinkConfig() DeepLinkConfig {
	return DeepLinkConfig{
		Enabled: true,
		Scheme:  "vortex",
	}
}
