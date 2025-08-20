package keeper_test

import (
	"testing"
	"math/rand"

	testkeeper "github.com/babylonlabs-io/babylon/v4/testutil/keeper"
	"github.com/babylonlabs-io/babylon/v4/testutil/datagen"
	"github.com/babylonlabs-io/babylon/v4/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

// TestGetMainChainFrom_CacheWorking tests that the cache actually works
func TestGetMainChainFrom_CacheWorking(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	r := rand.New(rand.NewSource(10))

	// Generate and insert a simple chain with known heights
	numHeaders := 5
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	// Generate first header with height 100
	baseHeader := datagen.GenRandomBTCHeaderInfo(r) 
	baseHeader.Height = 100
	headers[0] = baseHeader
	
	// Generate chain with consecutive heights
	for i := 1; i < numHeaders; i++ {
		header := datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
		header.Height = 100 + uint32(i)  // Force consecutive heights
		headers[i] = header
	}
	
	// Insert headers into keeper
	keeper.InsertHeaderInfos(ctx, headers)
	
	// Verify tip is set correctly
	tip := keeper.GetTipInfo(ctx)
	require.NotNil(t, tip)
	require.Equal(t, uint32(104), tip.Height) // Last header should be height 104
	
	t.Logf("Tip height: %d", tip.Height)
	
	// Test cache statistics
	initialStats := keeper.HeaderCache().Stats()
	require.Equal(t, int64(0), initialStats.HitCount)
	require.Equal(t, int64(0), initialStats.MissCount)
	
	// First call - should populate cache
	result1 := keeper.GetMainChainFrom(ctx, 102) // Should get headers 102, 103, 104
	require.Equal(t, 3, len(result1))
	
	// Verify we got the right headers
	require.Equal(t, uint32(102), result1[0].Height)
	require.Equal(t, uint32(103), result1[1].Height)
	require.Equal(t, uint32(104), result1[2].Height)
	
	stats1 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(0), stats1.HitCount)   // No hits on first call
	require.Equal(t, int64(3), stats1.MissCount)  // 3 misses (headers 102, 103, 104)
	
	// Second call with same parameters - should hit cache
	result2 := keeper.GetMainChainFrom(ctx, 102)
	require.Equal(t, 3, len(result2))
	require.Equal(t, result1[0].Height, result2[0].Height)
	require.Equal(t, result1[1].Height, result2[1].Height)
	require.Equal(t, result1[2].Height, result2[2].Height)
	
	stats2 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(3), stats2.HitCount)   // All 3 should be cache hits  
	require.Equal(t, int64(3), stats2.MissCount)  // Miss count unchanged
	require.Equal(t, 0.5, stats2.HitRate())       // 3 hits / 6 total = 0.5
	
	// Third call with different start but overlapping range
	result3 := keeper.GetMainChainFrom(ctx, 101) // Should get headers 101, 102, 103, 104
	require.Equal(t, 4, len(result3))
	
	stats3 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(6), stats3.HitCount)   // 3 more hits (102, 103, 104 were cached)
	require.Equal(t, int64(4), stats3.MissCount)  // 1 more miss (101 was not cached)
	
	t.Logf("Final stats - Hits: %d, Misses: %d, Hit Rate: %.2f", 
		stats3.HitCount, stats3.MissCount, stats3.HitRate())
}

// TestGetMainChainFrom_CacheInvalidation tests cache invalidation when new headers are added
func TestGetMainChainFrom_CacheInvalidation(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	r := rand.New(rand.NewSource(10))

	// Generate initial chain
	numHeaders := 3
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	baseHeader := datagen.GenRandomBTCHeaderInfo(r)
	baseHeader.Height = 200
	headers[0] = baseHeader
	
	for i := 1; i < numHeaders; i++ {
		header := datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
		header.Height = 200 + uint32(i)
		headers[i] = header
	}
	
	keeper.InsertHeaderInfos(ctx, headers)
	
	// Initial call to populate cache
	result1 := keeper.GetMainChainFrom(ctx, 200)
	require.Equal(t, 3, len(result1))
	
	oldTip := keeper.GetTipInfo(ctx)
	require.Equal(t, uint32(202), oldTip.Height)
	
	// Add a new header to extend the chain
	newHeader := datagen.GenRandomBTCHeaderInfoWithParent(r, headers[2])
	newHeader.Height = 203
	keeper.InsertHeaderInfos(ctx, []*types.BTCHeaderInfo{newHeader})
	
	// Cache should detect tip change and return updated results
	result2 := keeper.GetMainChainFrom(ctx, 200)
	require.Equal(t, 4, len(result2)) // Now should include the new header
	
	newTip := keeper.GetTipInfo(ctx)
	require.Equal(t, uint32(203), newTip.Height)
	
	// Verify the last header is the new one
	lastHeader := result2[len(result2)-1]
	require.Equal(t, uint32(203), lastHeader.Height)
}