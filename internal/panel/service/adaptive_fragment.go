package service

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/vortexui/vortexui/internal/domain"
)

// FragmentProfile represents a tested fragment configuration with its success rate.
type FragmentProfile struct {
	Size     string  `json:"size"`     // e.g. "10-50"
	Interval string  `json:"interval"` // e.g. "10-20"
	Packets  string  `json:"packets"`  // e.g. "tlshello"
	Success  float64 `json:"success"`  // 0.0-1.0 success rate
	Samples  int     `json:"samples"`
}

// AdaptiveFragmentService tracks connection success rates per ISP and adjusts
// fragment settings dynamically. It runs as a background service.
type AdaptiveFragmentService struct {
	mu       sync.RWMutex
	profiles map[domain.ISPPreset][]FragmentProfile
	log      *slog.Logger
}

// NewAdaptiveFragmentService creates a new adaptive fragment service seeded with
// known-good defaults for major Iranian ISPs.
func NewAdaptiveFragmentService(log *slog.Logger) *AdaptiveFragmentService {
	if log == nil {
		log = slog.Default()
	}
	svc := &AdaptiveFragmentService{
		profiles: make(map[domain.ISPPreset][]FragmentProfile),
		log:      log,
	}
	// Seed with known-good defaults
	svc.seedDefaults()
	return svc
}

func (s *AdaptiveFragmentService) seedDefaults() {
	s.profiles[domain.ISPHamrahAval] = []FragmentProfile{
		{Size: "10-50", Interval: "10-20", Packets: "tlshello", Success: 0.85, Samples: 100},
		{Size: "100-200", Interval: "10-20", Packets: "tlshello", Success: 0.75, Samples: 50},
	}
	s.profiles[domain.ISPIrancell] = []FragmentProfile{
		{Size: "1-3", Interval: "1-5", Packets: "1-3", Success: 0.90, Samples: 100},
		{Size: "10-30", Interval: "5-15", Packets: "tlshello", Success: 0.80, Samples: 50},
	}
	s.profiles[domain.ISPMokhaberat] = []FragmentProfile{
		{Size: "50-100", Interval: "20-50", Packets: "tlshello", Success: 0.70, Samples: 80},
		{Size: "100-200", Interval: "30-60", Packets: "tlshello", Success: 0.65, Samples: 40},
	}
}

// BestFragment returns the highest-success fragment profile for an ISP.
func (s *AdaptiveFragmentService) BestFragment(isp domain.ISPPreset) *FragmentProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profiles, ok := s.profiles[isp]
	if !ok || len(profiles) == 0 {
		return nil
	}
	best := &profiles[0]
	for i := range profiles {
		if profiles[i].Success > best.Success {
			best = &profiles[i]
		}
	}
	return best
}

// RecordResult records a connection success/failure for an ISP's fragment profile.
// This allows the system to learn which settings work best over time.
func (s *AdaptiveFragmentService) RecordResult(isp domain.ISPPreset, size, interval string, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles := s.profiles[isp]
	for i := range profiles {
		if profiles[i].Size == size && profiles[i].Interval == interval {
			profiles[i].Samples++
			if success {
				// Exponential moving average
				profiles[i].Success = profiles[i].Success*0.95 + 0.05
			} else {
				profiles[i].Success = profiles[i].Success * 0.95
			}
			return
		}
	}
	// New profile discovered
	rate := 0.0
	if success {
		rate = 1.0
	}
	s.profiles[isp] = append(profiles, FragmentProfile{
		Size: size, Interval: interval, Packets: "tlshello", Success: rate, Samples: 1,
	})
}

// FormatFragment returns a comma-separated fragment string for the best profile.
func (s *AdaptiveFragmentService) FormatFragment(isp domain.ISPPreset) string {
	best := s.BestFragment(isp)
	if best == nil {
		return ""
	}
	return fmt.Sprintf("%s,%s,%s", best.Size, best.Interval, best.Packets)
}
