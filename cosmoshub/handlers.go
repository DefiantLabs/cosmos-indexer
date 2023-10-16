package cosmoshub

import tendermintHandlers "github.com/DefiantLabs/cosmos-indexer/tendermint"

// EndBlockerEventTypeHandlers should be extended using these and an init func to set up CosmosHub end blocker handlers if we want more functionality.
var EndBlockerEventTypeHandlers = tendermintHandlers.EndBlockerEventTypeHandlers
