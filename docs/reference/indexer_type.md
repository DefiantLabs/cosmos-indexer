# Indexer Type

The `Indexer` type is the main controller for the indexer behavior. It is responsible for managing the indexer workflow and the underlying components that make up the indexer.

For full implementation details, see the [indexer package](https://github.com/DefiantLabs/cosmos-indexer/tree/main/indexer)

## Indexer Type and `index` Command Instantiation

The `Indexer` type contains the following notable elements:

1. A database connection
2. A chain RPC client from the Probe [client package](https://github.com/DefiantLabs/probe/tree/main/client)
3. A Block Enqueue function that handles passing block heights to be indexed to the processors
4. Filter configurations according to the chain's [filter](../usage/filtering.md) configuration
5. Custom Parser types for block events and transaction messages

The `index` command in the application instantiates the `Indexer` type at application runtime and ensures that it is properly configured before starting the indexing workflow.

There are some built-in behaviors that should not be overriden in the `Indexer` type. These are handled by the setup function called in the `index` command. Examples are:

1. Database connection based on the configuration
2. Chain RPC client based on the configuration

However, behavior of the `Indexer` type that can be modified will be noted here.

## Getting the Indexer Instance

The `Indexer` type is instantiated in the `index` command and is available to the application as a `cmd` package global variable.

The `cmd` package provides a `GetBuiltinIndexer() *indexerPackage.Indexer` function that returns this instance, allowing for overrides or calling functions available on the instance.

## Block Enqueue Function - Design and Modification

The Block Enqueue function is responsible for passing blocks to be indexed to the processors. It is the main entrypoint of the indexer workflow. Its only responsibility is to determine the next block height to be processed and pass it to the processors.

For this reason, the `Indexer` type allows for overriding the default Block Enqueue functions based on developer requirements. This can be done by modifying the `EnqueueBlock` function in the `Indexer` instance that is available to the `index` command.

The `BlockEnqueueFunction` function signature is as follows:

```go
func(chan *core.EnqueueData) error
```

The function takes a channel of `core.EnqueueData` and returns an error. The `core.EnqueueData` type is a struct that contains the block height to be indexed and what data should be pulled from the RPC node during RPC requests.

```go
type EnqueueData struct {
	Height            int64
	IndexBlockEvents  bool
	IndexTransactions bool
}
```

The `Height` field is the block height to be indexed. The `IndexBlockEvents` and `IndexTransactions` fields are flags that determine if block events and transactions should be indexed for the block.

At runtime, the `BlockEnqueueFunction` is called with a channel of `core.EnqueueData` that is used to pass blocks to be indexed by the processors.

This allows for various developer overrides of which blocks should be processed during the indexing workflow.

For examples of in-application block enqueue functions see the [core package block_enqueue.go](https://github.com/DefiantLabs/cosmos-indexer/blob/main/core/block_enqueue.go) file. The functions in this package return closures that define highly customized block enqueue functions. These are the built-in block enqueue functions that can be triggered by various configuration variables.

## DB Instance - Gorm Database Connection to PostgreSQL and Modification

The application relies on the [Gorm](https://gorm.io/docs/) ORM library for interacting with the underlying PostgreSQL database. The `Indexer` type contains a `DB` field that is a pointer to the Gorm database connection.

During `index` command setup, the application will connect to the database based on the passed in configuration. The `DB` field is then set on the `Indexer` instance.

However, if a customized Gorm instance is desired, the application will respect `DB` field overrides on the `Indexer` instance if it is not `nil` before setup runs.

## Chain RPC Client - Probe Client Connection

The application relies on the Probe [client package](https://github.com/DefiantLabs/probe/tree/main/client) for interacting with the chain's RPC node. The `Indexer` type contains a `Client` field that is a pointer to the Probe client.

The client package provides functionality that uses built-in Cosmos SDK functionality to make requests to the chain's RPC for raw blockchain data.
