# Cosmos Indexer

The Cosmos Indexer is an open-source application designed to index a Cosmos chain to a generalized Transaction/Event DB schema. Its mission is to offer a flexible DB schema compatible with all Cosmos SDK Chains while simplifying the indexing process to allow developers flexible ways to store custom indexed data.

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

The Cosmos Indexer tool provides several settings and commands which are accessible via a config file or through CLI flags. You can learn about the CLI flags and their function by running `go run main.go` to display the application help text.

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

#### Probe

The probe section configures [probe](https://github.com/DefiantLabs/probe) used by the tool to read data from the blockchain. This is built into the application and doesn't need to be installed separately.

For detailed descriptions of each setting in these sections, please refer to the [Detailed Config Explanation](#detailed-config-explanation) section below.

## Detailed Config Explanation

This section provides an in-depth description of each setting available in the config file. For further details, refer to the inline documentation within the config file.
