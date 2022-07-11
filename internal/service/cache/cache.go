package cache

import (
	"time"

	//

	//
	"github.com/twiny/carbon"
)

// Cache
type Cache struct {
	ttl time.Duration
	db  *carbon.Cache
}

// NewCache
func NewCache(ttl time.Duration, dir string) (*Cache, error) {
	db, err := carbon.NewCache(dir)
	if err != nil {
		return nil, err
	}
	return &Cache{
		ttl: ttl,
		db:  db,
	}, nil
}

// HasChecked
func (c *Cache) HasChecked(name string) bool {
	// first check if domain is in cache
	b, err := c.db.Get(name)
	if err != nil || b == nil {
		// if not found save to cache
		if err := c.db.Set(name, []byte(name), c.ttl); err != nil {
			return false
		}
		return false
	}
	return true
}

// Close
func (c *Cache) Close() error {
	c.db.Close()
	return nil
}
