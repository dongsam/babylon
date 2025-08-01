package keeper

import (
	"context"
)

// HasConsumer checks if a consumer is registered and is a Cosmos consumer
func (k Keeper) HasConsumer(ctx context.Context, consumerID string) bool {
	if !k.btcStkKeeper.IsConsumerRegistered(ctx, consumerID) {
		return false
	}

	isCosmosConsumer, err := k.btcStkKeeper.IsCosmosConsumer(ctx, consumerID)
	if err != nil {
		return false
	}

	return isCosmosConsumer
}

// GetAllConsumerIDs returns all registered Cosmos consumer IDs
func (k Keeper) GetAllConsumerIDs(ctx context.Context) []string {
	allConsumerIDs := k.btcStkKeeper.GetAllRegisteredConsumerIDs(ctx)

	var cosmosConsumerIDs []string
	for _, consumerID := range allConsumerIDs {
		isCosmosConsumer, err := k.btcStkKeeper.IsCosmosConsumer(ctx, consumerID)
		if err != nil {
			continue
		}
		if isCosmosConsumer {
			cosmosConsumerIDs = append(cosmosConsumerIDs, consumerID)
		}
	}

	return cosmosConsumerIDs
}
