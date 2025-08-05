package artifactory

import (
	"fmt"
	"testing"
	"time"

	l "github.com/peimanja/artifactory_exporter/logger"
)

func TestResponseCache(t *testing.T) {
	// Test creating cache with different configurations
	t.Run("Cache enabled", func(t *testing.T) {
		cache := NewResponseCache(true, 5*time.Minute, 30*time.Second)
		if cache == nil {
			t.Error("Expected cache to be created when enabled")
		}
	})

	t.Run("Cache disabled", func(t *testing.T) {
		cache := NewResponseCache(false, 5*time.Minute, 30*time.Second)
		if cache != nil {
			t.Error("Expected cache to be nil when disabled")
		}
	})
}

func TestCachedOperations(t *testing.T) {
	logger := l.New(l.Config{
		Format: l.FormatDefault,
		Level:  "debug",
	})

	cache := NewResponseCache(true, 1*time.Second, 500*time.Millisecond) // Short TTL for testing
	if cache == nil {
		t.Fatal("Failed to create cache for testing")
	}

	t.Run("Cache miss and set", func(t *testing.T) {
		key := "test_key_1"
		cached := NewCached(key, cache, logger)
		defer cached.Close()

		// Should not have cached response initially
		_, exists := cached.GetCachedResponse()
		if exists {
			t.Error("Expected cache miss for new key")
		}

		// Cache a response
		testResponse := &ApiResponse{
			Body:   []byte("test response"),
			NodeId: "test-node-1",
		}
		cached.CacheResponse(testResponse)

		// Should now have cached response
		cachedResp, exists := cached.GetCachedResponse()
		if !exists {
			t.Error("Expected cached response after setting")
		}

		if string(cachedResp.Body) != "test response" {
			t.Errorf("Cached body = %s, want 'test response'", string(cachedResp.Body))
		}

		if cachedResp.NodeId != "test-node-1" {
			t.Errorf("Cached NodeId = %s, want 'test-node-1'", cachedResp.NodeId)
		}
	})

	t.Run("Cache expiration", func(t *testing.T) {
		key := "test_key_2"
		cached := NewCached(key, cache, logger)
		defer cached.Close()

		// Cache a response
		testResponse := &ApiResponse{
			Body:   []byte("expiring response"),
			NodeId: "test-node-2",
		}
		cached.CacheResponse(testResponse)

		// Should have cached response
		_, exists := cached.GetCachedResponse()
		if !exists {
			t.Error("Expected cached response immediately after setting")
		}

		// Wait for expiration (TTL is 1 second)
		time.Sleep(1100 * time.Millisecond)

		// Should no longer have cached response
		_, exists = cached.GetCachedResponse()
		if exists {
			t.Error("Expected cache miss after TTL expiration")
		}
	})

	t.Run("Cache pruning", func(t *testing.T) {
		// Add multiple entries
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("prune_test_%d", i)
			cached := NewCached(key, cache, logger)
			testResponse := &ApiResponse{
				Body:   []byte(fmt.Sprintf("response %d", i)),
				NodeId: fmt.Sprintf("node-%d", i),
			}
			cached.CacheResponse(testResponse)
			cached.Close()
		}

		// Wait for expiration
		time.Sleep(1100 * time.Millisecond)

		// Prune expired entries
		pruned := cache.Prune()

		if pruned == 0 {
			t.Error("Expected some entries to be pruned")
		}

		t.Logf("Pruned %d expired cache entries", pruned)
	})
}

func TestCachedChannels(t *testing.T) {
	logger := l.New(l.Config{
		Format: l.FormatDefault,
		Level:  "debug",
	})

	cache := NewResponseCache(true, 5*time.Minute, 30*time.Second)
	if cache == nil {
		t.Fatal("Failed to create cache for testing")
	}

	t.Run("Channel operations", func(t *testing.T) {
		key := "channel_test"
		cached := NewCached(key, cache, logger)
		defer cached.Close()

		// Test that channels are created
		if cached.responses == nil {
			t.Error("Expected responses channel to be created")
		}

		if cached.errors == nil {
			t.Error("Expected errors channel to be created")
		}

		// Test sending response through channel
		go func() {
			testResponse := &ApiResponse{
				Body:   []byte("channel response"),
				NodeId: "channel-node",
			}
			cached.responses <- testResponse
		}()

		// Test receiving from channel with timeout
		select {
		case resp := <-cached.responses:
			if string(resp.Body) != "channel response" {
				t.Errorf("Channel response body = %s, want 'channel response'", string(resp.Body))
			}
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for response from channel")
		}
	})

	t.Run("Error channel", func(t *testing.T) {
		key := "error_test"
		cached := NewCached(key, cache, logger)
		defer cached.Close()

		// Test sending error through channel
		go func() {
			cached.errors <- fmt.Errorf("test error")
		}()

		// Test receiving error from channel
		select {
		case err := <-cached.errors:
			if err.Error() != "test error" {
				t.Errorf("Channel error = %v, want 'test error'", err)
			}
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for error from channel")
		}
	})
}
