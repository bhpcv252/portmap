package registry

import (
	"strings"
	"testing"
	"time"
)

func newRegistry() *Registry {
	return &Registry{Version: 1, Claims: make(map[string]Claim)}
}

func TestAddClaim_NewPort(t *testing.T) {
	r := newRegistry()
	path := "/home/user/projects/myapp"

	if err := r.AddClaim("3000", "myapp", "frontend", "Next.js dev server", path, false); err != nil {
		t.Fatal(err)
	}

	c := r.GetClaim("3000")
	if c == nil {
		t.Fatal("expected claim to exist")
		return
	}
	if c.Project != "myapp" || c.Service != "frontend" {
		t.Errorf("unexpected claim fields: %+v", c)
	}
	if c.Path != path {
		t.Errorf("expected path %q, got %q", path, c.Path)
	}
}

func TestAddClaim_SameProjectSameService(t *testing.T) {
	r := newRegistry()
	r.Set(
		"3000",
		Claim{
			Project:     "myapp",
			Service:     "frontend",
			Description: "old desc",
			ClaimedAt:   time.Now().UTC().Add(-time.Hour),
		},
	)
	before := r.GetClaim("3000").ClaimedAt

	if err := r.AddClaim("3000", "myapp", "frontend", "new desc", "/path", false); err != nil {
		t.Fatal(err)
	}

	c := r.GetClaim("3000")
	if c.Description != "new desc" {
		t.Errorf("expected description updated, got %q", c.Description)
	}
	if !c.ClaimedAt.After(before) {
		t.Error("expected claimed_at to be updated")
	}
}

func TestAddClaim_SameProjectDifferentService(t *testing.T) {
	r := newRegistry()
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})

	err := r.AddClaim("3000", "myapp", "api", "", "/path", false)
	if err == nil {
		t.Fatal("expected conflict error for same project different service")
	}
}

func TestAddClaim_DifferentProject_NoForce(t *testing.T) {
	r := newRegistry()
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})

	err := r.AddClaim("3000", "otherapp", "frontend", "", "/path", false)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "port 3000 is already claimed by myapp/frontend") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestAddClaim_DifferentProject_Force(t *testing.T) {
	r := newRegistry()
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})

	if err := r.AddClaim("3000", "otherapp", "frontend", "", "/path", true); err != nil {
		t.Fatal(err)
	}

	c := r.GetClaim("3000")
	if c.Project != "otherapp" {
		t.Errorf("expected claim overwritten, got project %q", c.Project)
	}
}

func TestAddClaim_ClaimedAtPopulated(t *testing.T) {
	r := newRegistry()

	if err := r.AddClaim("3000", "myapp", "frontend", "", "/path", false); err != nil {
		t.Fatal(err)
	}

	c := r.GetClaim("3000")
	if c.ClaimedAt.IsZero() {
		t.Error("expected claimed_at to be set")
	}
}

func TestRemoveClaim_Exists(t *testing.T) {
	r := newRegistry()
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})

	r.RemoveClaim("3000")

	if r.GetClaim("3000") != nil {
		t.Error("expected claim to be removed")
	}
}

func TestRemoveClaim_NotFound(t *testing.T) {
	r := newRegistry()
	before := len(r.Claims)

	r.RemoveClaim("9999")

	if len(r.Claims) != before {
		t.Error("expected registry unchanged after removing non-existent port")
	}
}

func TestGetClaim_Found(t *testing.T) {
	r := newRegistry()
	r.Set(
		"3000",
		Claim{
			Project:     "myapp",
			Service:     "frontend",
			Description: "Next.js dev server",
			ClaimedAt:   time.Now().UTC(),
		},
	)

	c := r.GetClaim("3000")
	if c == nil {
		t.Fatal("expected claim, got nil")
		return
	}
	if c.Project != "myapp" || c.Service != "frontend" || c.Description != "Next.js dev server" {
		t.Errorf("unexpected claim fields: %+v", c)
	}
}

func TestGetClaim_NotFound(t *testing.T) {
	r := newRegistry()

	c := r.GetClaim("9999")
	if c != nil {
		t.Errorf("expected nil, got %+v", c)
	}
}

func TestFilterByProject_Match(t *testing.T) {
	r := newRegistry()
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})
	r.Set("3001", Claim{Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()})
	r.Set("8000", Claim{Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()})

	result := r.FilterByProject("myapp")
	if len(result) != 2 {
		t.Fatalf("expected 2 claims, got %d", len(result))
	}
}

func TestFilterByProject_NoMatch(t *testing.T) {
	r := newRegistry()
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})

	result := r.FilterByProject("nonexistent")
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %d claims", len(result))
	}
}

func TestFilterByProject_AllClaims(t *testing.T) {
	r := newRegistry()
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})
	r.Set("8000", Claim{Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()})

	result := r.FilterByProject("")
	if len(result) != 2 {
		t.Fatalf("expected all 2 claims, got %d", len(result))
	}
}
