# Custom Message Type Registration

The Cosmos Indexer comes with a built-in codec that is used to decode JSON RPC responses and their Protobuf encoded transaction messages. The codec is used to decode the response into Go types where appropriate. By default, the codec is set up to handle the base Cosmos SDK modules.

However, the codec provides a way to register custom message types with the codec, allowing for decoding and encoding of custom message types to Go types at runtime by type URL.

The two main ways to register custom message types with the codec are:

1. Using the Cosmos SDK `AppModuleBasics` interface to register an entire Cosmos SDK module with the codec by providing the module's `AppModuleBasic` implementation to the `Indexer` type before application execution
2. Using the custom message type URL tied to an underlying type to register custom message types with the codec

These methods are described in detail below.

## Module Registration using AppModuleBasics

In normal usage of the Cosmos SDK, message types are registered with the codec using the `AppModuleBasics` interface. The `RegisterInterfaces` method of the `module.BasicManager` interface is used to register custom message types with the codec. This is how the base Cosmos SDK modules are provided in the `probe` package and are used by the Indexer by default.

This is done by:

1. Pulling the base Cosmos SDK modules into the `probe` client package and providing them in the `DefaultModuleBasics` variable, as can be seen in the [probe client package config.go file](https://github.com/DefiantLabs/probe/blob/main/client/config.go#L30-L31)
2. During `probe` client creation, using these module basics to register the base Cosmos SDK modules with the `probe` `ChainClientConfig` type, as can be seen in the [cosmos-indexer probe probe.go file](https://github.com/DefiantLabs/cosmos-indexer/blob/main/probe/probe.go#L26-L27)
3. These module basics then have thier module-specific interfaces registered with the codec during `probe` client creation, as can be seen in the [probe client encoding.go file](https://github.com/DefiantLabs/probe/blob/main/client/encoding.go#L30) `MakeCodec` function.

The list of AppModuleBasics registered to the probe client can be extended to include new modules. The `Indexer` type provides a `RegisterCustomModuleBasics` in the [indexer package types.go file](https://github.com/DefiantLabs/cosmos-indexer/blob/main/indexer/registration.go#L14-L16) method that registers custom module basics with the indexer. This provides the `probe` client with the ability to register module-specific message types with the codec.

The main difficulty with this approach is that it requires the developer to have access to the module's `AppModuleBasic` implementation. This is not always possible, especially when dealing with custom modules that are not part of the base Cosmos SDK. For example, the following list of reasons, amongst others, may prevent the developer from using the `AppModuleBasic` interface:

1. The module is not part of the base Cosmos SDK and does not use the exact version of the Cosmos SDK that the `cosmos-indexer` package is built on.
2. The module is not open source and the developer does not have access to the module's `AppModuleBasic` implementation for registration.

In these cases, the developer can use the custom message type URL registration method instead.

## Custom Message Type Registration using Type URL

This is useful for extending the supported transaction message types in the Cosmos Indexer. For instance, the `RegisterCustomTypeURL` function in the [client codec types package interface_registry.go file](https://github.com/DefiantLabs/probe/blob/main/client/codec/types/interface_registry.go) can be used to register custom message types with the codec.

This is exactly how the Cosmos Indexer extends the supported transaction message types. The `Indexer` provides a [`RegisterCustomMsgTypesByTypeURLs`]((https://github.com/DefiantLabs/cosmos-indexer/blob/main/indexer/registration.go#L18-L19)) method that registers custom message types with the indexer. During application setup, custom message types are registered with the `probe` package codec during `ChainClient` creation. This process is handled in the setup in the following manner:

1. During the setup of the Indexer in the [cosmos-indexer/cmd package index.go file](https://github.com/DefiantLabs/cosmos-indexer/blob/main/cmd/index.go#L200-L201), the `GetProbeClient` function is called with the registered custom message types.
2. The `GetProbeClient` function in the [cosmos-indexer/probe package probe.go file](https://github.com/DefiantLabs/cosmos-indexer/blob/main/probe/probe.go#L10) creates a `ChainClientConfig` with the custom message types registered
3. The `ChainClientConfig` is passed to the `NewChainClient` function in the [probe/client package client.go file](https://github.com/DefiantLabs/probe/blob/main/client/client.go#L28)
4. The `ChainClient` is created with the custom message types registered with the codec during the `MakeCodec` function in the [probe client encoding.go file](https://github.com/DefiantLabs/probe/blob/main/client/encoding.go#L30) `MakeCodec` function.
