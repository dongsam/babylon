package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonlabs-io/babylon/v4/testutil/datagen"
	testkeeper "github.com/babylonlabs-io/babylon/v4/testutil/keeper"
	"github.com/babylonlabs-io/babylon/v4/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

// TestGetMainChainFrom_CacheOptimization tests the cache optimization for GetMainChainFrom
func TestGetMainChainFrom_CacheOptimization(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	r := rand.New(rand.NewSource(10))

	// Generate and insert a chain of headers
	numHeaders := 10
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	// Generate base header
	baseHeader := datagen.GenRandomBTCHeaderInfoWithParent(r, nil)
	headers[0] = baseHeader
	
	// Generate chain of headers
	for i := 1; i < numHeaders; i++ {
		headers[i] = datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
	}
	
	// Insert headers into keeper
	keeper.InsertHeaderInfos(ctx, headers)
	
	// Verify tip is set correctly
	tip := keeper.GetTipInfo(ctx)
	require.NotNil(t, tip)
	require.Equal(t, headers[numHeaders-1].Height, tip.Height)
	require.True(t, headers[numHeaders-1].Hash.Eq(tip.Hash))
	
	// Test GetMainChainFrom with different start heights based on actual header heights
	baseHeight := headers[0].Height
	tipHeight := headers[numHeaders-1].Height
	midHeight := headers[5].Height
	nearTipHeight := headers[numHeaders-2].Height
	
	testCases := []struct {
		name        string
		startHeight uint32
		expected    int // number of headers expected
	}{
		{
			name:        "from base height (entire chain)",
			startHeight: baseHeight,
			expected:    numHeaders,
		},
		{
			name:        "from middle height",
			startHeight: midHeight,
			expected:    numHeaders - 5,
		},
		{
			name:        "from near tip",
			startHeight: nearTipHeight,
			expected:    2,
		},
		{
			name:        "from tip height",
			startHeight: tipHeight,
			expected:    1,
		},
		{
			name:        "from height higher than tip",
			startHeight: tipHeight + 5,
			expected:    0,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := keeper.GetMainChainFrom(ctx, tc.startHeight)
			require.Equal(t, tc.expected, len(result))
			
			// Verify the results are correct and in order
			for i, header := range result {
				expectedHeight := tc.startHeight + uint32(i)
				require.Equal(t, expectedHeight, header.Height)
				
				// Find corresponding original header
				var originalHeader *types.BTCHeaderInfo
				for _, h := range headers {
					if h.Height == expectedHeight {
						originalHeader = h
						break
					}
				}
				require.NotNil(t, originalHeader)
				require.True(t, header.Hash.Eq(originalHeader.Hash))
			}
		})
	}
}

// TestGetMainChainFrom_CacheEfficiency tests that cache reduces duplicate I/O
func TestGetMainChainFrom_CacheEfficiency(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	r := rand.New(rand.NewSource(10))

	// Generate and insert a chain of headers
	numHeaders := 20
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	// Generate base header
	baseHeader := datagen.GenRandomBTCHeaderInfoWithParent(r, nil)
	headers[0] = baseHeader
	
	// Generate chain of headers
	for i := 1; i < numHeaders; i++ {
		headers[i] = datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
	}
	
	// Insert headers into keeper
	keeper.InsertHeaderInfos(ctx, headers)
	
	// Get initial cache stats (should be empty)
	initialStats := keeper.HeaderCache().Stats()
	require.Equal(t, int64(0), initialStats.HitCount)
	require.Equal(t, int64(0), initialStats.MissCount)
	
	// First call to GetMainChainFrom - should populate cache
	result1 := keeper.GetMainChainFrom(ctx, 10)
	require.Equal(t, 10, len(result1)) // headers 10-19
	
	stats1 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(0), stats1.HitCount)   // No hits yet
	require.Equal(t, int64(10), stats1.MissCount) // 10 misses
	require.Equal(t, 10, stats1.Size)             // 10 cached headers
	
	// Second call with same start height - should hit cache
	result2 := keeper.GetMainChainFrom(ctx, 10)
	require.Equal(t, len(result1), len(result2))
	
	stats2 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(10), stats2.HitCount) // All 10 should be cache hits
	require.Equal(t, int64(10), stats2.MissCount) // Miss count unchanged
	require.Equal(t, 0.5, stats2.HitRate()) // 50% hit rate (10 hits / 20 total)
	
	// Third call with higher start height - should partially hit cache
	result3 := keeper.GetMainChainFrom(ctx, 15)
	require.Equal(t, 5, len(result3)) // headers 15-19
	
	stats3 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(15), stats3.HitCount) // 5 more hits (15-19 were cached)
	require.Equal(t, int64(10), stats3.MissCount) // Miss count unchanged
	
	// Fourth call with lower start height - should mostly hit cache with some misses
	result4 := keeper.GetMainChainFrom(ctx, 5)
	require.Equal(t, 15, len(result4)) // headers 5-19
	
	stats4 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(25), stats4.HitCount) // 10 more hits (10-19 were cached)
	require.Equal(t, int64(15), stats4.MissCount) // 5 more misses (5-9 were not cached)
}

// TestGetMainChainFrom_TipChanges tests cache behavior when tip changes
func TestGetMainChainFrom_TipChanges(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	r := rand.New(rand.NewSource(10))

	// Generate and insert initial chain
	numHeaders := 10
	headers := make([]*types.BTCHeaderInfo, numHeaders)
	
	// Generate base header
	baseHeader := datagen.GenRandomBTCHeaderInfoWithParent(r, nil)
	headers[0] = baseHeader
	
	// Generate chain of headers
	for i := 1; i < numHeaders; i++ {
		headers[i] = datagen.GenRandomBTCHeaderInfoWithParent(r, headers[i-1])
	}
	
	// Insert headers into keeper
	keeper.InsertHeaderInfos(ctx, headers)
	
	// Get initial result and cache stats
	result1 := keeper.GetMainChainFrom(ctx, 5)
	require.Equal(t, 5, len(result1))
	
	stats1 := keeper.HeaderCache().Stats()
	require.Equal(t, int64(5), stats1.MissCount)
	require.Equal(t, headers[numHeaders-1].Height, stats1.TipHeight)
	
	// Add a new header to extend the chain
	newHeader := datagen.GenRandomBTCHeaderInfoWithParent(r, headers[numHeaders-1])
	keeper.InsertHeaderInfos(ctx, []*types.BTCHeaderInfo{newHeader})
	
	// Cache should detect tip change and update
	result2 := keeper.GetMainChainFrom(ctx, 5)
	require.Equal(t, 6, len(result2)) // Now includes the new header
	
	stats2 := keeper.HeaderCache().Stats()
	require.Equal(t, newHeader.Height, stats2.TipHeight) // Tip should be updated
	
	// Verify the last header in result is the new header
	lastHeader := result2[len(result2)-1]
	require.True(t, lastHeader.Hash.Eq(newHeader.Hash))
	require.Equal(t, newHeader.Height, lastHeader.Height)
}

// TestGetMainChainFrom_EmptyChain tests behavior with empty chain
func TestGetMainChainFrom_EmptyChain(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	
	// Test with empty chain (no headers inserted)
	result := keeper.GetMainChainFrom(ctx, 0)
	require.Nil(t, result) // Should return nil when no tip exists
	
	result = keeper.GetMainChainFrom(ctx, 100)
	require.Nil(t, result) // Should return nil when no tip exists
}

