package store

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	dbus "github.com/godbus/dbus/v5"
)

func TestTranslateLinuxNoSecretStore(t *testing.T) {
	cases := map[string]error{
		"no session bus":       errors.New("dbus: couldn't determine address of session bus"),
		"dbus-launch missing":  &exec.Error{Name: "dbus-launch", Err: exec.ErrNotFound},
		"no provider":          dbus.Error{Name: "org.freedesktop.DBus.Error.ServiceUnknown", Body: []any{"The name org.freedesktop.secrets was not provided by any .service files"}},
		"no provider, pointer": &dbus.Error{Name: "org.freedesktop.DBus.Error.ServiceUnknown"},
		"name has no owner":    dbus.Error{Name: "org.freedesktop.DBus.Error.NameHasNoOwner"},
		"wrapped":              fmt.Errorf("get: %w", dbus.Error{Name: "org.freedesktop.DBus.Error.ServiceUnknown"}),
	}
	for name, in := range cases {
		if got := translate("linux", in); !errors.Is(got, ErrNoSecretStore) {
			t.Errorf("%s: got %v, want ErrNoSecretStore", name, got)
		}
	}
}

func TestTranslateLinuxLocked(t *testing.T) {
	cases := map[string]error{
		"is locked":        dbus.Error{Name: "org.freedesktop.Secret.Error.IsLocked", Body: []any{"Cannot create an item in a locked collection"}},
		"unlock dismissed": errors.New("failed to unlock correct collection '/org/freedesktop/secrets/aliases/default'"),
	}
	for name, in := range cases {
		if got := translate("linux", in); !errors.Is(got, ErrLocked) {
			t.Errorf("%s: got %v, want ErrLocked", name, got)
		}
	}
}

func TestTranslateLinuxKeepsOtherErrors(t *testing.T) {
	other := errors.New("boom")
	if got := translate("linux", other); got != other {
		t.Errorf("got %v, want it unchanged", got)
	}
	// Linux-only mappings must not leak to other platforms.
	noBus := errors.New("dbus: couldn't determine address of session bus")
	if got := translate("darwin", noBus); got != noBus {
		t.Errorf("darwin: got %v, want it unchanged", got)
	}
}
