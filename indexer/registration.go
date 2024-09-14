package indexer

import (
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/filter"
	"github.com/DefiantLabs/cosmos-indexer/parsers"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func (indexer *Indexer) RegisterCustomModuleBasics(basics []module.AppModuleBasic) {
	indexer.CustomModuleBasics = append(indexer.CustomModuleBasics, basics...)
}

func (indexer *Indexer) RegisterCustomMsgTypesByTypeURLs(customMessageTypeURLSToTypes map[string]sdkTypes.Msg) error {
	if indexer.CustomMsgTypeRegistry == nil {
		indexer.CustomMsgTypeRegistry = make(map[string]sdkTypes.Msg)
	}

	for url, msg := range customMessageTypeURLSToTypes {
		if _, ok := indexer.CustomMsgTypeRegistry[url]; ok {
			return fmt.Errorf("found duplicate message type with URL \"%s\", message types must be uniquely identified", url)
		}
		indexer.CustomMsgTypeRegistry[url] = msg
	}

	return nil
}

func (indexer *Indexer) RegisterMessageTypeFilter(filter filter.MessageTypeFilter) {
	indexer.MessageTypeFilters = append(indexer.MessageTypeFilters, filter)
}

func (indexer *Indexer) RegisterMessageFilter(filter filter.MessageFilter) {
	indexer.MessageFilters = append(indexer.MessageFilters, filter)
}

func (indexer *Indexer) RegisterCustomModels(models []any) {
	indexer.CustomModels = append(indexer.CustomModels, models...)
}

func (indexer *Indexer) RegisterCustomBeginBlockEventParser(eventKey string, parser parsers.BlockEventParser) {
	var err error
	indexer.CustomBeginBlockEventParserRegistry, indexer.CustomBeginBlockParserTrackers, err = customBlockEventRegistration(
		indexer.CustomBeginBlockEventParserRegistry,
		indexer.CustomBeginBlockParserTrackers,
		eventKey,
		parser,
		models.BeginBlockEvent,
	)
	if err != nil {
		config.Log.Fatal("Error registering BeginBlock custom parser", err)
	}
}

func (indexer *Indexer) RegisterCustomEndBlockEventParser(eventKey string, parser parsers.BlockEventParser) {
	var err error
	indexer.CustomEndBlockEventParserRegistry, indexer.CustomEndBlockParserTrackers, err = customBlockEventRegistration(
		indexer.CustomEndBlockEventParserRegistry,
		indexer.CustomEndBlockParserTrackers,
		eventKey,
		parser,
		models.EndBlockEvent,
	)
	if err != nil {
		config.Log.Fatal("Error registering EndBlock custom parser", err)
	}
}

func (indexer *Indexer) RegisterCustomMessageParser(messageKey string, parser parsers.MessageParser) {
	if indexer.CustomMessageParserRegistry == nil {
		indexer.CustomMessageParserRegistry = make(map[string][]parsers.MessageParser)
	}

	if indexer.CustomMessageParserTrackers == nil {
		indexer.CustomMessageParserTrackers = make(map[string]models.MessageParser)
	}

	indexer.CustomMessageParserRegistry[messageKey] = append(indexer.CustomMessageParserRegistry[messageKey], parser)

	if _, ok := indexer.CustomMessageParserTrackers[parser.Identifier()]; ok {
		config.Log.Fatalf("Found duplicate message parser with identifier \"%s\", parsers must be uniquely identified", parser.Identifier())
	}

	indexer.CustomMessageParserTrackers[parser.Identifier()] = models.MessageParser{
		Identifier: parser.Identifier(),
	}
}

func customBlockEventRegistration(registry map[string][]parsers.BlockEventParser, tracker map[string]models.BlockEventParser, eventKey string, parser parsers.BlockEventParser, lifecycleValue models.BlockLifecyclePosition) (map[string][]parsers.BlockEventParser, map[string]models.BlockEventParser, error) {
	if registry == nil {
		registry = make(map[string][]parsers.BlockEventParser)
	}

	if tracker == nil {
		tracker = make(map[string]models.BlockEventParser)
	}

	registry[eventKey] = append(registry[eventKey], parser)

	if _, ok := tracker[parser.Identifier()]; ok {
		return registry, tracker, fmt.Errorf("found duplicate block event parser with identifier \"%s\", parsers must be uniquely identified", parser.Identifier())
	}

	tracker[parser.Identifier()] = models.BlockEventParser{
		Identifier:             parser.Identifier(),
		BlockLifecyclePosition: lifecycleValue,
	}
	return registry, tracker, nil
}
