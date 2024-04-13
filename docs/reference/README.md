# Reference

This sections provides reference documentation on how the codebase works. It also provides documentation on how to use the `cosmos-indexer` codebase as an SDK to build a custom indexer.

## Application Workflow

* [`index` Command](./index_command.md) - The main command that starts the application built using the [cobra](https://cobra.dev/) framework
* [Application Workflow](./application_workflow.md) - The multi-processing workflow used by the application

## Default Data Indexing

The application indexes data into a default shape. The following sections provide details on the datasets that are pulled from the blockchain and the default data indexing:

* [Block Indexed Data](./block_indexed_data.md) - The shape of the data for blocks and how the application indexes it
* [Block Events Indexed Data](./block_events_indexed_data.md) - The shape of the data for block events and how the application indexes it
* [Transactions Indexed Data](./transactions_indexed_data.md) - The shape of the data for transactions and how the application indexes it

## Custom Data Indexing

The application allows for custom data indexing by providing developer access to the underlying types used by the indexer. The following sections provide details on how to use the `cosmos-indexer` codebase as an SDK to build a custom indexer:

* [Indexer Type](./indexer_type.md) - The main controller for indexer behavior and how to modify it
* [Indexer SDK and Custom Parsers](./indexer_sdk_and_custom_parsers.md) - Reference documentation on custom parsers and how to register them
* [Walkthrough](./custom_indexer_walkthrough.md) - A walkthrough of a real world example of creating a custom indexer
* [Examples](./custom_indexer_examples.md) - An explanation of the examples provided in the codebase [examples](https://github.com/DefiantLabs/cosmos-indexer/tree/main/examples) directory
