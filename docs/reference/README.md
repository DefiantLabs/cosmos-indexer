# Reference

This sections provides reference documentation on how the codebase works. It also provides documentation on how to use the `cosmos-indexer` codebase as an SDK to build a custom indexer.

## Application Workflow

* [`index` Command](./application_workflow/index_command.md) - The main command that starts the application built using the [cobra](https://cobra.dev/) framework
* [Application Workflow](./application_workflow/application_workflow.md) - The multi-processing workflow used by the application

## Default Data Indexing

The application indexes data into a default shape. The following sections provide details on the datasets that are pulled from the blockchain and the default data indexing:

* [Block Indexed Data](./default_data_indexing/block_indexed_data.md) - The shape of the data for blocks and how the application indexes it
* [Block Events Indexed Data](./default_data_indexing/block_events_indexed_data.md) - The shape of the data for block events and how the application indexes it
* [Transactions Indexed Data](./default_data_indexing/transactions_indexed_data.md) - The shape of the data for transactions and how the application indexes it

## Custom Data Indexing

The application allows for custom data indexing by providing developer access to the underlying types used by the indexer. The following sections provide details on how to use the `cosmos-indexer` codebase as an SDK to build a custom indexer:

* [Indexer Type](./custom_data_indexing/indexer_type.md) - The main controller for indexer behavior and how to modify it
* [Indexer SDK and Custom Parsers](./custom_data_indexing/indexer_sdk_and_custom_parsers.md) - Reference documentation on custom parsers and how to register them
* [Walkthrough](./custom_data_indexing/custom_indexer_walkthrough.md) - A walkthrough of a real world example of creating a custom indexer
* [Examples](./custom_data_indexing/custom_indexer_examples.md) - An explanation of the examples provided in the codebase [examples](https://github.com/DefiantLabs/cosmos-indexer/tree/main/examples) directory

## Custom Cosmos Module Extensions

The application allows for extending the supported transaction message types by providing developer access to the underlying types used by the indexer. This allows developers to bring in custom cosmos modules into the indexer, either through the usage of custom AppModuleBasic implementations with chain-specific message types or through the usage of registering custom message types in the indexer.

Depending on certain factors, such as the version of the Cosmos SDK the custom chain module is based on, developers may need to implement custom message types to be able to decode the transaction messages found on the chain.

The following sections provide details on how to use the `cosmos-indexer` codebase as an SDK to extend the supported transaction message types:

* [Custom Message Type Registration](./custom_cosmos_module_extensions/custom_message_type_registration.md) - Reference documentation on how to register custom message types in the indexer
* [Cosmos Indexer Modules](./custom_cosmos_module_extensions/cosmos_indexer_modules.md) - Reference documentation on the strategy for modules provided by the `cosmos-indexer-modules` package for extending the supported transaction message types
* [Probe Codec Walkthrough](./custom_cosmos_module_extensions/probe_codec_walkthrough.md) - Reference documentation on the probe package and its codec for decoding JSON RPC responses and their Protobuf encoded Transaction Messages