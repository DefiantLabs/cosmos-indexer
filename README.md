# Cosmos Tax CLI

The Cosmos Tax CLI is an open-source application designed to index a Cosmos chain to a generalized Taxable Transaction/Event DB schema. Its mission is to offer a flexible DB schema compatible with all Cosmos SDK Chains while simplifying the correlation of transactions with addresses and storing relevant data.

In addition to indexing a chain, this tool also queries the indexed data to find all transactions associated with one or more addresses on a specific chain.

**Note**: Work is actively being done on a more generalized Cosmos Indexer model. That work can be seen in the [Cosmos Indexer](https://github.com/DefiantLabs/cosmos-indexer) project.

**Watch our Tool Overview and How-To Videos:**
- [Overview of Cosmos Tax CLI](https://www.youtube.com/watch?v=Vx3t8uCnHqE)

## :star: Funding

The development and evolution of Cosmos Tax CLI across versions v0.1.0, v0.2.0, and v0.3.0 were significantly supported by the **Interchain Foundation**, facilitated by **Strangelove-Ventures**. After December 31, 2022, all subsequent development has been self-funded by **Defiant Labs**, demonstrating our commitment to this project and its potential.

## :handshake: Integrations

Applications like [Sycamore](https://app.sycamore.tax/) have been built on top of our Cosmos Tax CLI, showcasing its functionality and adaptability. If you're looking to integrate our indexer into your project, you can do so in the following ways:

1. **Direct Indexing**: Directly index the chain data into your own Database.
2. **API Usage**: Use our APIs to access the data and incorporate it into your service.
3. **Custom**: Work with our experts for custom integrations. DM us on twitter [@defiantlabs](https://twitter.com/defiantlabs).

This project is open-source, and we encourage developers to build upon our work, enhancing the Cosmos ecosystem.

## Quick Start

You can use our `docker-compose` file for a quick demo of how to run the indexer, DB, web client, and UI.

```shell
docker compose up
```
Keep an eye on the output for the index and access the web server through the provided link.

## Getting Started

It's indexing time! Follow the steps below to get started.

### Prerequisites

Before you can start indexing a chain, you need to set up the application's dependencies:

#### PostgreSQL
The application requires a PostgreSQL server with an established database and an owner user/role with password login. Here's a simple example of setting up a containerized database locally [here](https://towardsdatascience.com/local-development-set-up-of-postgresql-with-docker-c022632f13ea).

#### Go
The application is written in Go, so you need to build it from source. This requires a system installation of at minimum Go 1.19. Instructions for installing and configuring Go can be found [here](https://go.dev/doc/install).

## Indexing and Querying

You are now ready to index and query the chain. For detailed steps, check out the [Indexing](#indexing) and [Querying](#querying) sections below.

## CLI Syntax

The Cosmos Tax CLI tool provides several settings and commands which are accessible via a config file or through CLI flags. You can learn about the CLI flags and their function by running `go run main.go` to display the application help text.

For more detailed information on the settings, refer to the [Config](#config) section.

### Config

The config file, used to set up the Cosmos Tax CLI tool, is broken into four main

 sections:

1. [Log](#log)
2. [Database](#database)
3. [Base](#base)
4. [Probe](#probe)

#### Log

The Log section covers settings related to logging levels and formats, including log file paths and whether to use [ZeroLog's](https://github.com/rs/zerolog) pretty logging.

#### Database

The Database section defines the settings needed to connect to the database server and to configure the logging level of the ORM.

#### Base

The Base section contains the core settings for the tool, such as API endpoints, block ranges, indexing behavior, and more.

#### Lens

The probe section configures [lens](https://github.com/DefiantLabs/lens) used by the tool to read data from the blockchain. This is built into the application and doesn't need to be installed separately.

For detailed descriptions of each setting in these sections, please refer to the [Detailed Config Explanation](#detailed-config-explanation) section below.

## Detailed Config Explanation

This section provides an in-depth description of each setting available in the config file. For further details, refer to the inline documentation within the config file.

# 📝 Supported Message Types

During the chain indexing process, we parse individual messages to determine their significance. Certain messages, like **transfers**, carry tax implications, whereas others, such as **bonding/unbonding funds**, don't. Additionally, some message types have only partial support or are under active development, which safeguards the indexer from encountering errors when processing these types.

The applications current data model is oriented toward indexing data with taxable implications. Transaction messages that imply the taxable sending or receiving of tokens are indexed according to this data model. While message types that do not imply taxable sends and receives skip this process, we do still index these message types in the following manner:

1. All transactions are indexed along with the signer of the transaction
2. Fees for every transaction are indexed, no matter the taxable implications
3. All transaction messages with their type are indexed alongside the transaction they were executed in

While we strive to expand our list of supported messages, we acknowledge that we do not yet cover every possible message across all chains. If you identify a missing or improperly handled message type, we encourage you to **open an issue or submit a PR**.

For the most recent, comprehensive list of supported messages, please refer to the code [**here**](https://github.com/DefiantLabs/cosmos-tax-cli/blob/main/core/tx.go).

Below is the rundown of our current support for different types of messages:

## 🌌 Cosmos Modules
### 🛡️ Authz
- `MsgExec`
- `MsgGrant`
- `MsgRevoke`

### 🏦 Bank
- `MsgSendV0` (deprecated)
- `MsgSend`
- `MsgMultiSendV0` (deprecated)
- `MsgMultiSend`

### 📈 Distribution
- `MsgFundCommunityPool`
- `MsgWithdrawValidatorCommission`
- `MsgWithdrawDelegatorReward`
- `MsgWithdrawRewards`
- `MsgSetWithdrawAddress`

### 🏛️ Gov
- `MsgVote`
- `MsgDeposit`
- `MsgSubmitProposal`
- `MsgVoteWeighted`

### 🌐 IBC
- `MsgTransfer`
- `MsgAcknowledgement`
- `MsgChannelOpenTry`
- `MsgChannelOpenConfirm`
- `MsgChannelOpenInit`
- `MsgChannelOpenAck`
- `MsgRecvPacket`
- `MsgTimeout`
- `MsgTimeoutOnClose`
- `MsgConnectionOpenTry`
- `MsgConnectionOpenConfirm`
- `MsgConnectionOpenInit`
- `MsgConnectionOpenAck`
- `MsgCreateClient`
- `MsgUpdateClient`

### ⛏️ Slashing
- `MsgUnjail`
- `MsgUpdateParams`

### 🪙 Staking
- `MsgDelegate`
- `MsgUndelegate`
- `MsgBeginRedelegate`
- `MsgCreateValidator`
- `MsgEditValidator`

### ⏳ Vesting
- `MsgCreateVestingAccount`

## 🌊 Osmosis Modules

### 🎯 Concentrated Liquidity
- `MsgCreatePosition`
- `MsgWithdrawPosition`
- `MsgCollectSpreadRewards`
- `MsgCreateConcentratedPool`
- `MsgCollectIncentives`
- `MsgAddToPosition`

### 🔄 Gamm
- `MsgSwapExactAmountIn`
- `MsgSwapExactAmountOut`
- `MsgJoinSwapExternAmountIn`
- `MsgJoinSwapShareAmountOut`
- `MsgJoinPool`
- `MsgExitSwapShareAmountIn`
- `MsgExitSwapExternAmountOut`
- `MsgExitPool`

### 🎁 Incentives
- `MsgCreateGauge`
- `MsgAddToGauge`

### 🔒 Lockup
- `MsgBeginUnlocking`
- `MsgLockTokens`
- `MsgBeginUnlockingAll`

### 🌊 Superfluid
- `MsgSuperfluidDelegate`
- `MsgSuperfluidUndelegate`
- `MsgSuperfluidUnbondLock`
- `MsgLockAndSuperfluidDelegate`
- `MsgUnPoolWhitelistedPool`

### 🌟 Valset-Pref
- `MsgSetValidatorSetPreference`
- `MsgDelegateToValidatorSet`
- `MsgUndelegateFromValidatorSet`
- `MsgRedelegateValidatorSet`
- `MsgWithdrawDelegationRewards`
- `MsgDelegateBondedTokens`

## ⭐ Tendermint Modules
### 💧 Liquidity
- `MsgCreatePool`
- `MsgDepositWithinBatch`
- `MsgWithdrawWithinBatch`
- `MsgSwapWithinBatch`

## 🌐 CosmWasm Modules
### 🧩 Wasm (Coming soon)
- `MsgExecuteContract`
- `MsgInstantiateContract`
