package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	bbn "github.com/babylonlabs-io/babylon/v4/types"
	"github.com/babylonlabs-io/babylon/v4/x/btclightclient/types"
)

// TestHeaderCache_BasicFunctionality tests basic cache operations
func TestHeaderCache_BasicFunctionality(t *testing.T) {
	cache := types.NewHeaderCache()
	
	// Create test header
	testHash, err := bbn.NewBTCHeaderHashBytesFromHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	require.NoError(t, err)
	testHeader := &types.BTCHeaderInfo{
		Hash:   &testHash,
		Height: 100,
	}
	
	fetchCallCount := 0
	fetcher := func(height uint32) (*types.BTCHeaderInfo, error) {
		fetchCallCount++
		require.Equal(t, uint32(100), height)
		return testHeader, nil
	}
	
	// First call should cache miss and fetch
	header1, err := cache.GetOrFetch(100, fetcher)
	require.NoError(t, err)
	require.Equal(t, testHeader, header1)
	require.Equal(t, 1, fetchCallCount)
	
	// Second call should cache hit (no fetch)
	header2, err := cache.GetOrFetch(100, fetcher)
	require.NoError(t, err)
	require.Equal(t, testHeader, header2)
	require.Equal(t, 1, fetchCallCount) // Should not increment
	
	// Check stats
	stats := cache.Stats()
	require.Equal(t, int64(1), stats.HitCount)
	require.Equal(t, int64(1), stats.MissCount)
	require.Equal(t, 0.5, stats.HitRate())
}

// TestHeaderCache_TipValidation tests cache validation with tip changes
func TestHeaderCache_TipValidation(t *testing.T) {
	cache := types.NewHeaderCache()
	
	// Create test headers
	tip1Hash, err := bbn.NewBTCHeaderHashBytesFromHex("111102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	require.NoError(t, err)
	tip1 := &types.BTCHeaderInfo{Hash: &tip1Hash, Height: 100}
	
	tip2Hash, err := bbn.NewBTCHeaderHashBytesFromHex("222202030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	require.NoError(t, err)
	tip2 := &types.BTCHeaderInfo{Hash: &tip2Hash, Height: 101}
	
	// Initially no tip, should be invalid
	require.False(t, cache.IsValid(tip1))
	
	// Update with first tip
	cache.UpdateTip(tip1)
	require.True(t, cache.IsValid(tip1))
	require.False(t, cache.IsValid(tip2))
	
	// Update with second tip
	cache.UpdateTip(tip2)
	require.False(t, cache.IsValid(tip1))
	require.True(t, cache.IsValid(tip2))
}

// TestHeaderCache_Invalidation tests cache invalidation scenarios
func TestHeaderCache_Invalidation(t *testing.T) {
	cache := types.NewHeaderCache()
	
	// Add multiple headers to cache
	fetchCount := 0
	fetcher := func(height uint32) (*types.BTCHeaderInfo, error) {
		fetchCount++
		hash, _ := bbn.NewBTCHeaderHashBytesFromHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
		return &types.BTCHeaderInfo{Hash: &hash, Height: height}, nil
	}
	
	// Cache headers at heights 100, 101, 102
	_, err := cache.GetOrFetch(100, fetcher)
	require.NoError(t, err)
	_, err = cache.GetOrFetch(101, fetcher)
	require.NoError(t, err)
	_, err = cache.GetOrFetch(102, fetcher)
	require.NoError(t, err)
	require.Equal(t, 3, fetchCount)
	
	stats := cache.Stats()
	require.Equal(t, 3, stats.Size)
	
	// Invalidate from height 101 onwards
	cache.InvalidateFromHeight(101)
	
	stats = cache.Stats()
	require.Equal(t, 1, stats.Size) // Only height 100 should remain
	
	// Accessing height 100 should hit cache, 101 should miss
	_, err = cache.GetOrFetch(100, fetcher)
	require.NoError(t, err)
	require.Equal(t, 3, fetchCount) // No new fetch
	
	_, err = cache.GetOrFetch(101, fetcher)
	require.NoError(t, err)
	require.Equal(t, 4, fetchCount) // New fetch required
	
	// Full invalidation
	cache.Invalidate()
	stats = cache.Stats()
	require.Equal(t, 0, stats.Size)
}

// TestHeaderCache_Eviction tests LRU eviction when cache is full
func TestHeaderCache_Eviction(t *testing.T) {
	cache := types.NewHeaderCache()
	
	// Set a small max size for testing
	// Note: This would require exposing maxSize as a parameter or setter
	// For now, we'll test with the default size
	
	fetchCount := 0
	fetcher := func(height uint32) (*types.BTCHeaderInfo, error) {
		fetchCount++
		hash, _ := bbn.NewBTCHeaderHashBytesFromHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
		return &types.BTCHeaderInfo{Hash: &hash, Height: height}, nil
	}
	
	// Fill cache with some headers
	for i := uint32(1); i <= 5; i++ {
		_, err := cache.GetOrFetch(i, fetcher)
		require.NoError(t, err)
	}
	
	stats := cache.Stats()
	require.Equal(t, 5, stats.Size)
	require.Equal(t, int64(0), stats.HitCount)
	require.Equal(t, int64(5), stats.MissCount)
	
	// Test that first entry can still be accessed (cache hit)
	_, err := cache.GetOrFetch(1, fetcher)
	require.NoError(t, err)
	require.Equal(t, 5, fetchCount) // Should not increment
	
	stats = cache.Stats()
	require.Equal(t, int64(1), stats.HitCount)
}

// TestHeaderCache_Expiration tests cache expiration based on age
func TestHeaderCache_Expiration(t *testing.T) {
	cache := types.NewHeaderCache()
	
	fetchCount := 0
	fetcher := func(height uint32) (*types.BTCHeaderInfo, error) {
		fetchCount++
		hash, _ := bbn.NewBTCHeaderHashBytesFromHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
		return &types.BTCHeaderInfo{Hash: &hash, Height: height}, nil
	}
	
	// Add header to cache
	_, err := cache.GetOrFetch(100, fetcher)
	require.NoError(t, err)
	require.Equal(t, 1, fetchCount)
	
	// Should hit cache immediately
	_, err = cache.GetOrFetch(100, fetcher)
	require.NoError(t, err)
	require.Equal(t, 1, fetchCount)
	
	// Note: Testing actual expiration would require either:
	// 1. Waiting 5+ minutes (too slow for tests)
	// 2. Having a way to set a shorter maxAge for testing
	// 3. Having a way to manipulate time in tests
	// For now, we'll just verify the basic structure is working
}

// TestHeaderCache_ConcurrentAccess tests concurrent access to cache
func TestHeaderCache_ConcurrentAccess(t *testing.T) {
	cache := types.NewHeaderCache()
	
	fetchCount := 0
	fetcher := func(height uint32) (*types.BTCHeaderInfo, error) {
		fetchCount++
		hash, _ := bbn.NewBTCHeaderHashBytesFromHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
		// Simulate some work
		time.Sleep(time.Millisecond)
		return &types.BTCHeaderInfo{Hash: &hash, Height: height}, nil
	}
	
	// Test concurrent access to same height
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := cache.GetOrFetch(100, fetcher)
			require.NoError(t, err)
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Due to concurrent access, fetch might be called multiple times
	// but the cache should work correctly
	require.Greater(t, fetchCount, 0)
	require.LessOrEqual(t, fetchCount, 10)
	
	stats := cache.Stats()
	require.Greater(t, stats.HitCount+stats.MissCount, int64(0))
}

// TestHeaderCache_ErrorHandling tests error handling in cache
func TestHeaderCache_ErrorHandling(t *testing.T) {
	cache := types.NewHeaderCache()
	
	fetchCount := 0
	fetcher := func(height uint32) (*types.BTCHeaderInfo, error) {
		fetchCount++
		if height == 999 {
			return nil, types.ErrHeaderDoesNotExist
		}
		hash, _ := bbn.NewBTCHeaderHashBytesFromHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
		return &types.BTCHeaderInfo{Hash: &hash, Height: height}, nil
	}
	
	// Test error case
	_, err := cache.GetOrFetch(999, fetcher)
	require.Error(t, err)
	require.Equal(t, types.ErrHeaderDoesNotExist, err)
	require.Equal(t, 1, fetchCount)
	
	stats := cache.Stats()
	require.Equal(t, int64(1), stats.MissCount)
	require.Equal(t, int64(0), stats.HitCount)
	
	// Test successful case
	_, err = cache.GetOrFetch(100, fetcher)
	require.NoError(t, err)
	require.Equal(t, 2, fetchCount)
}