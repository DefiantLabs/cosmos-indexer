package indexer

import (
	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/core"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/filter"
	"github.com/DefiantLabs/cosmos-indexer/parsers"
	"github.com/DefiantLabs/probe/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	"gorm.io/gorm"
)

// DB is not safe to add here just yet, since the index command in cmd/ defers a close of the DB connection
// Maybe the defer should be removed?
type PostSetupDataset struct {
	Config      *config.IndexConfig
	DryRun      bool
	ChainClient *client.ChainClient
}

type PostSetupCustomDataset struct {
	Config      config.IndexConfig
	ChainClient *client.ChainClient
	DB          *gorm.DB
}

type PostIndexCustomMessageDataset struct {
	Config         config.IndexConfig
	DB             *gorm.DB
	DryRun         bool
	IndexedDataset *[]dbTypes.TxDBWrapper
	MessageParser  map[string]models.MessageParser
	IndexedBlock   models.Block
}

type PreExitCustomDataset struct {
	Config config.IndexConfig
	DB     *gorm.DB
	DryRun bool
}

type Indexer struct {
	Config                              *config.IndexConfig
	DryRun                              bool
	DB                                  *gorm.DB
	ChainClient                         *client.ChainClient
	BlockEnqueueFunction                func(chan *core.EnqueueData) error
	CustomModuleBasics                  []module.AppModuleBasic // Used for extending the AppModuleBasics registered in the probe ChainClientient
	BlockEventFilterRegistries          BlockEventFilterRegistries
	MessageTypeFilters                  []filter.MessageTypeFilter
	CustomBeginBlockEventParserRegistry map[string][]parsers.BlockEventParser // Used for associating parsers to block event types in BeginBlock events
	CustomEndBlockEventParserRegistry   map[string][]parsers.BlockEventParser // Used for associating parsers to block event types in EndBlock events
	CustomBeginBlockParserTrackers      map[string]models.BlockEventParser    // Used for tracking block event parsers in the database
	CustomEndBlockParserTrackers        map[string]models.BlockEventParser    // Used for tracking block event parsers in the database
	CustomMessageParserRegistry         map[string][]parsers.MessageParser    // Used for associating parsers to message types
	CustomMessageParserTrackers         map[string]models.MessageParser       // Used for tracking message parsers in the database
	CustomModels                        []any
	PostIndexCustomMessageFunction      func(*PostIndexCustomMessageDataset) error // Called post indexing of the custom messages with the indexed dataset, useful for custom indexing on the whole dataset or for additional processing
	PostSetupCustomFunction             func(PostSetupCustomDataset) error         // Called post setup of the indexer, useful for custom indexing on the whole dataset or for additional processing
	PostSetupDatasetChannel             chan *PostSetupDataset                     // passes configured indexer data to any reader
	PreExitCustomFunction               func(*PreExitCustomDataset) error          // Called post indexing of the custom messages with the indexed dataset, useful for custom indexing on the whole dataset or for additional processing
}

type BlockEventFilterRegistries struct {
	BeginBlockEventFilterRegistry *filter.StaticBlockEventFilterRegistry
	EndBlockEventFilterRegistry   *filter.StaticBlockEventFilterRegistry
}

type DBData struct {
	txDBWrappers []dbTypes.TxDBWrapper
	block        models.Block
}

type BlockEventsDBData struct {
	blockDBWrapper *dbTypes.BlockDBWrapper
}
