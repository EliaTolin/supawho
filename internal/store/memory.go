package store

// memBackend is an in-memory backend used by tests.
type memBackend struct {
	m map[string]string
}

// NewMemory returns a Store backed by an in-memory map. For tests only.
func NewMemory() *Store {
	return newStore(&memBackend{m: make(map[string]string)})
}

func (b *memBackend) get(key string) (string, error) {
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
