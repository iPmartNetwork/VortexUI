package migrations_test

import (
	"strings"
	"testing"

	"github.com/vortexui/vortexui/migrations"
)

func TestResellerMigrationsHaveGooseHeaders(t *testing.T) {
	for _, name := range []string{
		"0021_reseller_enhancements.sql",
		"0022_reseller_advanced.sql",
		"0023_reseller_policy_suspend.sql",
	} {
		b, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.HasPrefix(string(b), "-- +goose Up\n") && !strings.HasPrefix(string(b), "-- +goose Up\r\n") {
			t.Fatalf("%s must start with -- +goose Up; got %q", name, string(b[:min(40, len(b))]))
		}
		if !strings.Contains(string(b), "-- +goose Down") {
			t.Fatalf("%s missing -- +goose Down", name)
		}
	}
}
