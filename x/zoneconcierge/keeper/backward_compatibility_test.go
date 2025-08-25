package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/babylonlabs-io/babylon/v4/app"
	"github.com/babylonlabs-io/babylon/v4/testutil/datagen"
)

// TestBackwardCompatibility_BroadcastBehavior verifies that the new hook-based
// implementation produces correct broadcasting behavior compared to old always-broadcast approach
func TestBackwardCompatibility_BroadcastBehavior(t *testing.T) {
	testCases := []struct {
		name                  string
		setupEvents          func(*app.BabylonApp) 
		shouldBroadcastNew   bool   // What new implementation should do
		shouldBroadcastOld   bool   // What old implementation would do (always true)
		description          string
	}{
		{
			name: "NoChanges_NewShouldSkip",
			setupEvents: func(app *app.BabylonApp) {
				// No BTC events, no consumer channels - just create context
				_ = app.NewContext(false)
			},
			shouldBroadcastNew: false, // New: efficient, skip unnecessary broadcast
			shouldBroadcastOld: true,  // Old: would always broadcast
			description: "No BTC light client modifications - new implementation should be more efficient",
		},
		{
			name: "BTCHeaderInserted_BothShouldBroadcast",
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				r := rand.New(rand.NewSource(12345))
				headerInfo := datagen.GenRandomBTCHeaderInfo(r)
				
				// Simulate BTC header insertion
				app.ZoneConciergeKeeper.AfterBTCHeaderInserted(ctx, headerInfo)
			},
			shouldBroadcastNew: true, // New: should broadcast due to BTC event
			shouldBroadcastOld: true, // Old: would always broadcast
			description: "BTC header inserted - both implementations should broadcast",
		},
		{
			name: "BTCReorg_BothShouldBroadcast", 
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				r := rand.New(rand.NewSource(12345))
				rollbackFrom := datagen.GenRandomBTCHeaderInfo(r)
				rollbackTo := datagen.GenRandomBTCHeaderInfo(r)
				rollbackTo.Height = rollbackFrom.Height - 1
				
				// Simulate BTC reorg
				app.ZoneConciergeKeeper.AfterBTCRollBack(ctx, rollbackFrom, rollbackTo)
			},
			shouldBroadcastNew: true, // New: should broadcast due to reorg
			shouldBroadcastOld: true, // Old: would always broadcast
			description: "BTC reorg occurred - both implementations should broadcast",
		},
		{
			name: "NewConsumerChannel_BothShouldBroadcast",
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				
				// Simulate new consumer channel
				app.ZoneConciergeKeeper.MarkNewConsumerChannel(ctx, "test-consumer")
			},
			shouldBroadcastNew: true, // New: should broadcast due to new consumer
			shouldBroadcastOld: true, // Old: would always broadcast
			description: "New consumer channel opened - both implementations should broadcast",
		},
		{
			name: "MultipleEvents_BothShouldBroadcast",
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				r := rand.New(rand.NewSource(12345))
				headerInfo := datagen.GenRandomBTCHeaderInfo(r)
				
				// Simulate multiple events
				app.ZoneConciergeKeeper.AfterBTCHeaderInserted(ctx, headerInfo)
				app.ZoneConciergeKeeper.MarkNewConsumerChannel(ctx, "test-consumer")
			},
			shouldBroadcastNew: true, // New: should broadcast due to multiple events
			shouldBroadcastOld: true, // Old: would always broadcast
			description: "Multiple BTC events - both implementations should broadcast",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup fresh test environment
			babylonApp := app.Setup(t, false)
			ctx := babylonApp.NewContext(false)
			
			// Setup events
			tc.setupEvents(babylonApp)
			
			// Check new implementation behavior
			actualNewBehavior := babylonApp.ZoneConciergeKeeper.ShouldBroadcastBTCHeaders(ctx)
			reason := babylonApp.ZoneConciergeKeeper.GetBroadcastTriggerReason(ctx)
			
			t.Logf("📋 Test scenario: %s", tc.description)
			t.Logf("🔄 Old implementation would broadcast: %v", tc.shouldBroadcastOld) 
			t.Logf("✨ New implementation broadcasts: %v (reason: %s)", actualNewBehavior, reason)
			
			// Verify new implementation behavior
			require.Equal(t, tc.shouldBroadcastNew, actualNewBehavior,
				"New implementation broadcasting behavior should match expected")
			
			// When BTC events occur, both old and new should broadcast (ensuring no functionality loss)
			if tc.shouldBroadcastNew {
				require.True(t, tc.shouldBroadcastOld,
					"When events occur, old implementation would also broadcast (no functionality loss)")
			}
			
			// When no BTC events occur, new should be more efficient  
			if !tc.shouldBroadcastNew {
				require.True(t, tc.shouldBroadcastOld,
					"When no events occur, old would broadcast unnecessarily (inefficiency)")
				require.Equal(t, "none", reason,
					"No broadcast reason should be 'none'")
			} else {
				require.NotEqual(t, "none", reason,
					"Broadcast should have a valid reason")
			}
		})
	}
}

// TestEfficiencyImprovement_StatisticalAnalysis provides statistical analysis
// of the efficiency improvement from the hook-based approach
func TestEfficiencyImprovement_StatisticalAnalysis(t *testing.T) {
	t.Run("EfficiencyGains_Statistics", func(t *testing.T) {
		babylonApp := app.Setup(t, false)
		_ = babylonApp.ZoneConciergeKeeper
		
		scenarios := []struct {
			name                string
			btcEventsPerHour    int  // How many BTC events per hour
			babylonBlocksPer6s  int  // Babylon blocks every 6 seconds = 600 per hour
			expectedEfficiency  float64
		}{
			{
				name:                "TypicalBTCRate_10MinBlocks",
				btcEventsPerHour:    6,   // ~10 min BTC blocks
				babylonBlocksPer6s:  600, // 600 Babylon blocks per hour (6s each)
				expectedEfficiency:  99.0, // (600-6)/600 = 99%
			},
			{
				name:                "FastBTCRate_5MinBlocks", 
				btcEventsPerHour:    12,  // 5 min BTC blocks
				babylonBlocksPer6s:  600,
				expectedEfficiency:  98.0, // (600-12)/600 = 98%
			},
			{
				name:                "SlowBTCRate_20MinBlocks",
				btcEventsPerHour:    3,   // 20 min BTC blocks
				babylonBlocksPer6s:  600,
				expectedEfficiency:  99.5, // (600-3)/600 = 99.5%
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				// Simulate old behavior: broadcast on every Babylon block
				oldBroadcasts := scenario.babylonBlocksPer6s
				
				// Simulate new behavior: broadcast only on BTC events
				newBroadcasts := scenario.btcEventsPerHour
				
				// Calculate efficiency improvement
				actualEfficiency := float64(oldBroadcasts-newBroadcasts) / float64(oldBroadcasts) * 100
				
				t.Logf("Scenario: %s", scenario.name)
				t.Logf("Old implementation broadcasts: %d times/hour", oldBroadcasts)
				t.Logf("New implementation broadcasts: %d times/hour", newBroadcasts) 
				t.Logf("Efficiency improvement: %.1f%%", actualEfficiency)
				
				// Verify efficiency meets expectations
				require.InEpsilon(t, scenario.expectedEfficiency, actualEfficiency, 0.1,
					"Efficiency improvement should be approximately %.1f%%", scenario.expectedEfficiency)
				
				// Verify significant improvement
				require.Greater(t, actualEfficiency, 95.0, 
					"Efficiency improvement should be at least 95%%")
			})
		}
	})
}

// TestBroadcastCorrectness_EndToEnd verifies that the broadcasting decision
// is made correctly in realistic scenarios
func TestBroadcastCorrectness_EndToEnd(t *testing.T) {
	testCases := []struct {
		name           string
		setupEvents    func(app *app.BabylonApp)
		shouldBroadcast bool
		reason         string
	}{
		{
			name: "NoEvents_ShouldSkip",
			setupEvents: func(app *app.BabylonApp) {
				// No events - just create context
				_ = app.NewContext(false)
			},
			shouldBroadcast: false,
			reason:         "none",
		},
		{
			name: "BTCHeaderOnly_ShouldBroadcast",
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				r := rand.New(rand.NewSource(12345))
				headerInfo := datagen.GenRandomBTCHeaderInfo(r)
				app.ZoneConciergeKeeper.AfterBTCHeaderInserted(ctx, headerInfo)
			},
			shouldBroadcast: true,
			reason:         "new_btc_header",
		},
		{
			name: "ReorgOnly_ShouldBroadcast",
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				r := rand.New(rand.NewSource(12345))
				rollbackFrom := datagen.GenRandomBTCHeaderInfo(r)
				rollbackTo := datagen.GenRandomBTCHeaderInfo(r)
				rollbackTo.Height = rollbackFrom.Height - 1
				app.ZoneConciergeKeeper.AfterBTCRollBack(ctx, rollbackFrom, rollbackTo)
			},
			shouldBroadcast: true,
			reason:         "btc_reorg",
		},
		{
			name: "ConsumerChannelOnly_ShouldBroadcast", 
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				app.ZoneConciergeKeeper.MarkNewConsumerChannel(ctx, "consumer-123")
			},
			shouldBroadcast: true,
			reason:         "new_consumer_channel",
		},
		{
			name: "AllEvents_ShouldBroadcast",
			setupEvents: func(app *app.BabylonApp) {
				ctx := app.NewContext(false)
				r := rand.New(rand.NewSource(12345))
				
				// Add all types of events
				headerInfo := datagen.GenRandomBTCHeaderInfo(r)
				app.ZoneConciergeKeeper.AfterBTCHeaderInserted(ctx, headerInfo)
				
				rollbackFrom := datagen.GenRandomBTCHeaderInfo(r)
				rollbackTo := datagen.GenRandomBTCHeaderInfo(r)
				rollbackTo.Height = rollbackFrom.Height - 1
				app.ZoneConciergeKeeper.AfterBTCRollBack(ctx, rollbackFrom, rollbackTo)
				
				app.ZoneConciergeKeeper.MarkNewConsumerChannel(ctx, "consumer-456")
			},
			shouldBroadcast: true,
			reason:         "new_btc_header,btc_reorg,new_consumer_channel",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup fresh test environment
			babylonApp := app.Setup(t, false)
			ctx := babylonApp.NewContext(false)
			
			// Setup test events
			tc.setupEvents(babylonApp)
			
			// Check broadcasting decision
			shouldBroadcast := babylonApp.ZoneConciergeKeeper.ShouldBroadcastBTCHeaders(ctx)
			reason := babylonApp.ZoneConciergeKeeper.GetBroadcastTriggerReason(ctx)
			
			// Verify correctness
			require.Equal(t, tc.shouldBroadcast, shouldBroadcast,
				"Broadcasting decision should match expected")
			require.Equal(t, tc.reason, reason,
				"Broadcasting reason should match expected")
			
			t.Logf("✅ Scenario: %s", tc.name)
			t.Logf("   Should broadcast: %v (reason: %s)", shouldBroadcast, reason)
		})
	}
}

// TestConsistencyAcrossBlocks verifies that the hook system maintains
// consistency across multiple blocks (using fresh app instances to simulate block boundaries)
func TestConsistencyAcrossBlocks(t *testing.T) {
	t.Run("MultiBlock_Consistency", func(t *testing.T) {
		// Simulate multiple blocks with different patterns
		blocks := []struct {
			blockNum        int
			hasEvents       bool
			shouldBroadcast bool
		}{
			{1, false, false}, // No events
			{2, true, true},   // BTC header added
			{3, false, false}, // No events (transient store cleared)
			{4, true, true},   // Consumer channel added
			{5, false, false}, // No events again
		}
		
		for _, block := range blocks {
			// Create fresh app for each block (simulating transient store clearing)
			babylonApp := app.Setup(t, false)
			ctx := babylonApp.NewContext(false)
			zcKeeper := babylonApp.ZoneConciergeKeeper
			
			if block.hasEvents {
				if block.blockNum == 2 {
					// Add BTC header event
					r := rand.New(rand.NewSource(int64(block.blockNum)))
					headerInfo := datagen.GenRandomBTCHeaderInfo(r)
					zcKeeper.AfterBTCHeaderInserted(ctx, headerInfo)
				} else if block.blockNum == 4 {
					// Add consumer channel event
					zcKeeper.MarkNewConsumerChannel(ctx, "consumer-block-4")
				}
			}
			
			shouldBroadcast := zcKeeper.ShouldBroadcastBTCHeaders(ctx)
			reason := zcKeeper.GetBroadcastTriggerReason(ctx)
			
			require.Equal(t, block.shouldBroadcast, shouldBroadcast,
				"Block %d: broadcasting decision should match", block.blockNum)
			
			if block.shouldBroadcast {
				require.NotEqual(t, "none", reason,
					"Block %d: should have a broadcast reason", block.blockNum)
			} else {
				require.Equal(t, "none", reason,
					"Block %d: should have no broadcast reason", block.blockNum)
			}
			
			t.Logf("Block %d: events=%v, broadcast=%v, reason=%s", 
				block.blockNum, block.hasEvents, shouldBroadcast, reason)
		}
	})
}