#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

# default value, overide with: make -e FQCN="foo"
FQCN = ghcr.io/nodersteam/cosmos-indexer

grpc_gen:
	protoc proto/*.proto \
        --go_out=./proto \
        --go_opt=paths=source_relative \
        --go-grpc_out=./proto \
        --go-grpc_opt=paths=source_relative \
        --proto_path=./proto

all: install

install: go.sum
	go install .

build:
	go build -o bin/cosmos-indexer .

clean:
	rm -rf build

build-docker-amd:
	docker build -t $(FQCN):$(VERSION) -f ./Dockerfile \
	--build-arg TARGETPLATFORM=linux/amd64 .

build-docker-arm:
	docker build -t $(FQCN):$(VERSION) -f ./Dockerfile \
	--build-arg TARGETPLATFORM=linux/arm64 .

.PHONY: lint
lint: ## Run golangci-linter
	golangci-lint run --out-format=tab

.PHONY: format
format: ## Formats the code with gofumpt
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/*" | xargs gofumpt -w
