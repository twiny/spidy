package crawler

import "sync"

// Store struct
type Store struct {
	data map[string]struct{}
	mu   *sync.Mutex
}

/*
NewStore func
*/
func NewStore() *Store {
	return &Store{
		data: make(map[string]struct{}),
		mu:   new(sync.Mutex),
	}
}

/*
Found checks if key exists on store.
*/
func (s *Store) Found(key string) bool {
	s.mu.Lock()
	_, found := s.data[key]
	s.mu.Unlock()
	return found
}

/*
Save add value to store.
*/
func (s *Store) Save(key string) {
	s.mu.Lock()
	s.data[key] = struct{}{}
	s.mu.Unlock()
}

/*
Delete a value from store.
*/
func (s *Store) Delete(key string) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

/*
Close
*/
func (s *Store) Close() {
	s = nil
}
