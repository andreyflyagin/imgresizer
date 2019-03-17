package main

import (
	"github.com/hashicorp/golang-lru"
	"sync/atomic"
	"time"

	"log"
)

type cacheItem struct {
	item    []byte
	hash    string
	created time.Time
}

type cache struct {
	currSize int64
	store    *lru.Cache
	maxSize  int64
	liveTime time.Duration
}

func newCache(maxSize int64, liveTime time.Duration) *cache {
	res := cache{maxSize: maxSize, liveTime: liveTime}

	fn := func(key interface{}, value interface{}) {
		log.Printf("[INFO] cache purged %s", key)
		size := len(value.(cacheItem).item)
		atomic.AddInt64(&res.currSize, -1*int64(size))
	}

	var err error
	if res.store, err = lru.NewWithEvict(1000, fn); err != nil {
		log.Fatalf("[ERROR] failed init cache %v", err)
	}
	return &res
}

func (c *cache) add(key string, data []byte, hash string) {
	d := cacheItem{
		item:    data,
		hash:    hash,
		created: time.Now(),
	}

	c.store.Add(key, d)
	atomic.AddInt64(&c.currSize, int64(len(data)))

	log.Printf("[INFO] added to cache %d bytes", len(data))

	if c.currSize > c.maxSize {
		for atomic.LoadInt64(&c.currSize) > c.maxSize {
			c.store.RemoveOldest()
		}
	}
}

func (c *cache) get(key string) ([]byte, string) {
	if b, ok := c.store.Get(key); ok {
		v := b.(cacheItem)
		if v.created.Add(c.liveTime).After(time.Now()) {
			log.Printf("[INFO] cache hit %s", key)
			return v.item, v.hash
		}
		c.store.Remove(key)
	}
	return nil, ""
}
