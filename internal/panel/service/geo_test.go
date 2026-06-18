package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeResolver struct{ country string }

func (f fakeResolver) Country(string) string { return f.country }

type fakeGeoRepo struct {
	calls   int
	country string
	ip      string
}

func (r *fakeGeoRepo) Upsert(_ context.Context, _ uuid.UUID, country, ip string) error {
	r.calls++
	r.country = country
	r.ip = ip
	return nil
}

func TestRecordUserIPUpsertsResolvedCountry(t *testing.T) {
	repo := &fakeGeoRepo{}
	svc := NewGeoService(fakeResolver{country: "DE"}, repo)
	svc.RecordUserIP(context.Background(), uuid.New(), "1.2.3.4")
	if repo.calls != 1 {
		t.Fatalf("expected 1 upsert, got %d", repo.calls)
	}
	if repo.country != "DE" || repo.ip != "1.2.3.4" {
		t.Errorf("upsert got country=%q ip=%q, want DE/1.2.3.4", repo.country, repo.ip)
	}
}

func TestRecordUserIPSkipsWhenUnresolved(t *testing.T) {
	repo := &fakeGeoRepo{}
	svc := NewGeoService(fakeResolver{country: ""}, repo)
	svc.RecordUserIP(context.Background(), uuid.New(), "10.0.0.1")
	if repo.calls != 0 {
		t.Fatalf("expected no upsert for unresolved IP, got %d", repo.calls)
	}
}

func TestRecordUserIPNilServiceIsSafe(t *testing.T) {
	var svc *GeoService
	svc.RecordUserIP(context.Background(), uuid.New(), "1.2.3.4") // must not panic
}
