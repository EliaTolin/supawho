// Package store manages Supabase account tokens and a comma-separated index
// of account names in the OS secret store.
//
// The index format and service name are kept identical to the original bash
// implementation so existing users keep their accounts after upgrading.
package store

import (
	"errors"
	"strings"
)

const (
	// Service is the secret-store service under which all accounts live.
	// It matches the original bash script for backward compatibility.
	Service = "supabase-cli-accounts"
	// indexKey holds the comma-separated list of saved account names.
	indexKey = "_account_list"
)

// ErrNotFound is returned by a backend when a key does not exist.
var ErrNotFound = errors.New("not found")

// backend is the low-level secret key/value store (service is fixed per backend).
type backend interface {
	get(key string) (string, error) // returns ErrNotFound if the key is missing
	set(key, value string) error
	delete(key string) error // no error if the key is missing
}

// Store is the high-level account store shared by every backend.
type Store struct {
	b backend
}

func newStore(b backend) *Store { return &Store{b: b} }

// List returns the saved account names in insertion order (nil if none).
func (s *Store) List() ([]string, error) {
	raw, err := s.b.get(indexKey)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, ","), nil
}

func (s *Store) saveList(names []string) error {
	return s.b.set(indexKey, strings.Join(names, ","))
}

// Get returns the token for name, or ErrNotFound.
func (s *Store) Get(name string) (string, error) {
	return s.b.get(name)
}

// Add stores (or overwrites) the token for name and ensures it is in the index.
func (s *Store) Add(name, token string) error {
	if err := s.b.set(name, token); err != nil {
		return err
	}
	names, err := s.List()
	if err != nil {
		return err
	}
	if !contains(names, name) {
		names = append(names, name)
	}
	return s.saveList(names)
}

// Rename moves the token from oldName to newName and updates the index in place.
// Returns ErrNotFound if oldName does not exist.
func (s *Store) Rename(oldName, newName string) error {
	token, err := s.b.get(oldName)
	if errors.Is(err, ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if err := s.b.set(newName, token); err != nil {
		return err
	}
	if oldName != newName {
		if err := s.b.delete(oldName); err != nil {
			return err
		}
	}
	names, err := s.List()
	if err != nil {
		return err
	}
	for i, n := range names {
		if n == oldName {
			names[i] = newName
		}
	}
	return s.saveList(names)
}

// Remove deletes the token for name and drops it from the index.
func (s *Store) Remove(name string) error {
	if err := s.b.delete(name); err != nil {
		return err
	}
	names, err := s.List()
	if err != nil {
		return err
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		if n != name {
			out = append(out, n)
		}
	}
	return s.saveList(out)
}

func contains(names []string, target string) bool {
	for _, n := range names {
		if n == target {
			return true
		}
	}
	return false
}
