package store

// BlockCompositeRepo combines BlockStore (block CRUD + device/rack helpers)
// with TierAggregationStore (aggregation + port connection operations)
// to satisfy the service.BlockRepository interface.
type BlockCompositeRepo struct {
	*BlockStore
	*TierAggregationStore
}
