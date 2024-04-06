# Indexing

The indexer is built as a single binary to be run against and RPC server to retrieve block and transaction data. This section describes how to run the indexer and what to expect from its behavior.

## Running the indexer

### Basic Usage

The most basic way to run the indexer is like so:

`cosmos-indexer index`

The application will do the following:

1. Look for a config file in the current working directory and use the flags specified in that file (see [Configuration](./configuration.md)) for details
2. Connect to the database
3. Begin the block enqueue process
4. As blocks are enqueued:
   1. Blocks are picked up by RPC request workers
   2. RPC request workers get the data from the blockchain
   3. A processing worker picks up the RPC data and turns it into app-specific data types
   4. App specific data types are picked up by a database worker and inserted into the database

### Providing Flags

Flags can be passed to the indexer on the CLI or through a configuration `.toml` file. Either:

1. Provide a `config.toml` file in CWD or at a path specified with the `--config` flag that defines all required flags
2. Specify all flags that you want to override the default values for at the CLI

Indexer CLI flags are scoped to a generalized structure using `<scope>.<flag>` syntax to improve clarity. For example, base level flags are specified at the `base` scope like `base.start-block` for the block to start indexing from.

**Note**: CLI Flags will take precedence over flags provided in the config file.

### Docker and docker-compose

The application provides a basic, light-weight Dockerfile and docker-compose setup for using the application.

After building the application docker configurations as details in [Installation](./installation.md), the application can be run in the following manner:

Docker:
```bash
docker run -it <image name> cosmos-indexer index <all flags needed>
```

docker-compose:

1. Fill out the `.env.example` file and change its name to `.env` (or provide the environment variables according to the [Docker Compose docs](https://docs.docker.com/compose/environment-variables/set-environment-variables/#use-the-environment-attribute)).
2. Bring up the docker-compose:
    ```
    docker-compose up
    ```

## Advanced Usage

The application behavior can be changed in different ways based on flags and provided configurations. Some examples follow:

### Block List File

A set of specific blocks can be indexed explicitly by providing a block input file.

1. Create a file of block heights like so:
    * block-heights.json:

    ```json
    [
        1,
        2,
        3
    ]
    ```
2. Provide the block input file to the application with the `--base.block-input-file` flag
    ```
    cosmos-indexer index --config="<path to config file>" --base.block-input-file="block-heights.json"
    ```

All flags specific to which blocks to index will be ignored and the indexer will only index the blocks in the file and then exit.

### Message Type Reindexing

It can be useful to enqueue blocks that have a specific message type in the transactions of the block. This behavior has been built directly into a flag.

Start the indexer with the message type URL passed to the `--base.reindex-message-type` flag, e.g.:

```
cosmos-indexer index --config="<path to config file>" --base.reindex-message-type="/cosmos.bank.v1beta1.MsgSend"
```

The indexer will do the following:

1. Find all blocks in the database that have Transactions that contain the specified message type
2. Pass these blocks through the block enqueue process to the indexer workflow
3. Reindex all data for the blocks found

### Indexer Application SDK - Customized Indexing Parsers and Datasets

Advanced users/golang application developers may wish to extend the application to fit their app-specific needs beyond the built-in use-cases presented by the base application. To support this, the cosmos-indexer developers have developed ways to inject custom parsers and models into the application workflow by extending the golang application into a new binary.

The application provides extensive customization methods to insert custom behavior into the main indexing loop, such as:

1. Registering custom models
2. Turning block events or Transaction Messages into custom models
3. Inserting custom models into the application

For examples see the [examples/](https://github.com/DefiantLabs/cosmos-indexer/tree/main/examples) subfolder in the repository.

For reference documentation on how to customize the application code to fit your needs, see the [reference](../reference/README.md) documentation.
