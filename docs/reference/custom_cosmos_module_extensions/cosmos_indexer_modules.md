# Cosmos Indexer Modules

The [`cosmos-indexer-modules`](https://github.com/DefiantLabs/cosmos-indexer-modules) package provides a set of modules that extend the supported transaction message types in the Cosmos Indexer. These modules are used to extend the supported transaction message types in the Cosmos Indexer by providing custom message types for Cosmos SDK modules that are not part of the base Cosmos SDK.

## Strategy

The `cosmos-indexer-modules` package provides a set of packages that allow access to a type URL mapping for custom message types. These mappings can be used to register custom message types with the codec in the Cosmos Indexer.

The types defined in the subpackages must fit the Cosmos SDK `Msg` interface, which is used to define the transaction message types in the Cosmos SDK. These types are protobuf messages that are used to define the transaction messages that are sent to the blockchain and returned in the blockchain responses.

The `cosmos-indexer-modules` package includes full `Msg` implementations for various (and growing) Cosmos SDK modules. This is achieved by generating the protobuf message types for the modules and implementing the `Msg` interface for each message type. These are then provided in a module-specific type URL mapping that can be used to register the custom message types with the codec in the Cosmos Indexer.

## Usage

The following shows usage of how one of the `cosmos-indexer-modules` packages can be used to extend the supported transaction message types in the Cosmos Indexer. The `cosmos-indexer-modules` package contains a [`block-sdk`](https://github.com/DefiantLabs/cosmos-indexer-modules/tree/main/block-sdk) subpackage that provides a set of custom message types for the [Skip MEV `blocksdk`](https://github.com/skip-mev/block-sdk) module.

The `block-sdk` package defines a [`GetBlockSDKTypeMap` function](https://github.com/DefiantLabs/cosmos-indexer-modules/blob/main/block-sdk/msg_types.go#L17-L26) that returns a map of type URLs to the custom message types for the `blocksdk` module. This map can be used to register the custom message types with the codec in the Cosmos Indexer. The underlying types have been generated using protobuf definitions for the `blocksdk` module.

These can be passed to the `RegisterCustomMsgTypesByTypeURLs` method in the `Indexer` type in the Cosmos Indexer to register the custom message types with the codec. This allows the Cosmos Indexer to decode and encode the custom message types to Go types at runtime.

```go
package main

import (
	"log"

	blockSDKModules "github.com/DefiantLabs/cosmos-indexer-modules/block-sdk"
	"github.com/DefiantLabs/cosmos-indexer/cmd"
)

func main() {
	indexer := cmd.GetBuiltinIndexer()

	indexer.RegisterCustomMsgTypesByTypeURLs(blockSDKModules.GetBlockSDKTypeMap())

	err := cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
```

By providing the custom message types for the `blocksdk` module, the Cosmos Indexer can now decode custom message types to Go types at runtime. This allows the Cosmos Indexer to extend the supported transaction message types and index the custom message types for the `blocksdk` module.

This example can be found in the `cosmos-indexer-modules` [examples/block-sdk-indexer](https://github.com/DefiantLabs/cosmos-indexer/tree/d020840f44775bf1680765867d54338592ac3caa/examples/block-sdk-indexer) codebase. This example also provides an example `filter.json` file for indexing only the `blocksdk` module messages, which conforms to the filter file creation requirements documented in the [Filtering](../../usage/filtering.md) doc.
