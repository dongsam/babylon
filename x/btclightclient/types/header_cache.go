package types

import (
	"sync/atomic"
	"time"

	bbn "github.com/babylonlabs-io/babylon/v4/types"
)

// HeaderCache provides caching for individual BTC headers to eliminate
// duplicate KV store I/O operations across multiple GetMainChainFrom calls
type HeaderCache struct {
	// headers stores cached headers by height
	// TODO: need to verify height is unique for BTC headers
	headers map[uint32]*CachedHeader

	// tip state for cache validation
	tipHeight uint32
	tipHash   *bbn.BTCHeaderHashBytes

	// TODO: remove temporary statistics for benchmarking
	hitCount  int64
	missCount int64
}

// CachedHeader wraps a header with metadata
type CachedHeader struct {
	Header   *BTCHeaderInfo
	CachedAt time.Time
}

// NewHeaderCache creates a new header cache with default configuration
func NewHeaderCache() *HeaderCache {
	return &HeaderCache{
		headers: make(map[uint32]*CachedHeader),
	}
}

// GetOrFetch retrieves a header from cache or fetches it using the provided function
func (c *HeaderCache) GetOrFetch(height uint32, fetcher func(uint32) (*BTCHeaderInfo, error)) (*BTCHeaderInfo, error) {
	// Try cache first
	if cached, exists := c.headers[height]; exists {
		atomic.AddInt64(&c.hitCount, 1)
		return cached.Header, nil
	}

	// Cache miss or expired - fetch from source
	header, err := fetcher(height)
	if err != nil {
		atomic.AddInt64(&c.missCount, 1)
		return nil, err
	}

	// Store in cache
	if header != nil {
		c.headers[height] = &CachedHeader{
			Header:   header,
			CachedAt: time.Now(),
		}
	}

	atomic.AddInt64(&c.missCount, 1)
	return header, nil
}

// IsValid checks if the cache is valid for the current tip
func (c *HeaderCache) IsValid(currentTip *BTCHeaderInfo) bool {
	if currentTip == nil {
		return false
	}

	return c.tipHeight == currentTip.Height &&
		c.tipHash != nil &&
		c.tipHash.Eq(currentTip.Hash)
}

// UpdateTip updates the cache's tip state
func (c *HeaderCache) UpdateTip(tip *BTCHeaderInfo) {
	if tip != nil {
		c.tipHeight = tip.Height
		c.tipHash = tip.Hash
	}
}

// Invalidate clears all cached headers
func (c *HeaderCache) Invalidate() {
	c.headers = make(map[uint32]*CachedHeader)
}

// InvalidateFromHeight removes cached headers at or above the given height
func (c *HeaderCache) InvalidateFromHeight(height uint32) {
	for h := range c.headers {
		if h >= height {
			delete(c.headers, h)
		}
	}
}

// Stats returns cache statistics
func (c *HeaderCache) Stats() CacheStats {
	return CacheStats{
		Size:      len(c.headers),
		HitCount:  atomic.LoadInt64(&c.hitCount),
		MissCount: atomic.LoadInt64(&c.missCount),
		TipHeight: c.tipHeight,
		TipHash:   c.tipHash,
	}
}

// CacheStats provides cache metrics
type CacheStats struct {
	Size      int
	HitCount  int64
	MissCount int64
	TipHeight uint32
	TipHash   *bbn.BTCHeaderHashBytes
}

// HitRate returns the cache hit rate
func (stats CacheStats) HitRate() float64 {
	total := stats.HitCount + stats.MissCount
	if total == 0 {
		return 0
	}
	return float64(stats.HitCount) / float64(total)
}
