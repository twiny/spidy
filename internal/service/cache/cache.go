package cache

import (
	"time"

	"github.com/twiny/carbon"
)

type Cache struct {
	ttl time.Duration
	db  *carbon.Cache
}

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

func (c *Cache) HasChecked(name string) bool {
	b, err := c.db.Get(name)
	if err != nil || b == nil {
		return c.db.Set(name, []byte(name), c.ttl) != nil
	}
	return true
}

func (c *Cache) Close() error {
	c.db.Close()
	return nil
}
