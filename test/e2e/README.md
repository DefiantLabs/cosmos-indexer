# E2E Tests

This package provides a testing framework for the cosmos-indexer codebase that makes use of [interchaintest](https://github.com/strangelove-ventures/interchaintest) to setup a local chain Node and run the indexer on it.

## Why a Separate Package?

The cosmos-indexer comes with its own required versions of various packages in the Cosmos SDK ecosystem. The same is true for the interchaintest package. Including this package in the main repo would require pinning the indexer's versioning to that required by interchaintest.

This comes with some difficulties as the e2e test module will need to be clever to run proper tests on the codebase.