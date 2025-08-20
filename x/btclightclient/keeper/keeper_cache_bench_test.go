package keeper_test

import (
	"testing"
	"math/rand"

	testkeeper "github.com/babylonlabs-io/babylon/v4/testutil/keeper"
	"github.com/babylonlabs-io/babylon/v4/testutil/datagen"
	"github.com/babylonlabs-io/babylon/v4/x/btclightclient/types"
)

// BenchmarkGetMainChainFrom_WithCache benchmarks the cached version of GetMainChainFrom
func BenchmarkGetMainChainFrom_WithCache(b *testing.B) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(b)
	r := rand.New(rand.NewSource(10))

	// Generate and insert a longer chain for benchmarking
	numHeaders := 100
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	// Generate headers with consecutive heights starting from 1000
	baseHeader := datagen.GenRandomBTCHeaderInfo(r)
	baseHeader.Height = 1000
	headers[0] = baseHeader
	
	for i := 1; i < numHeaders; i++ {
		header := datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
		header.Height = 1000 + uint32(i)
		headers[i] = header
	}
	
	keeper.InsertHeaderInfos(ctx, headers)
	
	// Pre-populate cache with one call (simulate typical usage where cache gets populated)
	keeper.GetMainChainFrom(ctx, 1050)
	
	b.ResetTimer()
	
	// Benchmark repeated calls that should hit cache (simulating multiple consumers scenario)
	for i := 0; i < b.N; i++ {
		// Alternate between different start heights that have overlapping cached data
		startHeight := uint32(1050 + (i % 10))
		result := keeper.GetMainChainFrom(ctx, startHeight)
		if len(result) == 0 {
			b.Errorf("Expected non-empty result for height %d", startHeight)
		}
	}
}

// BenchmarkGetMainChainFrom_CacheVsNoCache compares performance with cache enabled vs theoretical no-cache
func BenchmarkGetMainChainFrom_CacheVsNoCache(b *testing.B) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(b)
	r := rand.New(rand.NewSource(10))

	// Generate headers
	numHeaders := 50
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	baseHeader := datagen.GenRandomBTCHeaderInfo(r) 
	baseHeader.Height = 2000
	headers[0] = baseHeader
	
	for i := 1; i < numHeaders; i++ {
		header := datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
		header.Height = 2000 + uint32(i)
		headers[i] = header
	}
	
	keeper.InsertHeaderInfos(ctx, headers)
	
	b.Run("WithCache", func(b *testing.B) {
		// Pre-populate cache
		keeper.GetMainChainFrom(ctx, 2025)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// This should hit cache for most headers
			result := keeper.GetMainChainFrom(ctx, 2025)
			if len(result) != 25 {
				b.Errorf("Expected 25 headers, got %d", len(result))
			}
		}
	})
	
	b.Run("FreshCacheEachTime", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Clear cache before each call to simulate no-cache scenario
			keeper.HeaderCache().Invalidate()
			result := keeper.GetMainChainFrom(ctx, 2025)
			if len(result) != 25 {
				b.Errorf("Expected 25 headers, got %d", len(result))
			}
		}
	})
}

// BenchmarkGetMainChainFrom_MultiConsumerScenario benchmarks the scenario where multiple consumers
// request overlapping header ranges (the primary use case for the optimization)
func BenchmarkGetMainChainFrom_MultiConsumerScenario(b *testing.B) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(b)
	r := rand.New(rand.NewSource(10))

	// Generate headers
	numHeaders := 30
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	baseHeader := datagen.GenRandomBTCHeaderInfo(r)
	baseHeader.Height = 3000
	headers[0] = baseHeader
	
	for i := 1; i < numHeaders; i++ {
		header := datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
		header.Height = 3000 + uint32(i)
		headers[i] = header
	}
	
	keeper.InsertHeaderInfos(ctx, headers)
	
	b.ResetTimer()
	
	// Simulate BroadcastBTCTimestamps scenario with 5 consumers requesting overlapping ranges
	for i := 0; i < b.N; i++ {
		// Consumer 1: from height 3010
		keeper.GetMainChainFrom(ctx, 3010)
		
		// Consumer 2: from height 3010 (same range - should be all cache hits)
		keeper.GetMainChainFrom(ctx, 3010)
		
		// Consumer 3: from height 3015 (subset of previous - should be cache hits)
		keeper.GetMainChainFrom(ctx, 3015)
		
		// Consumer 4: from height 3005 (extends the range - partial cache hits)
		keeper.GetMainChainFrom(ctx, 3005)
		
		// Consumer 5: from height 3020 (different range - some cache hits)
		keeper.GetMainChainFrom(ctx, 3020)
	}
}

// BenchmarkHeaderCache_GetOrFetch benchmarks the cache itself
func BenchmarkHeaderCache_GetOrFetch(b *testing.B) {
	cache := types.NewHeaderCache()
	r := rand.New(rand.NewSource(10))
	
	// Pre-populate cache with some headers
	for i := uint32(1); i <= 100; i++ {
		header := datagen.GenRandomBTCHeaderInfo(r)
		header.Height = i
		cache.GetOrFetch(i, func(height uint32) (*types.BTCHeaderInfo, error) {
			return header, nil
		})
	}
	
	b.ResetTimer()
	
	// Benchmark cache access (should be all hits)
	for i := 0; i < b.N; i++ {
		height := uint32(1 + (i % 100))
		cache.GetOrFetch(height, func(height uint32) (*types.BTCHeaderInfo, error) {
			b.Errorf("Should not be called - cache miss for height %d", height)
			return nil, nil
		})
	}
}