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

// TestAllMigrationsHaveGooseUpHeader guards against a repeat of the 0042
// incident: a migration file missing "-- +goose Up" fails goose's parser at
// panel startup, crash-looping the whole service. Every embedded .sql file
// must start with the directive.
func TestAllMigrationsHaveGooseUpHeader(t *testing.T) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	found := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		found++
		b, err := migrations.FS.ReadFile(e.Name())
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if !strings.HasPrefix(string(b), "-- +goose Up\n") && !strings.HasPrefix(string(b), "-- +goose Up\r\n") {
			t.Fatalf("%s must start with -- +goose Up; got %q", e.Name(), string(b[:min(40, len(b))]))
		}
	}
	if found == 0 {
		t.Fatal("no .sql migration files found in embedded FS")
	}
}
