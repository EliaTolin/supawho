package store

// memBackend is an in-memory backend used by tests.
type memBackend struct {
	m map[string]string
	// errs makes get fail for specific keys, simulating a broken OS store.
	errs map[string]error
}

// NewMemory returns a Store backed by an in-memory map. For tests only.
func NewMemory() *Store {
	return newStore(&memBackend{m: make(map[string]string)})
}

func (b *memBackend) name() string { return "memory" }

func (b *memBackend) get(key string) (string, error) {
	if err, ok := b.errs[key]; ok {
		return "", err
	}
	v, ok := b.m[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (b *memBackend) set(key, value string) error {
	b.m[key] = value
	return nil
}

func (b *memBackend) delete(key string) error {
	delete(b.m, key)
	return nil
}

// failBackend fails every operation with a fixed error.
type failBackend struct{ err error }

// NewFailing returns a Store whose every operation fails with err. For tests only.
func NewFailing(err error) *Store {
	return newStore(&failBackend{err: err})
}

func (b *failBackend) name() string               { return "memory" }
func (b *failBackend) get(string) (string, error) { return "", b.err }
func (b *failBackend) set(string, string) error   { return b.err }
func (b *failBackend) delete(string) error        { return b.err }

// NewMemoryWithOrphans returns an in-memory Store whose index already lists
// orphans without a stored token, simulating secrets that vanished from the
// OS store. For tests only.
func NewMemoryWithOrphans(orphans ...string) *Store {
	s := NewMemory()
	_ = s.saveList(orphans)
	return s
}

// NewMemoryWithBrokenSecrets returns an in-memory Store whose index lists
// names but reading any of their tokens fails with err, simulating a locked
// or unreachable OS store. For tests only.
func NewMemoryWithBrokenSecrets(err error, names ...string) *Store {
	b := &memBackend{m: make(map[string]string), errs: make(map[string]error)}
	for _, n := range names {
		b.errs[n] = err
	}
	s := newStore(b)
	_ = s.saveList(names)
	return s
}
