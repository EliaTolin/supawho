package store

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

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
	return err
}
