package dns

import (
	"fmt"
	"sync"
	"time"
)

// Cache implements the DNSCache interface with LRU eviction
type Cache struct {
	mu          sync.RWMutex
	entries     map[string]*cacheNode
	head        *cacheNode
	tail        *cacheNode
	maxSize     int
	currentSize int
	stats       CacheStats
	config      CacheConfig
}

// cacheNode represents a node in the LRU cache
type cacheNode struct {
	key   string
	entry *CacheEntry
	prev  *cacheNode
	next  *cacheNode
}

// NewCache creates a new DNS cache
func NewCache(config CacheConfig) DNSCache {
	cache := &Cache{
		entries: make(map[string]*cacheNode),
		maxSize: config.MaxSize,
		config:  config,
		stats: CacheStats{
			MaxSize: config.MaxSize,
		},
	}

	// Initialize doubly linked list with sentinel nodes
	cache.head = &cacheNode{}
	cache.tail = &cacheNode{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Get retrieves a cached response
func (c *Cache) Get(question DNSQuestion) (*CacheEntry, bool) {
	if !c.config.Enabled {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(question)
	node, exists := c.entries[key]

	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if entry is expired
	if time.Now().After(node.entry.ExpiresAt) {
		c.removeNode(node)
		delete(c.entries, key)
		c.currentSize--
		c.stats.Misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.moveToFront(node)

	// Update access statistics
	node.entry.AccessTime = time.Now()
	node.entry.HitCount++

	c.stats.Hits++
	c.updateHitRate()

	return node.entry, true
}

// Set stores a response in cache
func (c *Cache) Set(question DNSQuestion, response *DNSResponse, ttl time.Duration) {
	if !c.config.Enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(question)

	// Check if entry already exists
	if node, exists := c.entries[key]; exists {
		// Update existing entry
		node.entry.Response = response
		node.entry.ExpiresAt = time.Now().Add(ttl)
		node.entry.AccessTime = time.Now()
		c.moveToFront(node)
		return
	}

	// Evict if cache is full
	if c.currentSize >= c.maxSize {
		c.evictLRU()
	}

	// Create new entry
	entry := &CacheEntry{
		Response:   response,
		ExpiresAt:  time.Now().Add(ttl),
		AccessTime: time.Now(),
		HitCount:   0,
	}

	// Create new node and add to front
	node := &cacheNode{
		key:   key,
		entry: entry,
	}

	c.addToFront(node)
	c.entries[key] = node
	c.currentSize++

	c.stats.Size = c.currentSize
}

// Delete removes an entry from cache
func (c *Cache) Delete(question DNSQuestion) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(question)
	if node, exists := c.entries[key]; exists {
		c.removeNode(node)
		delete(c.entries, key)
		c.currentSize--
		c.stats.Size = c.currentSize
	}
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheNode)
	c.head.next = c.tail
	c.tail.prev = c.head
	c.currentSize = 0

	c.stats.Size = 0
	c.stats.Hits = 0
	c.stats.Misses = 0
	c.stats.Evictions = 0
	c.stats.HitRate = 0.0
}

// GetStats returns cache statistics
func (c *Cache) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = c.currentSize
	return &stats
}

// Cleanup removes expired entries
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var toRemove []string

	// Find expired entries
	for key, node := range c.entries {
		if now.After(node.entry.ExpiresAt) {
			toRemove = append(toRemove, key)
		}
	}

	// Remove expired entries
	for _, key := range toRemove {
		if node, exists := c.entries[key]; exists {
			c.removeNode(node)
			delete(c.entries, key)
			c.currentSize--
		}
	}

	c.stats.Size = c.currentSize
	c.stats.LastCleanup = now
}

// makeKey creates a cache key from a DNS question
func (c *Cache) makeKey(question DNSQuestion) string {
	return fmt.Sprintf("%s:%d:%d", question.Name, question.Type, question.Class)
}

// addToFront adds a node to the front of the LRU list
func (c *Cache) addToFront(node *cacheNode) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

// removeNode removes a node from the LRU list
func (c *Cache) removeNode(node *cacheNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

// moveToFront moves a node to the front of the LRU list
func (c *Cache) moveToFront(node *cacheNode) {
	c.removeNode(node)
	c.addToFront(node)
}

// evictLRU evicts the least recently used entry
func (c *Cache) evictLRU() {
	if c.tail.prev != c.head {
		lru := c.tail.prev
		c.removeNode(lru)
		delete(c.entries, lru.key)
		c.currentSize--
		c.stats.Evictions++
	}
}

// updateHitRate calculates the current cache hit rate
func (c *Cache) updateHitRate() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.Hits) / float64(total)
	}
}
