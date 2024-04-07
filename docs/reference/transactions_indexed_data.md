# Transactions Indexed Data

The application indexes Transactions and the Messages that are executed in them into a well-structured data shape. In this section, you will find an overview of what Transactions and Messages are and how they are indexed by the application.

## Anatomy of a Transaction and its Messages

### Transactions

In Cosmos, every block has a list of transactions that are executed. Each transaction has any number of messages attached that define the actions that are executed in the transaction.

When transactions for a block are requested through RPC, the returned dataset has the following shape (from the GetTxsEvent RPC service endpoint):

```json
{
    "txs": [
        "body": {
            "messages": [
                {
                    "type_url": "<message type url>",
                    "value": "<protobuf encoded message>"
                },
                ...<more messages>
            ]
        }
    ],
    "tx_responses": [
        {
            "code": "<response code>",
            "logs": [<event logs>],
        },
        }
    ]
}
```

Each item in the `txs` array is a transaction that was executed in the block. Each transaction has a `body` field that contains the messages that were executed in the transaction.

The `tx_responses` array contains the response data for each transaction. The `code` field contains the response code for the transaction and the `logs` field contains the event logs that were emitted during the transaction execution.

### Transaction Messages

Transaction messages have the following data shape:

```json
{
    "type_url": "<message type url>",
    "value": "<protobuf encoded message>"
}
```

Each message has a `type_url` field that indicates the type of message that was executed. The `value` field contains the protobuf encoded message, which contains message-specific data.
