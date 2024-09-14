# Indexer SDK and Custom Parsers

The `cosmos-indexer` relies on the `Indexer` type from the `indexer` package to control the behavior of the indexer. The `Indexer` type is instantiated in the `index` command and is available to the application through a getter function in the `cmd` package.

## Getter

The `cmd` package provides a `GetBuiltinIndexer() *indexerPackage.Indexer` function that returns the `Indexer` instance. This allows for certain overrides or calling functions available on the instance.

```go
indexer := cmd.GetBuiltinIndexer()
```

Certain changes made to the indexer type will be persisted when calling the `index` command

## Custom Type Registration

The `Indexer` type provides registration functions that will modify the behavior of the indexer. The following registration functions are available on the `Indexer` type in the [registration.go file](https://github.com/DefiantLabs/cosmos-indexer/blob/30f689fc4914f41cb5b7599a9e6ef730d71a7c3d/indexer/registration.go) in the `indexer` package:

1. `RegisterCustomModuleBasics` - Registers custom module basics for the chain, used for injecting custom Cosmos SDK modules into the Codec for the chain to allow RPC parsing of custom module transaction messages
2. `RegisterMessageTypeFilter` - Registers a message type filter for the chain, used for filtering out transaction messages that should not be indexed. Allows SDK access to the UX-provided message type filter described in the [filtering](../usage/filtering.md) documentation
3. `RegisterCustomModels` - Registers custom models into the application's database schema. These will be migrated into the database when the application starts. Used for custom data storage.
4. `RegisterCustomBeginBlockEventParser` - Registers a custom begin block event parser for the chain, used for parsing custom begin block events into custom data types
5. `RegisterCustomEndBlockEventParser` - Registers a custom end block event parser for the chain, used for parsing custom end block events into custom data types
6. `RegisterCustomMessageParser` - Registers a custom message parser for the chain, used for parsing custom transaction messages into custom data types

When these functions are called before the `index` command is executed, the custom behavior will be persisted in the indexer instance. During the application workflow, the indexer will call custom parsers during data processing and database insertion steps.

## Custom Parser Interfaces

The `cosmos-indexer` application provides interfaces for custom parsers to implement. These interfaces are used by the indexer to call custom parsing functions during the indexing workflow. You can find the definitions of the interfaces in the [parsers package](https://github.com/DefiantLabs/cosmos-indexer/tree/main/parsers).There are 2 types of custom parser interfaces available in the application:

1. `BlockEventParser` - Used for parsing block events into custom data types
2. `MessageParser` - Used for parsing transaction messages into custom data types

These are highly generalized interfaces with a reliance on type wrappers and Go `any` types to transport the parsed dataset along the workflow.

SDK developer users should implement these interfaces in their custom parsers to ensure that the indexer can call the custom parsing functions during the indexing workflow.

Each of the custom parser registration functions in the `Indexer` type will take a custom parser that implements one of these interfaces and a unique identifier. The custom parser will be called during the indexing workflow to parse the data into custom data types and insert it into the database.
