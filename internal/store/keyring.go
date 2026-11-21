package store

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	dbus "github.com/godbus/dbus/v5"
	"github.com/zalando/go-keyring"
)

// keyringBackend is the real OS secret store backend:
// macOS Keychain, Linux Secret Service, Windows Credential Manager.
type keyringBackend struct {
	service string
}

// NewKeyring returns a Store backed by the OS secret store.
func NewKeyring() *Store {
	return newStore(&keyringBackend{service: Service})
}

func (k *keyringBackend) name() string { return BackendKeyring }

func (k *keyringBackend) get(key string) (string, error) {
	v, err := keyring.Get(k.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return v, translate(runtime.GOOS, err)
}

func (k *keyringBackend) set(key, value string) error {
	return translate(runtime.GOOS, keyring.Set(k.service, key, value))
}

func (k *keyringBackend) delete(key string) error {
	err := keyring.Delete(k.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return translate(runtime.GOOS, err)
}

// securityInteractionNotAllowed is the exit status of /usr/bin/security for
// errSecInteractionNotAllowed (OSStatus -25308, truncated to its low byte):
// the keychain is locked and macOS cannot show the unlock dialog, typically
// over SSH or in a session without a graphical login.
const securityInteractionNotAllowed = 36

// translate maps opaque backend failures to sentinel errors the CLI can
// explain. Everything it does not recognise is returned unchanged.
func translate(goos string, err error) error {
	var exit *exec.ExitError
	if goos == "darwin" && errors.As(err, &exit) && exit.ExitCode() == securityInteractionNotAllowed {
		return fmt.Errorf("%w (%v)", ErrLocked, err)
	}
	if goos == "linux" && err != nil {
		if sentinel := linuxSentinel(err); sentinel != nil {
			return fmt.Errorf("%w (%v)", sentinel, err)
		}
	}
	return err
}

// linuxSentinel recognises the Secret Service failures (via godbus and
// go-keyring) that mean the store is missing or locked, or returns nil.
func linuxSentinel(err error) error {
	switch dbusErrorName(err) {
	case "org.freedesktop.DBus.Error.ServiceUnknown",
		"org.freedesktop.DBus.Error.NameHasNoOwner",
		"org.freedesktop.DBus.Error.Spawn.ServiceNotFound":
		return ErrNoSecretStore
	case "org.freedesktop.Secret.Error.IsLocked":
		return ErrLocked
	}
	var execErr *exec.Error
	if errors.As(err, &execErr) && execErr.Name == "dbus-launch" {
		return ErrNoSecretStore
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "couldn't determine address of session bus"):
		return ErrNoSecretStore
	case strings.Contains(msg, "failed to unlock correct collection"):
		return ErrLocked
	}
	return nil
}

// dbusErrorName returns the D-Bus error name in err's chain, or "".
// godbus returns dbus.Error both by value and by pointer.
func dbusErrorName(err error) string {
	var v dbus.Error
	if errors.As(err, &v) {
		return v.Name
	}
	var p *dbus.Error
	if errors.As(err, &p) && p != nil {
		return p.Name
	}
	return ""
}
