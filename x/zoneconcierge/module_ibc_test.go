package zoneconcierge_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	"github.com/babylonlabs-io/babylon/v3/app"
	bsctypes "github.com/babylonlabs-io/babylon/v3/x/btcstkconsumer/types"
	zctypes "github.com/babylonlabs-io/babylon/v3/x/zoneconcierge/types"
)

//func init() {
//	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
//		return setupTestingApp()
//	}
//}

//func setupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
//	// Create the app using the same approach as the working version
//	// but return it without pre-initialization to let IBC testing framework handle it
//	babylonApp, genesisState, err := app.NewBabylonAppForIBCTesting(false)
//	if err != nil {
//		panic(err)
//	}
//
//	return babylonApp, genesisState
//}

type IBCChannelCreationTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator
	chainA      *ibctesting.TestChain // Babylon chain
	chainB      *ibctesting.TestChain // Consumer chain
	path        *ibctesting.Path
}

func TestIBCChannelCreationTestSuite(t *testing.T) {
	suite.Run(t, new(IBCChannelCreationTestSuite))
}

func (suite *IBCChannelCreationTestSuite) SetupTest() {
	//suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1)
	//suite.coordinator.Chains = make(map[string]*ibctesting.TestChain)

	tb := suite.T()
	//tb.Helper()

	chains := make(map[string]*ibctesting.TestChain)
	coord := &ibctesting.Coordinator{
		T:           tb,
		CurrentTime: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	coord.Chains = chains
	suite.coordinator = coord

	app.NewBabylonAppForIBCTesting(tb, false, suite.coordinator, 1)
	app.NewBabylonAppForIBCTesting(tb, false, suite.coordinator, 2)

	//// -----
	////chainId := ibctesting.GetChainID(1)
	//_, _, testChain, err := app.NewBabylonAppForIBCTesting(false)
	//if err != nil {
	//	panic(err)
	//}
	//// create current header and call begin block
	//header := cmtproto.Header{
	//	ChainID: chainId,
	//	Height:  1,
	//	Time:    suite.coordinator.CurrentTime.UTC(),
	//}
	//testChain.TB = tb
	//testChain.Coordinator = suite.coordinator
	//testChain.ChainID = chainId
	//testChain.ProposedHeader = header
	//suite.coordinator.Chains[chainId] = testChain
	//
	//// -----
	//chainId2 := ibctesting.GetChainID(2)
	//_, _, testChain2, err := app.NewBabylonAppForIBCTesting(false)
	//if err != nil {
	//	panic(err)
	//}
	//// create current header and call begin block
	//header2 := cmtproto.Header{
	//	ChainID: chainId2,
	//	Height:  1,
	//	Time:    suite.coordinator.CurrentTime.UTC(),
	//}
	//testChain2.TB = tb
	//testChain2.Coordinator = suite.coordinator
	//testChain2.ChainID = chainId2
	//testChain2.ProposedHeader = header2
	//suite.coordinator.Chains[chainId2] = testChain2

	//suite.coordinator = ibctesting.NewCustomAppCoordinator(suite.T(), 1, setupTestingApp)

	// Setup Babylon chain (chainA) and Consumer chain (chainB)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	// Create IBC path between the chains
	suite.path = ibctesting.NewPath(suite.chainA, suite.chainB)
	suite.path.EndpointA.ChannelConfig.PortID = zctypes.PortID
	suite.path.EndpointB.ChannelConfig.PortID = ibctesting.MockPort
	suite.path.EndpointA.ChannelConfig.Version = zctypes.Version
	suite.path.EndpointB.ChannelConfig.Version = zctypes.Version
	suite.path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	suite.path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
}

// getBabylonApp returns the BabylonApp instance
func (suite *IBCChannelCreationTestSuite) getBabylonApp() *app.BabylonApp {
	return suite.chainA.App.(*app.BabylonApp)
}

// registerConsumer registers a consumer in the btcstkconsumer module
func (suite *IBCChannelCreationTestSuite) registerConsumer(consumerID, consumerName, consumerDescription string) {
	babylonApp := suite.getBabylonApp()

	// Create consumer register
	consumerRegister := bsctypes.NewCosmosConsumerRegister(
		consumerID,
		consumerName,
		consumerDescription,
		math.LegacyNewDecWithPrec(1, 2), // 1% commission
	)

	ctx := suite.chainA.GetContext()
	err := babylonApp.BTCStkConsumerKeeper.RegisterConsumer(ctx, consumerRegister)
	suite.Require().NoError(err)
}

func (suite *IBCChannelCreationTestSuite) TestSuccessfulChannelCreation() {
	// Setup: Register consumer first
	clientID := suite.path.EndpointA.ClientID
	suite.registerConsumer(clientID, "test-consumer", "Test Consumer Chain")

	// Create connection
	suite.coordinator.SetupConnections(suite.path)

	// Test: Initialize channel creation
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().NoError(err)

	// Verify the channel was created successfully
	babylonApp := suite.getBabylonApp()
	ctx := suite.chainA.GetContext()

	// Check that consumer register was updated with channel ID
	consumerRegister, err := babylonApp.BTCStkConsumerKeeper.GetConsumerRegister(ctx, clientID)
	suite.Require().NoError(err)
	suite.Require().NotNil(consumerRegister.GetCosmosConsumerMetadata())
	suite.Require().Equal(suite.path.EndpointA.ChannelID, consumerRegister.GetCosmosConsumerMetadata().ChannelId)
}

func (suite *IBCChannelCreationTestSuite) TestDuplicateChannelCreationFails() {
	// Setup: Register consumer and create first channel
	clientID := suite.path.EndpointA.ClientID
	suite.registerConsumer(clientID, "test-consumer", "Test Consumer Chain")

	// Create connection
	suite.coordinator.SetupConnections(suite.path)

	// Create first channel successfully
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().NoError(err)

	// Complete the handshake to make channel OPEN
	err = suite.path.EndpointB.ChanOpenTry()
	suite.Require().NoError(err)
	err = suite.path.EndpointA.ChanOpenAck()
	suite.Require().NoError(err)
	err = suite.path.EndpointB.ChanOpenConfirm()
	suite.Require().NoError(err)

	// Attempt to create second channel - should fail
	secondPath := ibctesting.NewPath(suite.chainA, suite.chainB)
	secondPath.EndpointA.ChannelConfig.PortID = zctypes.PortID
	secondPath.EndpointB.ChannelConfig.PortID = ibctesting.MockPort
	secondPath.EndpointA.ChannelConfig.Version = zctypes.Version
	secondPath.EndpointB.ChannelConfig.Version = zctypes.Version
	secondPath.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	secondPath.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	// Use same connection
	secondPath.EndpointA.ConnectionID = suite.path.EndpointA.ConnectionID
	secondPath.EndpointB.ConnectionID = suite.path.EndpointB.ConnectionID

	// This should fail due to existing open channel
	err = secondPath.EndpointA.ChanOpenInit()
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "already has an open IBC channel")
}

func (suite *IBCChannelCreationTestSuite) TestClosedChannelReplacement() {
	// Setup: Register consumer and create first channel
	clientID := suite.path.EndpointA.ClientID
	suite.registerConsumer(clientID, "test-consumer", "Test Consumer Chain")

	// Create connection
	suite.coordinator.SetupConnections(suite.path)

	// Create and open first channel
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().NoError(err)
	err = suite.path.EndpointB.ChanOpenTry()
	suite.Require().NoError(err)
	err = suite.path.EndpointA.ChanOpenAck()
	suite.Require().NoError(err)
	err = suite.path.EndpointB.ChanOpenConfirm()
	suite.Require().NoError(err)

	// Simulate channel timeout by manually closing the channel
	babylonApp := suite.getBabylonApp()
	ctx := suite.chainA.GetContext()

	// Manually set channel state to CLOSED to simulate timeout
	channel, found := babylonApp.IBCKeeper.ChannelKeeper.GetChannel(ctx, zctypes.PortID, suite.path.EndpointA.ChannelID)
	suite.Require().True(found)
	channel.State = channeltypes.CLOSED
	babylonApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, zctypes.PortID, suite.path.EndpointA.ChannelID, channel)

	// Create new channel - should succeed as replacement
	newPath := ibctesting.NewPath(suite.chainA, suite.chainB)
	newPath.EndpointA.ChannelConfig.PortID = zctypes.PortID
	newPath.EndpointB.ChannelConfig.PortID = ibctesting.MockPort
	newPath.EndpointA.ChannelConfig.Version = zctypes.Version
	newPath.EndpointB.ChannelConfig.Version = zctypes.Version
	newPath.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	newPath.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	// Use same connection
	newPath.EndpointA.ConnectionID = suite.path.EndpointA.ConnectionID
	newPath.EndpointB.ConnectionID = suite.path.EndpointB.ConnectionID

	// This should succeed as replacement for closed channel
	err = newPath.EndpointA.ChanOpenInit()
	suite.Require().NoError(err)

	// Verify consumer register was updated with new channel ID
	ctx = suite.chainA.GetContext()
	consumerRegister, err := babylonApp.BTCStkConsumerKeeper.GetConsumerRegister(ctx, clientID)
	suite.Require().NoError(err)
	suite.Require().Equal(newPath.EndpointA.ChannelID, consumerRegister.GetCosmosConsumerMetadata().ChannelId)
}

func (suite *IBCChannelCreationTestSuite) TestUnregisteredConsumerFails() {
	// Create connection without registering consumer
	suite.coordinator.SetupConnections(suite.path)

	// Attempt to create channel - should fail
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "is not registered as a consumer")
}

func (suite *IBCChannelCreationTestSuite) TestInvalidChannelOrdering() {
	// Setup: Register consumer
	clientID := suite.path.EndpointA.ClientID
	suite.registerConsumer(clientID, "test-consumer", "Test Consumer Chain")

	// Create connection
	suite.coordinator.SetupConnections(suite.path)

	// Set invalid channel ordering (UNORDERED instead of ORDERED)
	suite.path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	suite.path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED

	// Attempt to create channel - should fail
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "expected ORDERED channel")
}

func (suite *IBCChannelCreationTestSuite) TestInvalidVersion() {
	// Setup: Register consumer
	clientID := suite.path.EndpointA.ClientID
	suite.registerConsumer(clientID, "test-consumer", "Test Consumer Chain")

	// Create connection
	suite.coordinator.SetupConnections(suite.path)

	// Set invalid version
	suite.path.EndpointA.ChannelConfig.Version = "invalid-version"

	// Attempt to create channel - should fail
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid version")
}

func (suite *IBCChannelCreationTestSuite) TestInvalidPortID() {
	// Setup: Register consumer
	clientID := suite.path.EndpointA.ClientID
	suite.registerConsumer(clientID, "test-consumer", "Test Consumer Chain")

	// Create connection
	suite.coordinator.SetupConnections(suite.path)

	// Set invalid port ID
	suite.path.EndpointA.ChannelConfig.PortID = "invalid-port"

	// Attempt to create channel - should fail
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid port")
}

func (suite *IBCChannelCreationTestSuite) TestRollupConsumerFails() {
	// Setup: Register rollup consumer (not Cosmos consumer)
	clientID := suite.path.EndpointA.ClientID
	babylonApp := suite.getBabylonApp()
	ctx := suite.chainA.GetContext()

	rollupConsumer := bsctypes.NewRollupConsumerRegister(
		clientID,
		"rollup-consumer",
		"Rollup Consumer Chain",
		"0x1234567890123456789012345678901234567890", // contract address
		math.LegacyNewDecWithPrec(1, 2),              // 1% commission
	)

	err := babylonApp.BTCStkConsumerKeeper.RegisterConsumer(ctx, rollupConsumer)
	suite.Require().NoError(err)

	// Create connection
	suite.coordinator.SetupConnections(suite.path)

	// Attempt to create channel - should fail for rollup consumer
	err = suite.path.EndpointA.ChanOpenInit()
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "is not a Cosmos consumer")
}

func (suite *IBCChannelCreationTestSuite) TestPacketTimeoutChannelClosure() {
	// Setup: Register consumer and create channel
	clientID := suite.path.EndpointA.ClientID
	suite.registerConsumer(clientID, "test-consumer", "Test Consumer Chain")

	// Complete handshake
	suite.coordinator.Setup(suite.path)

	// Send a packet that will timeout
	babylonApp := suite.getBabylonApp()
	ctx := suite.chainA.GetContext()

	// Send packet using the correct signature
	_, err := babylonApp.IBCKeeper.ChannelKeeper.SendPacket(ctx,
		zctypes.PortID,
		suite.path.EndpointA.ChannelID,
		clienttypes.NewHeight(0, uint64(ctx.BlockHeight())+1),
		uint64(ctx.BlockTime().Add(time.Minute).UnixNano()),
		[]byte("test packet data"))
	suite.Require().NoError(err)

	// Advance time and height to cause timeout
	suite.chainA.NextBlock()
	suite.chainA.NextBlock()
	suite.chainA.NextBlock()

	// The packet should timeout, potentially closing the channel
	// This is a simulation of real timeout behavior

	// Now verify that a new channel can be created to replace the timed-out one
	newPath := ibctesting.NewPath(suite.chainA, suite.chainB)
	newPath.EndpointA.ChannelConfig.PortID = zctypes.PortID
	newPath.EndpointB.ChannelConfig.PortID = ibctesting.MockPort
	newPath.EndpointA.ChannelConfig.Version = zctypes.Version
	newPath.EndpointB.ChannelConfig.Version = zctypes.Version
	newPath.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	newPath.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	// Use same connection
	newPath.EndpointA.ConnectionID = suite.path.EndpointA.ConnectionID
	newPath.EndpointB.ConnectionID = suite.path.EndpointB.ConnectionID

	// Manually close the old channel to simulate timeout closure
	ctx = suite.chainA.GetContext()
	channel, found := babylonApp.IBCKeeper.ChannelKeeper.GetChannel(ctx, zctypes.PortID, suite.path.EndpointA.ChannelID)
	suite.Require().True(found)
	channel.State = channeltypes.CLOSED
	babylonApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, zctypes.PortID, suite.path.EndpointA.ChannelID, channel)

	// New channel creation should succeed
	err = newPath.EndpointA.ChanOpenInit()
	suite.Require().NoError(err)
}

func (suite *IBCChannelCreationTestSuite) TestMultipleConsumersWithDifferentClients() {
	// Setup: Create second connection with different client
	path2 := ibctesting.NewPath(suite.chainA, suite.chainB)

	// Setup connections (this will create different client IDs)
	suite.coordinator.SetupConnections(suite.path)
	suite.coordinator.SetupConnections(path2)

	clientID1 := suite.path.EndpointA.ClientID
	clientID2 := path2.EndpointA.ClientID

	// Ensure we have different client IDs
	suite.Require().NotEqual(clientID1, clientID2)

	// Register both consumers
	suite.registerConsumer(clientID1, "consumer-1", "First Consumer Chain")
	suite.registerConsumer(clientID2, "consumer-2", "Second Consumer Chain")

	// Configure channels
	suite.path.EndpointA.ChannelConfig.PortID = zctypes.PortID
	suite.path.EndpointB.ChannelConfig.PortID = ibctesting.MockPort
	suite.path.EndpointA.ChannelConfig.Version = zctypes.Version
	suite.path.EndpointB.ChannelConfig.Version = zctypes.Version
	suite.path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	suite.path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	path2.EndpointA.ChannelConfig.PortID = zctypes.PortID
	path2.EndpointB.ChannelConfig.PortID = ibctesting.MockPort
	path2.EndpointA.ChannelConfig.Version = zctypes.Version
	path2.EndpointB.ChannelConfig.Version = zctypes.Version
	path2.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path2.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	// Both channel creations should succeed
	err := suite.path.EndpointA.ChanOpenInit()
	suite.Require().NoError(err)

	err = path2.EndpointA.ChanOpenInit()
	suite.Require().NoError(err)

	// Verify both consumers were updated with their respective channel IDs
	babylonApp := suite.getBabylonApp()
	ctx := suite.chainA.GetContext()

	consumer1, err := babylonApp.BTCStkConsumerKeeper.GetConsumerRegister(ctx, clientID1)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.path.EndpointA.ChannelID, consumer1.GetCosmosConsumerMetadata().ChannelId)

	consumer2, err := babylonApp.BTCStkConsumerKeeper.GetConsumerRegister(ctx, clientID2)
	suite.Require().NoError(err)
	suite.Require().Equal(path2.EndpointA.ChannelID, consumer2.GetCosmosConsumerMetadata().ChannelId)
}
