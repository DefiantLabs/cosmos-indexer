# Quickstart

This guide will help you get up and running with the Cosmos Indexer quickly. The Cosmos Indexer is a tool for indexing and querying data from Cosmos SDK based blockchains.

## Installation

Download the latest release from the [Releases](https://github.com/DefiantLabs/cosmos-indexer/releases) page.

## Configuration

The Cosmos Indexer uses a `.toml` configuration file to set up the indexer. The configuration file is used to set up the database connection, the chain configuration, and the indexer configuration.

1. Take the configuration file example (config.toml.example) from the root of the repository and copy it to a new file named `config.toml`.
2. Edit the `config.toml` file to match your database connection and chain configuration.
3. Save the `config.toml` file in a location that the indexer can access.

## Running the Indexer

The indexer can be run using the following command:

```bash
cosmos-indexer index --config /path/to/config.toml
```

The indexer will start and begin indexing blocks from the chain. The indexer will continue to run based on the configuration values passed in for start and end blocks.

## Important Configuration Values

The following configuration values are important to understand when setting up the indexer:

- `database.*` - The database connection configuration
- `probe.*` - The probe configuration, used to determine the RPC connection to the chain
- `base.start-block` - The block to start indexing from
- `base.end-block` - The block to end indexing at
- `base.index-transactions` - Whether to index transactions
- `base.index-block-events` - Whether to index block events
