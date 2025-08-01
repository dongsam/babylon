package types_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/babylonlabs-io/babylon/v3/testutil/datagen"
	"github.com/babylonlabs-io/babylon/v3/x/zoneconcierge/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	gs := datagen.GenRandomZoneconciergeGenState(r)
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
		errMsg   string
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc:     "empty is invalid",
			genState: &types.GenesisState{},
			valid:    false,
			errMsg:   "identifier cannot be blank",
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				PortId: types.PortID,
				Params: types.Params{IbcPacketTimeoutSeconds: 100},
			},
			valid: true,
		},
		{
			desc: "invalid port id",
			genState: &types.GenesisState{
				PortId: "invalid!port",
				Params: types.DefaultParams(),
			},
			valid:  false,
			errMsg: "invalid identifier",
		},
		{
			desc: "duplicate finalized header entries",
			genState: &types.GenesisState{
				PortId:           types.PortID,
				FinalizedHeaders: append(gs.FinalizedHeaders, gs.FinalizedHeaders[0]),
			},
			valid:  false,
			errMsg: "duplicate entry",
		},
		{
			desc: "duplicate consumer BTC state entries",
			genState: &types.GenesisState{
				PortId:       types.PortID,
				BsnBtcStates: append(gs.BsnBtcStates, gs.BsnBtcStates[0]),
			},
			valid:  false,
			errMsg: "duplicate entry",
		},
		{
			desc: "invalid finalized header entry (nil header)",
			genState: &types.GenesisState{
				PortId: types.PortID,
				FinalizedHeaders: []*types.FinalizedHeaderEntry{
					{ConsumerId: "consumer1", EpochNumber: 1, HeaderWithProof: nil},
				},
			},
			valid:  false,
			errMsg: "empty header with proof",
		},
		{
			desc: "invalid sealed epoch proof (nil proof)",
			genState: &types.GenesisState{
				PortId: types.PortID,
				SealedEpochsProofs: []*types.SealedEpochProofEntry{
					{EpochNumber: 1, Proof: nil},
				},
			},
			valid:  false,
			errMsg: "empty proof",
		},
		{
			desc: "invalid params",
			genState: &types.GenesisState{
				PortId: types.PortID,
				Params: types.Params{IbcPacketTimeoutSeconds: 0},
			},
			valid:  false,
			errMsg: "IbcPacketTimeoutSeconds must be positive",
		},
		{
			desc:     "valid full genesis state",
			genState: gs,
			valid:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.ErrorContains(t, err, tc.errMsg)
		})
	}
}
