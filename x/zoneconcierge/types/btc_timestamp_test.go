package types_test

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/boljen/go-bitmap"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"

	txformat "github.com/babylonlabs-io/babylon/v3/btctxformatter"
	"github.com/babylonlabs-io/babylon/v3/crypto/bls12381"
	"github.com/babylonlabs-io/babylon/v3/testutil/datagen"
	testhelper "github.com/babylonlabs-io/babylon/v3/testutil/helper"
	btcctypes "github.com/babylonlabs-io/babylon/v3/x/btccheckpoint/types"
	btcstkconsumertypes "github.com/babylonlabs-io/babylon/v3/x/btcstkconsumer/types"
	checkpointingtypes "github.com/babylonlabs-io/babylon/v3/x/checkpointing/types"
	"github.com/babylonlabs-io/babylon/v3/x/zoneconcierge/types"
)

func signBLSWithBitmap(blsSKs []bls12381.PrivateKey, bm bitmap.Bitmap, msg []byte) (bls12381.Signature, error) {
	sigs := []bls12381.Signature{}
	for i := 0; i < len(blsSKs); i++ {
		if bitmap.Get(bm, i) {
			sig := bls12381.Sign(blsSKs[i], msg)
			sigs = append(sigs, sig)
		}
	}
	return bls12381.AggrSigList(sigs)
}

func FuzzBTCTimestamp(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// generate the validator set with 10 validators as genesis
		genesisValSet, privSigner, err := datagen.GenesisValidatorSetWithPrivSigner(10)
		require.NoError(t, err)
		h := testhelper.NewHelperWithValSet(t, genesisValSet, privSigner)
		ek := &h.App.EpochingKeeper
		zck := h.App.ZoneConciergeKeeper

		// empty BTC timestamp
		btcTs := &types.BTCTimestamp{}
		btcTs.Proof = &types.ProofFinalizedHeader{}

		// chain is at height 1 thus epoch 1

		/*
			generate Consumer header and its inclusion proof to an epoch
		*/
		// enter block 11, 1st block of epoch 2
		epochInterval := ek.GetParams(h.Ctx).EpochInterval
		for j := 0; j < int(epochInterval)-2; j++ {
			h.Ctx, err = h.ApplyEmptyBlockWithVoteExtension(r)
			h.NoError(err)
		}

		// handle a random header from a random consumer chain
		consumerID := datagen.GenRandomHexStr(r, 10)

		// Register the consumer through the btcstkconsumer keeper
		consumerRegister := &btcstkconsumertypes.ConsumerRegister{
			ConsumerId:          consumerID,
			ConsumerName:        "test-consumer",
			ConsumerDescription: "Test consumer for BTC timestamp",
			ConsumerMetadata: &btcstkconsumertypes.ConsumerRegister_CosmosConsumerMetadata{
				CosmosConsumerMetadata: &btcstkconsumertypes.CosmosConsumerMetadata{},
			},
			BabylonRewardsCommission: datagen.GenBabylonRewardsCommission(r),
		}
		err = h.App.BTCStkConsumerKeeper.RegisterConsumer(h.Ctx, consumerRegister)
		require.NoError(t, err)

		height := datagen.RandomInt(r, 100) + 1
		ibctmHeader := datagen.GenRandomIBCTMHeader(r, height)
		headerInfo := datagen.NewZCHeaderInfo(ibctmHeader, consumerID)
		zck.HandleHeaderWithValidCommit(h.Ctx, datagen.GenRandomByteArray(r, 32), headerInfo, false)

		// ensure the header is successfully inserted in latest epoch headers
		indexedHeader := zck.GetLatestEpochHeader(h.Ctx, consumerID)
		require.NotNil(t, indexedHeader)

		// enter block 21, 1st block of epoch 3
		for j := 0; j < int(epochInterval); j++ {
			h.Ctx, err = h.ApplyEmptyBlockWithVoteExtension(r)
			h.NoError(err)
		}
		// seal last epoch
		h.Ctx, err = h.ApplyEmptyBlockWithVoteExtension(r)
		h.NoError(err)

		epochWithHeader, err := ek.GetHistoricalEpoch(h.Ctx, indexedHeader.BabylonEpoch)
		h.NoError(err)

		// generate inclusion proof
		btcTs.EpochInfo = epochWithHeader
		btcTs.Header = indexedHeader
		/*
			seal the epoch and generate ProofEpochSealed
		*/
		// construct the rawCkpt
		// Note that the BlsMultiSig will be generated and assigned later
		bm := datagen.GenFullBitmap()
		blockHash := checkpointingtypes.BlockHash(epochWithHeader.SealerBlockHash)
		rawCkpt := &checkpointingtypes.RawCheckpoint{
			EpochNum:    epochWithHeader.EpochNumber,
			BlockHash:   &blockHash,
			Bitmap:      bm,
			BlsMultiSig: nil,
		}
		// let the subset generate a BLS multisig over sealer header's app_hash
		multiSig, err := signBLSWithBitmap(h.GenValidators.GetBLSPrivKeys(), bm, rawCkpt.SignedMsg())
		require.NoError(t, err)
		// assign multiSig to rawCkpt
		rawCkpt.BlsMultiSig = &multiSig

		// prove
		btcTs.Proof.ProofEpochSealed, err = zck.ProveEpochSealed(h.Ctx, epochWithHeader.EpochNumber)
		require.NoError(t, err)

		btcTs.RawCheckpoint = rawCkpt

		/*
			forge two BTC headers including the checkpoint
		*/
		// encode ckpt to BTC txs in BTC blocks
		submitterAddr := datagen.GenRandomByteArray(r, txformat.AddressLength)
		rawBTCCkpt, err := checkpointingtypes.FromRawCkptToBTCCkpt(rawCkpt, submitterAddr)
		h.NoError(err)
		testRawCkptData := datagen.EncodeRawCkptToTestData(rawBTCCkpt)
		idxs := []uint64{datagen.RandomInt(r, 5) + 1, datagen.RandomInt(r, 5) + 1}
		offsets := []uint64{datagen.RandomInt(r, 5) + 1, datagen.RandomInt(r, 5) + 1}
		btcBlocks := []*datagen.BlockCreationResult{
			datagen.CreateBlock(r, 1, uint32(idxs[0]+offsets[0]), uint32(idxs[0]), testRawCkptData.FirstPart),
			datagen.CreateBlock(r, 2, uint32(idxs[1]+offsets[1]), uint32(idxs[1]), testRawCkptData.SecondPart),
		}
		// create MsgInsertBtcSpvProof for the rawCkpt
		msgInsertBtcSpvProof := datagen.GenerateMessageWithRandomSubmitter([]*datagen.BlockCreationResult{btcBlocks[0], btcBlocks[1]})

		// assign BTC submission key and ProofEpochSubmitted
		btcTs.BtcSubmissionKey = &btcctypes.SubmissionKey{
			Key: []*btcctypes.TransactionKey{
				&btcctypes.TransactionKey{Index: uint32(idxs[0]), Hash: btcBlocks[0].HeaderBytes.Hash()},
				&btcctypes.TransactionKey{Index: uint32(idxs[1]), Hash: btcBlocks[1].HeaderBytes.Hash()},
			},
		}
		btcTs.Proof.ProofEpochSubmitted = []*btcctypes.TransactionInfo{
			{
				Key:         btcTs.BtcSubmissionKey.Key[0],
				Transaction: msgInsertBtcSpvProof.Proofs[0].BtcTransaction,
				Proof:       msgInsertBtcSpvProof.Proofs[0].MerkleNodes,
			},
			{
				Key:         btcTs.BtcSubmissionKey.Key[1],
				Transaction: msgInsertBtcSpvProof.Proofs[1].BtcTransaction,
				Proof:       msgInsertBtcSpvProof.Proofs[1].MerkleNodes,
			},
		}

		// get headers for verification
		btcHeaders := []*wire.BlockHeader{
			btcBlocks[0].HeaderBytes.ToBlockHeader(),
			btcBlocks[1].HeaderBytes.ToBlockHeader(),
		}

		// net param, babylonTag
		powLimit := chaincfg.SimNetParams.PowLimit
		babylonTag := btcctypes.DefaultCheckpointTag
		tagAsBytes, _ := hex.DecodeString(babylonTag)

		err = btcTs.VerifyStateless(btcHeaders, powLimit, tagAsBytes)
		h.NoError(err)
	})
}
