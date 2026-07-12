package decoy

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestReloadFromInboundsServesStatic(t *testing.T) {
	s := NewServer()
	t.Cleanup(func() { _ = s.Stop(context.Background()) })

	in := domain.Inbound{
		Enabled: true,
		Raw: map[string]any{
			"decoy": map[string]any{
				"mode":        "static",
				"static_html": "<html><body>ok</body></html>",
			},
		},
	}
	if err := s.ReloadFromInbounds([]domain.Inbound{in}); err != nil {
		t.Fatalf("reload: %v", err)
	}

	resp, err := http.Get("http://" + domain.DefaultDecoyListen)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html><body>ok</body></html>" {
		t.Fatalf("body = %q", string(body))
	}
}
