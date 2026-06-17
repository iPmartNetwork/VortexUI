package domain

import (
	"time"

	"github.com/google/uuid"
)

// ISPPreset identifies a known ISP for auto-tuned TLS trick settings.
type ISPPreset string

const (
	ISPHamrahAval ISPPreset = "hamrah_aval" // MCI
	ISPIrancell   ISPPreset = "irancell"
	ISPMokhaberat ISPPreset = "mokhaberat"  // TCI
	ISPShatel     ISPPreset = "shatel"
	ISPAsiatech   ISPPreset = "asiatech"
	ISPCustom     ISPPreset = "custom"
)

// TLSTrickProfile is a reusable configuration for TLS fragmentation and
// anti-DPI techniques, supporting ISP-specific presets.
type TLSTrickProfile struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	ISP         ISPPreset `json:"isp"`
	Description string    `json:"description,omitempty"`

	// Fragment settings
	FragmentEnabled   bool   `json:"fragment_enabled"`
	FragmentSize      string `json:"fragment_size"`      // e.g. "10-50"
	FragmentInterval  string `json:"fragment_interval"`  // e.g. "10-20" ms
	FragmentPackets   string `json:"fragment_packets"`   // "tlshello" | "1-3"

	// Mux settings
	MuxEnabled     bool `json:"mux_enabled"`
	MuxConcurrency int  `json:"mux_concurrency"`
	MuxXudp        bool `json:"mux_xudp"`

	// TLS fingerprint
	UTLSFingerprint string `json:"utls_fingerprint"` // "chrome" | "firefox" | "safari" | "random"

	// Padding
	PaddingEnabled bool   `json:"padding_enabled"`
	PaddingSize    string `json:"padding_size"` // e.g. "100-200"

	// ECH (Encrypted Client Hello)
	ECHEnabled bool   `json:"ech_enabled"`
	ECHConfig  string `json:"ech_config,omitempty"` // base64 ECH config

	// HTTP/2 specific
	H2Fingerprint string `json:"h2_fingerprint,omitempty"` // "chrome" | "firefox"

	// Auto-detect: if true the system probes the best settings
	AutoDetect bool `json:"auto_detect"`

	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// ISPPresetDefaults returns recommended defaults for known ISPs.
func ISPPresetDefaults(isp ISPPreset) TLSTrickProfile {
	base := TLSTrickProfile{
		FragmentEnabled: true,
		MuxEnabled:      true,
		MuxConcurrency:  8,
		UTLSFingerprint: "chrome",
		PaddingEnabled:  true,
		Enabled:         true,
	}
	switch isp {
	case ISPHamrahAval:
		base.Name = "Hamrah Aval (MCI)"
		base.FragmentSize = "10-50"
		base.FragmentInterval = "10-20"
		base.FragmentPackets = "tlshello"
		base.PaddingSize = "100-200"
	case ISPIrancell:
		base.Name = "Irancell (MTN)"
		base.FragmentSize = "1-3"
		base.FragmentInterval = "1-5"
		base.FragmentPackets = "1-3"
		base.PaddingSize = "50-100"
		base.MuxXudp = true
	case ISPMokhaberat:
		base.Name = "Mokhaberat (TCI)"
		base.FragmentSize = "50-100"
		base.FragmentInterval = "20-50"
		base.FragmentPackets = "tlshello"
		base.PaddingSize = "200-300"
		base.ECHEnabled = true
	case ISPShatel:
		base.Name = "Shatel"
		base.FragmentSize = "10-30"
		base.FragmentInterval = "5-15"
		base.FragmentPackets = "tlshello"
		base.PaddingSize = "100-150"
	case ISPAsiatech:
		base.Name = "Asiatech"
		base.FragmentSize = "20-60"
		base.FragmentInterval = "10-30"
		base.FragmentPackets = "1-5"
		base.PaddingSize = "100-200"
	default:
		base.Name = "Custom"
		base.FragmentSize = "10-50"
		base.FragmentInterval = "10-20"
		base.FragmentPackets = "tlshello"
		base.PaddingSize = "100-200"
	}
	base.ISP = isp
	return base
}
