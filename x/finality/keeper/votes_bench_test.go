package keeper_test

import (
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/babylonlabs-io/babylon/v3/app/signingcontext"
	"github.com/babylonlabs-io/babylon/v3/testutil/datagen"
	keepertest "github.com/babylonlabs-io/babylon/v3/testutil/keeper"
	bbn "github.com/babylonlabs-io/babylon/v3/types"
	"github.com/babylonlabs-io/babylon/v3/x/finality/keeper"
	"github.com/babylonlabs-io/babylon/v3/x/finality/types"
)

func benchmarkAddFinalitySig(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
	cKeeper := types.NewMockCheckpointingKeeper(ctrl)
	fKeeper, ctx := keepertest.FinalityKeeper(b, bsKeeper, nil, cKeeper)
	ms := keeper.NewMsgServerImpl(*fKeeper)

	// create a random finality provider
	btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(b, err)
	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	fpBTCPKBytes := fpBTCPK.MustMarshal()
	require.NoError(b, err)

	fpPopContext := signingcontext.FpPopContextV0(ctx.ChainID(), fKeeper.ModuleAddress())
	randCommitContext := signingcontext.FpRandCommitContextV0(ctx.ChainID(), fKeeper.ModuleAddress())
	finalitySigContext := signingcontext.FpPopContextV0(ctx.ChainID(), fKeeper.ModuleAddress())

	fp, err := datagen.GenRandomFinalityProviderWithBTCSK(r, btcSK, fpPopContext, "")
	require.NoError(b, err)

	// register the finality provider
	bsKeeper.EXPECT().HasFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(true).AnyTimes()
	bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(fp, nil).AnyTimes()

	// commit enough public randomness
	// TODO: generalise commit public randomness to allow arbitrary benchtime
	randListInfo, msg, err := datagen.GenRandomMsgCommitPubRandList(r, btcSK, randCommitContext, 0, 100000)
	require.NoError(b, err)
	_, err = ms.CommitPubRandList(ctx, msg)
	require.NoError(b, err)

	// Start the CPU profiler
	cpuProfileFile := "/tmp/finality-submit-finality-sig-cpu.pprof"
	f, err := os.Create(cpuProfileFile)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		b.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// Reset timer before the benchmark loop starts
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		height := uint64(i)

		// generate a vote
		blockHash := datagen.GenRandomByteArray(r, 32)
		signer := datagen.GenRandomAccount().Address
		msg, err := datagen.NewMsgAddFinalitySig(signer, btcSK, finalitySigContext, 0, height, randListInfo, blockHash)
		require.NoError(b, err)
		ctx = ctx.WithHeaderInfo(header.Info{Height: int64(height), AppHash: blockHash})

		b.StartTimer()

		_, err = ms.AddFinalitySig(ctx, msg)
		require.Error(b, err)
	}
}

func BenchmarkAddFinalitySig(b *testing.B) { benchmarkAddFinalitySig(b) }
