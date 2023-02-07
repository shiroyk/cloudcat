package bolt

import (
	"github.com/shiroyk/cloudcat/cache"
	"golang.org/x/exp/slog"
)

// Cache is an implementation of Cache that stores bytes in bolt.DB.
type Cache struct {
	db *DB
}

// Get returns the []byte and true, if not existing returns false.
func (c *Cache) Get(key string) (value []byte, ok bool) {
	var err error
	if value, err = c.db.Get([]byte(key)); err != nil || value == nil {
		return []byte{}, false
	}
	return value, true
}

// Set saves []byte to the cache with key.
func (c *Cache) Set(key string, value []byte) {
	err := c.db.Put([]byte(key), value)
	if err != nil {
		slog.Error("failed to set cache with key %s", err, key)
	}
}

// Del removes key from the cache.
func (c *Cache) Del(key string) {
	err := c.db.Delete([]byte(key))
	if err != nil {
		slog.Error("failed to delete cache with key %s", err, key)
	}
}

// NewCache returns a new Cache that will store items in bolt.DB.
func NewCache(path string) (cache.Cache, error) {
	db, err := NewDB(path, "cache", 0)
	if err != nil {
		return nil, err
	}
	return &Cache{db: db}, nil
}
