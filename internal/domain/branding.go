package domain

// Branding holds panel customization options stored in settings. Admins can
// change the panel name, logo URL, accent color, and footer text via the UI.
type Branding struct {
	PanelName   string `json:"panel_name"`   // displayed in login/sidebar (default "VortexUI")
	LogoURL     string `json:"logo_url"`     // custom logo URL (empty = default SVG)
	AccentColor string `json:"accent_color"` // CSS color for primary accent (e.g. "#6366f1")
	FooterText  string `json:"footer_text"`  // copyright footer override
	FaviconURL  string `json:"favicon_url"`  // custom favicon
}

// DefaultBranding returns the built-in look.
func DefaultBranding() Branding {
	return Branding{
		PanelName:   "VortexUI",
		AccentColor: "#6366f1",
		FooterText:  "© 2026 iPmart Network. All rights reserved.",
	}
}
