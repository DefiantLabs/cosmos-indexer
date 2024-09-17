# Probe Codec Walkthrough

The Cosmos Indexer uses the [probe](https://github.com/DefiantLabs/probe) package for Cosmos SDK protobuf codec management, RPC client generation, and blockchain RPC data querying/processing.

## Cosmos SDK Codec Management

The Cosmos SDK uses Protobuf for encoding and decoding transaction messages, as well as other blockchain data returned in RPC responses. The `probe` package provides a codec for decoding JSON RPC responses and their Protobuf encoded transaction messages. Types are registered with the codec to allow for decoding and encoding of Protobuf messages to Go types at runtime.

This allows data to be passed from the blockchain to the `probe` package in a format that can be decoded into Go types, allowing for easy processing and indexing of blockchain data.

## RPC Client Generation

The `probe` package also provides a client generator that can be used to generate a client for a specific blockchain. The client provides methods for querying the blockchain for data, such as blocks, transactions, and events. 

The `ChainClient` type defined in the [client package client.go file](https://github.com/DefiantLabs/probe/blob/main/client/client.go) is used to generate the client for a specific blockchain. The client is generated using the `NewChainClient` function, which takes a `ChainClientConfig` type as an argument that contains the configuration for the client, such as RPC endpoint, chain ID and others.

## Blockchain RPC Data Querying/Processing

The `ChainClient` type is attached to a `Query` type defined in the [query package query.go file](https://github.com/DefiantLabs/probe/blob/main/query/query.go) that provides methods for querying the blockchain for data. During JSON RPC response decoding, the codec is used to decode the response into Go types where appropriate.

## Probe Interface Registry

The `probe` package provides an interface registry that allows for registering custom message types with the codec. This registry provides methods for registering custom message types with the codec, allowing for decoding and encoding of custom message types to Go types at runtime by type URL.

There are two main ways to register custom message types with the codec, using the Cosmos SDK `AppModuleBasics` interface to register an entire Cosmos SDK module with the codec, or by using the custom message type URL tied to an underlying type. See the [Custom Message Type Registration](./custom_message_type_registration.md) docs for more details.
