package main

import (
	"sync"
)

type Cache struct {
	store map[string][]DnsRecord
	sync.RWMutex
}

func (c *Cache) Get(key string) ([]DnsRecord, bool) {
	c.RLock()
	defer c.RUnlock()

	record, bool := c.store[key]
	return record, bool
}

func (c *Cache) Set(key string, records []DnsRecord) {
	c.Lock()
	defer c.Unlock()

	c.store[key] = records
}

func (c *Cache) Reset() {
	c.Lock()
	defer c.Unlock()

	c.store = make(map[string][]DnsRecord)
}