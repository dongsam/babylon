package keeper

import (
	"context"
	"fmt"

	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	btcctypes "github.com/babylonlabs-io/babylon/v3/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonlabs-io/babylon/v3/x/checkpointing/types"
	epochingtypes "github.com/babylonlabs-io/babylon/v3/x/epoching/types"
	"github.com/babylonlabs-io/babylon/v3/x/zoneconcierge/types"
)

func (k Keeper) ProveConsumerHeaderInEpoch(_ context.Context, header *types.IndexedHeader, epoch *epochingtypes.Epoch) (*cmtcrypto.ProofOps, error) {
	consumerHeaderKey := types.GetConsumerHeaderKey(header.ConsumerId, header.Height)
	_, _, proof, err := k.QueryStore(types.StoreKey, consumerHeaderKey, int64(epoch.GetSealerBlockHeight()))
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (k Keeper) ProveEpochInfo(epoch *epochingtypes.Epoch) (*cmtcrypto.ProofOps, error) {
	epochInfoKey := types.GetEpochInfoKey(epoch.EpochNumber)
	_, _, proof, err := k.QueryStore(epochingtypes.StoreKey, epochInfoKey, int64(epoch.GetSealerBlockHeight()))
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (k Keeper) ProveValSet(epoch *epochingtypes.Epoch) (*cmtcrypto.ProofOps, error) {
	valSetKey := types.GetValSetKey(epoch.EpochNumber)
	_, _, proof, err := k.QueryStore(checkpointingtypes.StoreKey, valSetKey, int64(epoch.GetSealerBlockHeight()))
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// ProveEpochSealed proves an epoch has been sealed, i.e.,
// - the epoch's validator set has a valid multisig over the sealer header
// - the epoch's validator set is committed to the sealer header's app_hash
// - the epoch's metadata is committed to the sealer header's app_hash
func (k Keeper) ProveEpochSealed(ctx context.Context, epochNumber uint64) (*types.ProofEpochSealed, error) {
	var (
		proof = &types.ProofEpochSealed{}
		err   error
	)

	// get the validator set of the sealed epoch
	proof.ValidatorSet, err = k.checkpointingKeeper.GetBLSPubKeySet(ctx, epochNumber)
	if err != nil {
		return nil, err
	}

	// get sealer header and the query height
	epoch, err := k.epochingKeeper.GetHistoricalEpoch(ctx, epochNumber)
	if err != nil {
		return nil, err
	}

	// proof of inclusion for epoch metadata in sealer header
	proof.ProofEpochInfo, err = k.ProveEpochInfo(epoch)
	if err != nil {
		return nil, err
	}

	// proof of inclusion for validator set in sealer header
	proof.ProofEpochValSet, err = k.ProveValSet(epoch)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// ProveEpochSubmitted generates proof that the epoch's checkpoint is submitted to BTC
// i.e., the two `TransactionInfo`s for the checkpoint
func (k Keeper) ProveEpochSubmitted(ctx context.Context, sk *btcctypes.SubmissionKey) ([]*btcctypes.TransactionInfo, error) {
	bestSubmissionData := k.btccKeeper.GetSubmissionData(ctx, *sk)
	if bestSubmissionData == nil {
		return nil, fmt.Errorf("the best submission key for epoch %d has no submission data", bestSubmissionData.Epoch)
	}
	return bestSubmissionData.TxsInfo, nil
}

// proveFinalizedBSN generates proofs that a BSN header has been finalised by the given epoch with epochInfo
// It includes proofEpochSealed and proofEpochSubmitted
// The proofs can be verified by a verifier with access to a BTC and Babylon light client
// CONTRACT: this is only a private helper function for simplifying the implementation of RPC calls
func (k Keeper) proveFinalizedBSN(
	ctx context.Context,
	indexedHeader *types.IndexedHeader,
	epochInfo *epochingtypes.Epoch,
	bestSubmissionKey *btcctypes.SubmissionKey,
) (*types.ProofFinalizedHeader, error) {
	var (
		err   error
		proof = &types.ProofFinalizedHeader{}
	)

	// proof that the epoch is sealed
	proof.ProofEpochSealed, err = k.ProveEpochSealed(ctx, epochInfo.EpochNumber)
	if err != nil {
		return nil, err
	}

	// proof that the epoch's checkpoint is submitted to BTC
	// i.e., the two `TransactionInfo`s for the checkpoint
	proof.ProofEpochSubmitted, err = k.ProveEpochSubmitted(ctx, bestSubmissionKey)
	if err != nil {
		// The only error in ProveEpochSubmitted is the nil bestSubmission.
		// Since the epoch w.r.t. the bestSubmissionKey is finalised, this
		// can only be a programming error, so we should panic here.
		panic(err)
	}

	// proof that the consumer header is included in the epoch
	proof.ProofConsumerHeaderInEpoch, err = k.ProveConsumerHeaderInEpoch(ctx, indexedHeader, epochInfo)
	if err != nil {
		return nil, err
	}

	return proof, nil
}
