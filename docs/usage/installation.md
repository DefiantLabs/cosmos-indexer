# Installation

The indexer provides multiple installation methods based on preference.

## Releases

Visit the [Releases](https://github.com/DefiantLabs/cosmos-indexer/releases) page to keep up to date with the latest releases.

## Building from Source

You may install the application by building from source.

Prerequisites:

1. make
2. Go 1.19+
3. The repository downloaded


From the root of the codebase run:

```
make install
```

Run the following to ensure the installation is available:

```
cosmos-indexer --help
```

## Dockerfile

The root of the codebase provides a Dockerfile for building a light-weight image that contains the installation of the indexer.

Prerequisites:

1. Docker
2. (Optional) docker-compose


From the root of the codebase run:

```
docker build -t cosmos-indexer .
```

Run the following to ensure the docker build was successful:

```
docker run -it cosmos-indexer cosmos-indexer --help
```

Optionally, the repo has provided a generalized `docker-compose.yaml` file that makes use of:

1. A PostgreSQL container with a volume for the database
2. An indexer service that uses the repo Dockerfile and a set of required environment variables at runtime

From the root of the codebase run:

```
docker-compose build
```
