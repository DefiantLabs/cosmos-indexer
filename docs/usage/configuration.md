# Configuration

The application provides extensive configuration options to the `index` command that modify the behavior of the application.

You can run `cosmos-indexer index --help` to get a full list of flag usages, or read on here for a detailed explanation of all the flags.

## Config

All of the following configuration flags can be created in a `.toml` config file and passed to the application that way. See [config.toml.example](../../config.toml.example) for an example that will require further setup.

- **Configuration File**
  - Description: config file location.
  - Flag: `--config`
  - Default Value: `""`
  - Note: default is `<CWD>/config.toml`

## Base Settings - Main

The main base settings are the most important to understand and set.

- **Start Block**
  - Description: Block to start indexing at.
  - Flag: `--base.start-block`
  - Default Value: `0`
  - Note: Use `-1` to resume from the highest block indexed.

- **End Block**
  - Description: Block to stop indexing at.
  - Flag: `--base.end-block`
  - Default Value: `-1`
  - Note: Use `-1` to index indefinitely.

- **Block Input File**
  - Description: A file location containing a JSON list of block heights to index. This flag will override start and end block flags.
  - Flag: `--base.block-input-file`
  - Default Value: `""`

- **Reindex**
  - Description: If true, this will re-attempt to index blocks that have already been indexed.
  - Flag: `--base.reindex`
  - Default Value: `false`

- **Reattempt Failed Blocks**
  - Description: Re-enqueue failed blocks for reattempts at startup.
  - Flag: `--base.reattempt-failed-blocks`
  - Default Value: `false`

- **Reindex Message Type**
  - Description: A Cosmos message type URL. When set, the block enqueue method will reindex all blocks between start and end block that contain this message type.
  - Flag: `--base.reindex-message-type`
  - Default Value: `""`

- **Block Enqueue Throttle Delay**
  - Description: Block enqueue throttle delay.
  - Flag: `--base.throttling`
  - Default Value: `0.5`

## Base Indexing

These flags indicate what will be indexed during the main indexing loop.

- **Transaction Indexing Enabled**
  - Description: Enable transaction indexing.
  - Flag: `--base.index-transactions`
  - Default Value: `false`

- **Block Event Indexing Enabled**
  - Description: Enable block beginblocker and endblocker event indexing.
  - Flag: `--base.index-block-events`
  - Default Value: `false`

## Filter Configurations

- **Filter File**
  - Description: Path to a file containing a JSON config of block event and message type filters to apply to beginblocker events, endblocker events, and TX messages. See [Filtering](./filtering.md) for how to create filters.
  - Flag: `--base.filter-file`
  - Default Value: `""`

## Other Base Settings

- **Dry**
  - Description: Index the chain but don't insert data in the DB.
  - Flag: `--base.dry`
  - Default Value: `false`

- **RPC Workers**
  - Description: The number of concurrent RPC request workers to spin up.
  - Flag: `--base.rpc-workers`
  - Default Value: `1`

- **Wait For Chain**
  - Description: Wait for chain to be in sync.
  - Flag: `--base.wait-for-chain`
  - Default Value: `false`

- **Wait For Chain Delay**
  - Description: Seconds to wait between each check for the node to catch up to the chain.
  - Flag: `--base.wait-for-chain-delay`
  - Default Value: `10`

- **Block Timer**
  - Description: Print out how long it takes to process this many blocks.
  - Flag: `--base.block-timer`
  - Default Value: `10000`

- **Exit When Caught Up**
  - Description: Gets the latest block at runtime and exits when this block has been reached.
  - Flag: `--base.exit-when-caught-up`
  - Default Value: `false`

- **Request Retry Attempts**
  - Description: Number of RPC query retries to make.
  - Flag: `--base.request-retry-attempts`
  - Default Value: `0`

- **Request Retry Max Wait**
  - Description: Max retry incremental backoff wait time in seconds.
  - Flag: `--base.request-retry-max-wait`
  - Default Value: `30`

## Flags

Extended flags that modify how the indexer handles parsed datasets.

- **Index Tx Message Raw**
  - Description: If true, this will index the raw message bytes. This will significantly increase the size of the database.
  - Flag: `--flags.index-tx-message-raw`
  - Default Value: `false`

- **Block Events Base64 Encoded**
  - Description: If true, decode the block event attributes and keys as base64. Some versions of CometBFT encode the block event attributes and keys as base64 in the response from RPC.
  - Flag: `--flags.block-events-base64-encoded`
  - Default Value: `false`

### Logging Configuration

- **Log Level**
  - Description: Log level.
  - Flag: `--log.level`
  - Default Value: `info`

- **Pretty Logs**
  - Description: Enable pretty logs.
  - Flag: `--log.pretty`
  - Default Value: `false`

- **Log Path**
  - Description: Log path. Default is `$HOME/.cosmos-indexer/logs.txt`.
  - Flag: `--log.path`
  - Default Value: `""`

### Database Configuration

- **Database Host**
  - Description: Database host.
  - Flag: `--database.host`
  - Default Value: `""`

- **Database Port**
  - Description: Database port.
  - Flag: `--database.port`
  - Default Value: `5432`

- **Database Name**
  - Description: Database name.
  - Flag: `--database.database`
  - Default Value: `""`

- **Database User**
  - Description: Database user.
  - Flag: `--database.user`
  - Default Value: `""`

- **Database Password**
  - Description: Database password.
  - Flag: `--database.password`
  - Default Value: `""`

- **Database Log Level**
  - Description: Database log level.
  - Flag: `--database.log-level`
  - Default Value: `""`

### Probe Configuration

These flags modify the behavior of the usage of the [probe](https://github.com/DefiantLabs/probe) package, which is the main way the application uses to get data from the RPC server.

- **Node RPC Endpoint**
  - Description: Node RPC endpoint.
  - Flag: `--probe.rpc`
  - Default Value: `""`

- **Probe Account Prefix**
  - Description: Probe account prefix.
  - Flag: `--probe.account-prefix`
  - Default Value: `""`

- **Probe Chain ID**
  - Description: Probe chain ID.
  - Flag: `--probe.chain-id`
  - Default Value: `""`

- **Probe Chain Name**
  - Description: Probe chain name.
  - Flag: `--probe.chain-name`
  - Default Value: `""`
