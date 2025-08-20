package keeper

import (
	"context"

	"github.com/babylonlabs-io/babylon/v3/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BTCLightClientHook implements the BTCLightClientHooks interface
// to track when BTC light client state is modified
type BTCLightClientHook struct {
	keeper *Keeper
}

// NewBTCLightClientHook creates a new BTC light client hook
func NewBTCLightClientHook(keeper *Keeper) *BTCLightClientHook {
	return &BTCLightClientHook{
		keeper: keeper,
	}
}

// AfterBTCHeaderInserted is called after a BTC header is inserted
func (h BTCLightClientHook) AfterBTCHeaderInserted(ctx context.Context, headerInfo *types.BTCHeaderInfo) {
	h.markBTCLightClientModified(ctx)
}

// AfterBTCRollBack is called after a BTC rollback
func (h BTCLightClientHook) AfterBTCRollBack(ctx context.Context, rollbackFrom, rollbackTo *types.BTCHeaderInfo) {
	h.markBTCLightClientModified(ctx)
}

// AfterBTCRollForward is called after a BTC roll forward
func (h BTCLightClientHook) AfterBTCRollForward(ctx context.Context, headerInfo *types.BTCHeaderInfo) {
	h.markBTCLightClientModified(ctx)
}

// markBTCLightClientModified sets a flag in transient store indicating
// that the BTC light client state was modified in this block
func (h BTCLightClientHook) markBTCLightClientModified(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.keeper.SetBTCLightClientModified(sdkCtx)
}