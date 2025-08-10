package dns

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	config := CacheConfig{
		Enabled:         true,
		MaxSize:         10,
		DefaultTTL:      300 * time.Second,
		MaxTTL:          24 * time.Hour,
		MinTTL:          10 * time.Second,
		CleanupInterval: 5 * time.Minute,
		EvictionPolicy:  "lru",
	}
	
	cache := NewCache(config)
	
	// Test cache miss
	question := DNSQuestion{
		Name:  "example.com",
		Type:  TypeA,
		Class: ClassIN,
	}
	
	_, found := cache.Get(question)
	if found {
		t.Error("Expected cache miss, got hit")
	}
	
	// Test cache set and get
	response := &DNSResponse{
		ID:       1234,
		Question: question,
		Answers: []DNSRecord{
			{
				Name:  "example.com",
				Type:  TypeA,
				Class: ClassIN,
				TTL:   300,
				Data:  []byte{192, 168, 1, 1},
			},
		},
		ResponseCode: RCodeNoError,
	}
	
	cache.Set(question, response, 300*time.Second)
	
	entry, found := cache.Get(question)
	if !found {
		t.Error("Expected cache hit, got miss")
	}
	
	if entry.Response.ID != response.ID {
		t.Errorf("Expected response ID %d, got %d", response.ID, entry.Response.ID)
	}
	
	if len(entry.Response.Answers) != 1 {
		t.Errorf("Expected 1 answer, got %d", len(entry.Response.Answers))
	}
}

func TestCache_Expiration(t *testing.T) {
	config := CacheConfig{
		Enabled:         true,
		MaxSize:         10,
		DefaultTTL:      1 * time.Second, // Short TTL for testing
		MaxTTL:          24 * time.Hour,
		MinTTL:          1 * time.Second,
		CleanupInterval: 5 * time.Minute,
		EvictionPolicy:  "lru",
	}
	
	cache := NewCache(config)
	
	question := DNSQuestion{
		Name:  "example.com",
		Type:  TypeA,
		Class: ClassIN,
	}
	
	response := &DNSResponse{
		ID:           1234,
		Question:     question,
		ResponseCode: RCodeNoError,
	}
	
	// Set with short TTL
	cache.Set(question, response, 1*time.Second)
	
	// Should be found immediately
	_, found := cache.Get(question)
	if !found {
		t.Error("Expected cache hit immediately after set")
	}
	
	// Wait for expiration
	time.Sleep(2 * time.Second)
	
	// Should be expired now
	_, found = cache.Get(question)
	if found {
		t.Error("Expected cache miss after expiration")
	}
}

func TestCache_LRUEviction(t *testing.T) {
	config := CacheConfig{
		Enabled:         true,
		MaxSize:         3, // Small cache for testing eviction
		DefaultTTL:      300 * time.Second,
		MaxTTL:          24 * time.Hour,
		MinTTL:          10 * time.Second,
		CleanupInterval: 5 * time.Minute,
		EvictionPolicy:  "lru",
	}
	
	cache := NewCache(config)
	
	// Fill cache to capacity
	for i := 1; i <= 3; i++ {
		question := DNSQuestion{
			Name:  fmt.Sprintf("example%d.com", i),
			Type:  TypeA,
			Class: ClassIN,
		}
		
		response := &DNSResponse{
			ID:           uint16(i),
			Question:     question,
			ResponseCode: RCodeNoError,
		}
		
		cache.Set(question, response, 300*time.Second)
	}
	
	// All three should be in cache
	for i := 1; i <= 3; i++ {
		question := DNSQuestion{
			Name:  fmt.Sprintf("example%d.com", i),
			Type:  TypeA,
			Class: ClassIN,
		}
		
		_, found := cache.Get(question)
		if !found {
			t.Errorf("Expected to find example%d.com in cache", i)
		}
	}
	
	// Add fourth item (should evict least recently used)
	question4 := DNSQuestion{
		Name:  "example4.com",
		Type:  TypeA,
		Class: ClassIN,
	}
	
	response4 := &DNSResponse{
		ID:           4,
		Question:     question4,
		ResponseCode: RCodeNoError,
	}
	
	cache.Set(question4, response4, 300*time.Second)
	
	// First item should be evicted (least recently used)
	question1 := DNSQuestion{
		Name:  "example1.com",
		Type:  TypeA,
		Class: ClassIN,
	}
	
	_, found := cache.Get(question1)
	if found {
		t.Error("Expected example1.com to be evicted from cache")
	}
	
	// Fourth item should be in cache
	_, found = cache.Get(question4)
	if !found {
		t.Error("Expected example4.com to be in cache")
	}
}

func TestCache_Clear(t *testing.T) {
	config := CacheConfig{
		Enabled:         true,
		MaxSize:         10,
		DefaultTTL:      300 * time.Second,
		MaxTTL:          24 * time.Hour,
		MinTTL:          10 * time.Second,
		CleanupInterval: 5 * time.Minute,
		EvictionPolicy:  "lru",
	}
	
	cache := NewCache(config)
	
	// Add some entries
	for i := 1; i <= 5; i++ {
		question := DNSQuestion{
			Name:  fmt.Sprintf("example%d.com", i),
			Type:  TypeA,
			Class: ClassIN,
		}
		
		response := &DNSResponse{
			ID:           uint16(i),
			Question:     question,
			ResponseCode: RCodeNoError,
		}
		
		cache.Set(question, response, 300*time.Second)
	}
	
	// Clear cache
	cache.Clear()
	
	// All entries should be gone
	for i := 1; i <= 5; i++ {
		question := DNSQuestion{
			Name:  fmt.Sprintf("example%d.com", i),
			Type:  TypeA,
			Class: ClassIN,
		}
		
		_, found := cache.Get(question)
		if found {
			t.Errorf("Expected example%d.com to be cleared from cache", i)
		}
	}
	
	// Stats should be reset
	stats := cache.GetStats()
	if stats.Size != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", stats.Size)
	}
}

func TestCache_Stats(t *testing.T) {
	config := CacheConfig{
		Enabled:         true,
		MaxSize:         10,
		DefaultTTL:      300 * time.Second,
		MaxTTL:          24 * time.Hour,
		MinTTL:          10 * time.Second,
		CleanupInterval: 5 * time.Minute,
		EvictionPolicy:  "lru",
	}
	
	cache := NewCache(config)
	
	question := DNSQuestion{
		Name:  "example.com",
		Type:  TypeA,
		Class: ClassIN,
	}
	
	response := &DNSResponse{
		ID:           1234,
		Question:     question,
		ResponseCode: RCodeNoError,
	}
	
	// Test miss
	_, found := cache.Get(question)
	if found {
		t.Error("Expected cache miss")
	}
	
	// Add entry
	cache.Set(question, response, 300*time.Second)
	
	// Test hit
	_, found = cache.Get(question)
	if !found {
		t.Error("Expected cache hit")
	}
	
	// Check stats
	stats := cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	
	if stats.Size != 1 {
		t.Errorf("Expected cache size 1, got %d", stats.Size)
	}
	
	expectedHitRate := float64(1) / float64(2) // 1 hit out of 2 total requests
	if stats.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate %f, got %f", expectedHitRate, stats.HitRate)
	}
}