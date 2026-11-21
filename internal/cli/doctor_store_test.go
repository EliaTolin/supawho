package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/EliaTolin/supawho/internal/store"
)

type jsonSystem struct {
	System []struct {
		Name, Status, Detail, Hint string
	} `json:"system"`
}

func doctorJSON(t *testing.T, h *harness, wantCode int) jsonSystem {
	t.Helper()
	h.run(t, wantCode, "doctor", "--json")
	var got jsonSystem
	if err := json.Unmarshal(h.out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, h.out.String())
	}
	if len(got.System) == 0 || got.System[0].Name != "secret_store" {
		t.Fatalf("first system check is not secret_store: %+v", got.System)
	}
	return got
}

func TestDoctorSecretStorePasses(t *testing.T) {
	h := newDoctorHarness("a", "sbp_ok_a")
	got := doctorJSON(t, h, 0)
	if c := got.System[0]; c.Status != statusPass || !strings.Contains(c.Detail, "memory") {
		t.Fatalf("secret_store = %+v, want pass naming the backend", c)
	}
}

func TestDoctorReportsMissingSecretStoreOnLinux(t *testing.T) {
	h := newDoctorHarness()
	h.app.Store = store.NewFailing(store.ErrNoSecretStore)
	h.app.GOOS = "linux"
	got := doctorJSON(t, h, 1)
	c := got.System[0]
	if c.Status != statusFail || !strings.Contains(c.Detail, "no Secret Service") {
		t.Fatalf("secret_store = %+v, want fail with 'no Secret Service'", c)
	}
	if !strings.Contains(c.Hint, "SUPAWHO_BACKEND=pass") {
		t.Fatalf("hint = %q, want the pass backend", c.Hint)
	}
}

func TestDoctorReportsLockedKeychainOnMac(t *testing.T) {
	h := newDoctorHarness()
	h.app.Store = store.NewFailing(store.ErrLocked)
	h.app.GOOS = "darwin"
	c := doctorJSON(t, h, 1).System[0]
	if c.Status != statusFail || !strings.Contains(c.Hint, "security unlock-keychain") {
		t.Fatalf("secret_store = %+v", c)
	}
}

func TestDoctorReportsMissingPass(t *testing.T) {
	h := newDoctorHarness()
	h.app.Store = store.NewFailing(store.ErrPassNotInstalled)
	c := doctorJSON(t, h, 1).System[0]
	if c.Status != statusFail || !strings.Contains(c.Hint, "pass init") {
		t.Fatalf("secret_store = %+v", c)
	}
}
