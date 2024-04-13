# Indexer SDK and Custom Parsers

The `cosmos-indexer` application allows for custom block event and transaction message parsers to be injected into the indexing workflow. This allows for custom data indexing and storage based on the chain's data structure.

The workflow for building custom parsers has the following overview:

1. Create a new `main.go` file and import the `cosmos-indexer` package
2. Get the built-in `Indexer` instance
3. Register custom models into the application's database schema
4. Create custom parsers for block events or transaction messages
5. Register the custom parsers to inject them into the indexer workflow
6. Register message type filters to reduce the size of the data being indexed
7. Start the root command to begin the indexing workflow

The code examples in this section will build up a custom parser for a hypothetical custom parser for IBC Transactions.

The example indexer for this walkthrough can be found in the [examples/ibc-patterns](https://github.com/DefiantLabs/cosmos-indexer/tree/main/examples/ibc-patterns) directory.

## Step 1 - Creating a New `main.go` File and Importing the `cosmos-indexer` Package

Just as with any Go application, the first step is to create a new `main.go` file and import the `cosmos-indexer` package.

```go
package main

import (
	"log"

	"github.com/DefiantLabs/cosmos-indexer/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
```

This basic `main.go` file imports the `cosmos-indexer` package and starts the application. The `cmd.Execute()` function is the main entrypoint for the application and starts the `index` command. This is the exact same setup as the `main.go` file found in the root of the repository.

## Step 2 - Getting the Built-In Indexer Instance

The `Indexer` type is instantiated in the `index` command and is available to areas outside of the package.

The `cmd` package provides a `GetBuiltinIndexer() *indexerPackage.Indexer` function that returns this instance, allowing for certain overrides or calling functions available on the instance.

```
indexer := cmd.GetBuiltinIndexer()
```

## Step 3 - Registering Custom Models

The `cosmos-indexer` application uses the Gorm ORM to interact with the database. To register custom models, you can use the `RegisterModels` function on the `Indexer` instance.

This function takes an array of `any`, appending the items to the list of `any` custom models that are tracked by the application.

```go

import (
	"log"

	"github.com/DefiantLabs/cosmos-indexer/cmd"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
)

type MsgType int

const (
	MsgRecvPacket MsgType = iota
	MsgAcknowledgement
)

type TransactionType string

const (
	ValidatorUpdatesTransactionType TransactionType = "ValidatorUpdates"
	TokenTransferTransactionType    TransactionType = "TokenTransfer"
	CCVSlashTransactionType         TransactionType = "CCVSlash"
	CCVVSCMaturedTransactionType    TransactionType = "CCVVSCMatured"
)

type IBCTransactionType struct {
	ID              uint `gorm:"primaryKey"`
	TransactionType TransactionType
}

type IBCTransaction struct {
	ID                   uint `gorm:"primaryKey"`
	SourceChannel        string
	SourcePort           string
	DestinationChannel   string
	DestinationPort      string
	MessageID            uint `gorm:"uniqueIndex"`
	Message              models.Message
	IBCMsgType           MsgType
	IBCTransactionTypeID uint
	IBCTransactionType   IBCTransactionType
}

func main() {
	indexer := cmd.GetBuiltinIndexer()

	indexer.RegisterCustomModels([]any{IBCTransactionType{}, IBCTransaction{}})
    // ... rest of the main function
}
```

We intend to parse out the type of IBC message from the data inside of it. For this, we have a custom model `IBCTransactionType` that will store the type of IBC message. We also have a custom model `IBCTransaction` that will store the parsed data from the IBC message. This model also contains the indexer's built-in `Message` model as a foreign key, allowing us to link our parsed data to the default dataset of the indexer.

## Step 4 - Creating Custom Parsers

Custom parsers are used to parse block events and transaction messages into custom data types. The `cosmos-indexer` application provides interfaces for custom parsers to implement. These interfaces are used by the indexer to call custom parsing functions during the indexing workflow.

The walkthrough will focus on creating a custom parser for IBC Transactions. The custom parser will implement the `MessageParser` interface:

```go
type MessageParser interface {
	Identifier() string
	ParseMessage(sdkTypes.Msg, *txtypes.LogMessage, config.IndexConfig) (*any, error)
	IndexMessage(*any, *gorm.DB, models.Message, []MessageEventWithAttributes, config.IndexConfig) error
}
```

Our parser will be able to handle a number of different IBC transaction message types. We will create a custom parser that can parse these transactions into custom data types.

### Define the Custom Parser Struct that Will Satisfy the `MessageParser` Interface

```go
type IBCTransactionParser struct {
	UniqueID string // Used to identify the parser and register it with a message type URL
}
```

### Implement the Identifier Function

The `Identifier` function is used to return the unique identifier for the parser. This identifier is used to register the parser with the indexer at a particular message type URL for faster lookups of parsers.

```go
func (p *IBCTransactionParser) Identifier() string {
	return p.UniqueID
}
```

### Implement the ParseMessage Function

The `ParseMessage` function is used to parse the transaction message into a custom data type. The function takes the Cosmos SDK message, the log message, and the index configuration as arguments. The interface function should return a pointer to an `any` type that contains the parsed data, and an error for if something went wrong. This design was chosen for maximum flexibility. It allows the indexer to pass the dataset to downstream functions without knowing the underlying data type.

Another decision to make is whether to split out our parsers into separate functions or try to implement the functionality in a single parser. Since parsers can be registered with multiple message types, the decision is up to the developer whether to define a single parser for all message types or multiple parsers for each message type.

The plan here is to parse IBC channel messages into a custom data type that can be inserted into the database. We pick MsgAcknowledgement and MsgRecvPacket for the messages to parse since these are the "success" message types that actually indicate a successful IBC transaction.

To determine the type of transaction at runtime, we take advantage of the Cosmos SDK to cast the message to the appropriate type and check if it is the type we are looking for. If it is, we parse the message into our custom data type.

```go

import (

	//Updated imports for the ParseMessage command
	"github.com/DefiantLabs/cosmos-indexer/config"

	indexerTxTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	stdTypes "github.com/cosmos/cosmos-sdk/types"
	chanTypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

func (c *IBCTransactionParser) ParseMessage(cosmosMsg stdTypes.Msg, log *indexerTxTypes.LogMessage, cfg config.IndexConfig) (*any, error) {

	// Check if this is a MsgAcknowledgement
	msgAck, okMsgAck := cosmosMsg.(*chanTypes.MsgAcknowledgement)

	if okMsgAck {
		parsedMsgAck, err := parseMsgAcknowledgement(msgAck)

		if err != nil {
			return nil, fmt.Errorf("error parsing MsgAcknowledgement: %w", err)
		}

		anyCast := any(parsedMsgAck)

		return &anyCast, nil
	}

	// Check if this is a MsgRecvPacket
	msgRecvPacket, okMsgRecvPacket := cosmosMsg.(*chanTypes.MsgRecvPacket)

	if okMsgRecvPacket {
		parsedMsgRecvPacket, err := parseMsgRecvPacket(msgRecvPacket)

		if err != nil {
			return nil, fmt.Errorf("error parsing MsgAcknowledgement: %w", err)
		}

		anyCast := any(parsedMsgRecvPacket)

		return &anyCast, nil
	}

	return nil, fmt.Errorf("unsupported message type passed to parser")
}
```

With the main parsing function implemented, we can now define the parsing functions for the MsgAcknowledgement and MsgRecvPacket message types. Since these functions can return any data type, and our custom indexer contains 2 models, we will define a wrapper struct for the main parsed dataset.

```go
type IBCTransactionParsedData struct {
	ParsedIBCMessage *IBCTransaction
	IBCMessageType *IBCTransactionType
}
```

This will be the `any` value returned from the `parseMsgAcknowledgement` and `parseMsgRecvPacket` functions if they do not return an error. For brevity, the implementation details for these 2 functions will be left out of this walkthrough. The full implementation can be found in the [examples/ibc-patterns](https://github.com/DefiantLabs/cosmos-indexer/tree/main/examples/ibc-patterns) directory. To summarize the implementations, the functions will parse out the message data to determine the basic IBC source/destination types, and then parse out the packet data to determine the type of IBC message.

### Implement the IndexMessage Function

The `IndexMessage` function is used to insert the parsed data into the database. The function takes the parsed data, the Gorm database connection, the built-in message model, the message events with attributes, and the index configuration as arguments. All of this data can be used to insert the parsed data into the database depending on the indexer developer's requirements.

This function, when called on the parser, is wrapped in a database transaction to ensure that the data is inserted into the database in a consistent manner. The function **must** return an error if something goes wrong during the database insertion.

```go

import (

	//Updated imports for the IndexMessage command
	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/parsers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	indexerTxTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	stdTypes "github.com/cosmos/cosmos-sdk/types"
	chanTypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

func (c *IBCTransactionParser) IndexMessage(dataset *any, db *gorm.DB, message models.Message, messageEvents []parsers.MessageEventWithAttributes, cfg config.IndexConfig) error {

	ibcTransaction, ok := (*dataset).(IBCTransactionParsedData)

	if !ok {
		return fmt.Errorf("invalid IBC transaction type passed to parser index message function")
	}

	ibcTransactionType := ibcTransaction.IBCTransactionType
	parsedIBCMessage := ibcTransaction.ParsedIBCMessage

	// Create or update the IBC transaction type
	err := db.Where(&ibcTransactionType).FirstOrCreate(&ibcTransactionType).Error

	if err != nil {
		return err
	}

	// Set the IBC transaction type ID on the IBC transaction and link it to the default message model
	parsedIBCMessage.IBCTransactionTypeID = ibcTransactionType.ID
	parsedIBCMessage.IBCTransactionType = *ibcTransactionType
	parsedIBCMessage.Message = message
	parsedIBCMessage.MessageID = message.ID

	// Create or update the IBC transaction

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"source_channel", "destination_channel", "message_id", "ibc_msg_type", "ibc_transaction_type_id"}),
	}).Create(parsedIBCMessage).Error; err != nil {
		return err
	}

	return nil
}
```

This function takes advantage of the Gorm ORM to insert the parsed data into the database. The function first checks if the dataset is of the correct type, then inserts the IBC transaction type into the database. It then links the IBC transaction type to the IBC transaction and inserts the IBC transaction into the database. Due to the usage of a unique index on the `message_id` foreign key, the function will either create a new IBC transaction or update an existing one if the message ID already exists in the database.

## Step 5 - Registering the Custom Parsers

To inject these new message parsers into the indexer workflow, we need to register them with the indexer instance. The `RegisterCustomMessageParser` function on the `Indexer` instance is used to register custom message parsers.

This function expects a message type URL and a custom parser that implements the `MessageParser` interface. The message type URL is used to register the parser with the indexer for faster lookups of parsers. Multiple parsers can be registered to the same type URL, and they will be called in order of registration.

```go
func main() {
	// ... previous code
	ibcRecvParser := &IBCTransactionParser{
		UniqueID: "ibc-recv-parser",
	}

	ibcAckParser := &IBCTransactionParser{
		UniqueID: "ibc-ack-parser",
	}

	indexer.RegisterCustomMessageParser("/ibc.core.channel.v1.MsgRecvPacket", ibcRecvParser)
	indexer.RegisterCustomMessageParser("/ibc.core.channel.v1.MsgAcknowledgement", ibcAckParser)

	// ... rest of the main function
}
```

This function also sets up parser trackers that will track the execution of the parser during the indexing workflow. The parser trackers are used to fill out the `MessageParsers` and `MessageParserErrors` models in the database. This data can be used to track the performance of the custom parsers during the indexing workflow.

## Step 6 - Registering Message Type Filters

The `cosmos-indexer` application allows for message type filters to be registered with the indexer. These filters are used to filter out transaction messages that should not be indexed. The `RegisterMessageTypeFilter` function on the `Indexer` instance is used to register message type filters.

This function expects a filter type that satisfies the [filter package's](https://github.com/DefiantLabs/cosmos-indexer/tree/main/filter) `MessageTypeFilter` interface. The filter type is used to filter out transaction messages that should not be indexed. There are a few built in Message Type Filters that can be used to filter out messages. Our example uses the `MessageTypeRegexFilter` to filter out messages that do not match the regex pattern.

```go
func main() {
	// ... previous code
	ibcRegexMessageTypeFilter, err := filter.NewRegexMessageTypeFilter("^/ibc.core.channel.v1.Msg(RecvPacket|Acknowledgement)$")
	if err != nil {
		log.Fatalf("Failed to create regex message type filter. Err: %v", err)
	}

	indexer.RegisterMessageTypeFilter(ibcRegexMessageTypeFilter)

	// ... rest of the main function
}
```

During execution of the indexer, the type URL of all messages will be checked against the regex pattern. If the type URL matches the pattern, the message will be indexed. If the type URL does not match the pattern, the message will be filtered out.

## Step 7 - Starting the Root Command

With all of the custom parsers and filters registered, the final step is to start the root command to begin the indexing workflow.

```go
func main(){
	// ... previous code
	err = cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
```
## Conclusion

This walkthrough has shown how to create a custom parser for IBC transactions. To see the full end-to-end implementation of this custom indexer, check out the [examples/ibc-patterns](https://github.com/DefiantLabs/cosmos-indexer/tree/main/examples/ibc-patterns) directory.
