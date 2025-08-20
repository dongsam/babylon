package keeper_test

import (
	"testing"
	"math/rand"

	testkeeper "github.com/babylonlabs-io/babylon/v4/testutil/keeper"
	"github.com/babylonlabs-io/babylon/v4/testutil/datagen"
	"github.com/babylonlabs-io/babylon/v4/x/btclightclient/types"
	"github.com/stretchr/testify/require"
)

// TestGetMainChainFrom_Simple tests the basic functionality without complex scenarios
func TestGetMainChainFrom_Simple(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	r := rand.New(rand.NewSource(10))

	// Generate a single header
	header := datagen.GenRandomBTCHeaderInfo(r)
	
	// Insert header into keeper
	keeper.InsertHeaderInfos(ctx, []*types.BTCHeaderInfo{header})
	
	// Verify tip is set correctly
	tip := keeper.GetTipInfo(ctx)
	require.NotNil(t, tip)
	t.Logf("Tip height: %d, Header height: %d", tip.Height, header.Height)
	
	// Test GetMainChainFrom from height 0
	result := keeper.GetMainChainFrom(ctx, 0)
	t.Logf("Result length: %d", len(result))
	
	if len(result) > 0 {
		t.Logf("First header height: %d", result[0].Height)
	}
	
	// Test GetMainChainFrom from tip height
	result2 := keeper.GetMainChainFrom(ctx, tip.Height)
	t.Logf("Result2 length: %d", len(result2))
}