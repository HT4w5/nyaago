package server

import (
	"fmt"

	"github.com/HT4w5/nyaago/pkg/db"
)

const sizeCacheLimit = 1024

// Cached record of resource sizes
type sizeCache struct {
	sizeMap map[string]int64
	db      db.DBAdapter
}

func makeSizeCache(db db.DBAdapter) *sizeCache {
	return &sizeCache{
		sizeMap: make(map[string]int64),
		db:      db,
	}
}

func (c *sizeCache) GetSize(url string) (int64, error) {
	s, ok := c.sizeMap[url]
	if ok {
		return s, nil
	}

	res, err := c.db.GetResource(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get resource: %w", err)
	}

	if len(res.URL) == 0 {
		return 0, nil // not found
	}

	// Clear cache
	if len(c.sizeMap) > sizeCacheLimit {
		c.sizeMap = make(map[string]int64)
	}

	// Register size
	c.sizeMap[url] = res.Size

	return res.Size, nil
}
