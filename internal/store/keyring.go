package store

import (
	"errors"

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
	return v, err
}

func (k *keyringBackend) set(key, value string) error {
	return keyring.Set(k.service, key, value)
}

func (k *keyringBackend) delete(key string) error {
	err := keyring.Delete(k.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
