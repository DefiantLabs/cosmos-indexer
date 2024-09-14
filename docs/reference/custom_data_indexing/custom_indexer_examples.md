# Custom Indexer Examples

The `cosmos-indexer` codebase provides a number of examples in the [examples](https://github.com/DefiantLabs/cosmos-indexer/tree/main/examples) directory. These examples are intended to provide a starting point for building custom indexers using the `cosmos-indexer` codebase as an Indexer SDK.

## IBC Patterns Example

The IBC Patterns example demonstrates how to build a custom indexer that indexes IBC packets and acknowledgements. This example indexer is the subject of the [Custom Indexer Walkthrough](./custom_indexer_walkthrough.md) documentation, see that document for a detailed explanation of how it works.

## Governance Patterns Example

The Governance Patterns example demonstrates how to build a custom indexer that indexes governance proposals and votes.

It takes message data from the `cosmos-sdk/x/gov` module and indexes it into a database. The example indexer listens for `MsgSubmitProposal` and `MsgVote` messages and indexes them into a custom model.

The example also implements a filter mechanism to filter out message types that are not of interest to this indexer. This significantly reduces the amount of data that needs to be indexed.

## Validator Delegations Patterns Example

The Validator Delegations Patterns example demonstrates how to build a custom indexer that indexes validator delegations and undelegations.

It takes message data from the `cosmos-sdk/x/staking` module and indexes it into a database. The example indexer listens for `MsgDelegate` and `MsgUndelegate` messages and indexes them into a custom model.

The example also implements a filter mechanism to filter out message types that are not of interest to this indexer. This significantly reduces the amount of data that needs to be indexed.
