package artifactory

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type CacheEntry struct {
	data      *ApiResponse
	timestamp time.Time
}

// ResponseCache stores API responses and allows thread-safe access.
type ResponseCache struct {
	mutex   sync.RWMutex
	data    map[string]CacheEntry
	ttl     time.Duration // duration before entries go stale (conf.CacheTTL)
	timeout time.Duration // request timeout for cached requests (conf.CacheTimeout)
}

func (r *ResponseCache) Prune() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var removed = 0
	now := time.Now()
	for key, entry := range r.data {
		if now.Sub(entry.timestamp) > r.ttl {
			delete(r.data, key)
			removed++
		}
	}
	return removed
}

func NewResponseCache(useCache bool, ttl time.Duration, timeout time.Duration) *ResponseCache {
	if !useCache {
		return nil
	}
	return &ResponseCache{
		data:    make(map[string]CacheEntry),
		ttl:     ttl,
		timeout: timeout,
	}
}

func (r *ResponseCache) GetCachedResponse(key string) (*ApiResponse, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	resp, exists := r.data[key]
	if !exists {
		return nil, false
	} else if time.Since(resp.timestamp) > r.ttl {
		// Entry is expired, ignore it
		return nil, false
	}
	return resp.data, true
}

func (r *ResponseCache) SetCachedResponse(key string, response *ApiResponse) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.data[key] = CacheEntry{
		data:      response,
		timestamp: time.Now(),
	}
}

// Cached is a wrapper for http requests that allows caching of API responses.
// The actual request runs in a separate goroutine which will update the cache whenever a response is received.
// If we hit a timeout, we can instead return a cached response if available.
type Cached struct {
	errors      chan error
	responses   chan *ApiResponse
	timeout     context.Context
	cancel      context.CancelFunc
	stopTimeout func() bool

	responseCache *ResponseCache
	cacheKey      string

	logger *slog.Logger
}

func NewCached(cacheKey string, r *ResponseCache, logger *slog.Logger) *Cached {
	errors := make(chan error, 1)
	responses := make(chan *ApiResponse, 1)

	var timeout context.Context
	var cancel context.CancelFunc
	var stopTimeout func() bool
	// Only use timeout if response cache is configured.
	if r != nil {
		timeout, cancel = context.WithTimeout(context.Background(), r.timeout)
		stopTimeout = context.AfterFunc(timeout, func() {
			logger.Warn("Cache request timed out", "timeout", r.timeout)
			errors <- fmt.Errorf("request timed out after %d seconds", int(r.timeout.Seconds()))
		})
	} else {
		timeout, cancel = context.WithCancel(context.Background())
		stopTimeout = func() bool {
			// No timeout function to stop, so just return true
			return true
		}
	}

	return &Cached{
		errors:        errors,
		responses:     responses,
		timeout:       timeout,
		cancel:        cancel,
		stopTimeout:   stopTimeout,
		responseCache: r,
		cacheKey:      cacheKey,
		logger:        logger,
	}
}

func (c *Cached) Close() {
	if !c.stopTimeout() {
		// timeout func already triggered
		<-c.errors
	}
	c.cancel()
	close(c.errors)
	close(c.responses)
}

func (c *Cached) CacheResponse(response *ApiResponse) {
	if c.responseCache != nil {
		c.logger.Debug("Caching response for key", "key", c.cacheKey)
		c.responseCache.SetCachedResponse(c.cacheKey, response)
	}
}

func (c *Cached) GetCachedResponse() (*ApiResponse, bool) {
	if c.responseCache != nil {
		c.logger.Debug("Getting cached response for key", "key", c.cacheKey)
		return c.responseCache.GetCachedResponse(c.cacheKey)
	}
	return nil, false
}
