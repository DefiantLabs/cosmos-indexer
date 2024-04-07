# Filtering

The application has built-in methods to allow building data-specific filters to modify what data is indexed by the database.

## Filtering Overview

There are currently 2 types of filters that will modify the behavior of the indexer at application runtime:

1. Block Event Filters - Filter the dataset for Block BeginBlocker and EndBlocker events
2. Transaction Message Type Filters - Filter the dataset for Transaction Messages

These filters are applied to the data returned by RPC requests for Block Events and Transactions and will include/exclude data based on the filter type.

### Writing Filters

Filters are created in a JSON configuration file that is loaded at runtime. These rules are then applied during execution of the main indexing loop to filter the indexed data.

To write filters:

1. Create a JSON file according to the filtering rules defined below
2. Pass the location of the JSON file at runtime to the flag `--base.filter-file` or in the config `.toml` file

They are loaded from the file, validated and then applied to all blocks.

## Block Event Filters Overview

Part of the indexed dataset are Block BeginBlock and EndBlock events. See [Block Events Indexed Data](../reference/block_events_indexed_data.md) for an overview of what data from the block is gathered, indexed and why.

Block events contain useful information on blockchain execution, but can significantly grow the size of the indexed dataset. For this reason, filtering mechanisms have been introduced to reduce the dataset according to some simple rules.

Block event filters are specified in the filter file in the following way:

```json
{
    "begin_block_filters": [...],
    "end_block_filters": [...]
}
```

The filters are applied to the specific block lifecycle event set.

### Filtering Rules for Block Events

Before you read this section, make sure you have read the [Block Events Indexed Data - Anatomy of a Block and Begin Block and End Block Events ](../reference/block_events_indexed_data.md#anatomy-of-a-block-and-begin-block-and-end-block-events) document so that you understand the shape of the data you will be writing filter rules for.

There are 4 types of filters currently provided by the application for block events:

1. Event type filters - applies a filter to the `event_type` field of the block event
2. Regex event type filters - same as above but uses a regular expression instead of an exact string match
3. Block event type and attribute filter - applies a filter to the event type and then searches the attributes to ensure it has a specific value as well
4. Rolling window filters - applies any number of the above filters to a window of events and includes all of them if all rules match

**Note**: Each filter configuration value has an associated `type` field that will identify it. This is used for loading the filter into the application at runtime and validating that it has the expected fields.

#### Event Type Filter

An event type filter applies an exact string match search to the block event to include or exclude it:

```json
{
    "type": "event_type",
    "event_type": "<event type to filter on>",
    "inclusive": <true or false>
}
```

#### Regex Event Type Filter

A regex event type filter applies a regular expression to the event type and matches if the regex matches. This is useful to reduce the amount of filters built to fulfill your application usage requirements:

```json
{
    "type": "regex_event_type",
    "event_type_regex": "<event type regex>",
    "inclusive": <true or false>
}
```

#### Block Event Type and Attribute Filter

A block event type and attribute filter applies an exact string match search to the event type and then searches the block event attributes for a specific attribute key/value pair:

```json
{
    "type": "event_type_and_attribute_value",
    "event_type": "<event type to filter on>",
    "attribute_key": "<attribute key to filter on>",
    "attribute_value": "<attribute value to filter on>"
    "inclusive": <true or false>
}
```

#### Rolling Window Filter

Sometimes it can be useful to filter for a set of events in a specific window of events. See [Block Events Indexed Data - Block Event Windows](../reference/block_events_indexed_data.md#block-event-windows) for details.

The application provides **Rolling Window Filters** that will allow you to filter for a number of events in a row for application-specific indexing requirements.

Rolling Window Filters apply `n` number of filters to a rolling window of `n` events. It loops through the entire array of events, applying the filters to each event as it goes. As soon as all filters match, the window is said to be matching and it is included or excluded from the indexed dataset.

Rolling window filters make use of the previous filter types:

```json
{
    "type": "rolling_window",
    "subfilters": [
        ...<n> number of the above filters
    ],
    "inclusive": <true or false>
}
```

The algorithm for rolling window event matching is a very simple brute force search that fails early as soon as an event does not match, continuing on to the next window.

Given an array of block events:

```json
[a, b, c, d]
```

And a rolling window subfilter configuration:

```
[r1, r2]
```

Run the filters on the array like so:

Loop 1:

Does `r1` match `a` and does `r2` match `b`? If so, both `a` and `b` are considered a match.

Loop 2:

Does `r1` match `b` and does `r2` match `c`? If so, both `b` and `c` are considered a match.

And so on.

## Transaction Message Filters Overview

Part of the indexed dataset are Transactions and the Messages that are executed in them. See [Transactions Indexed Data](../reference/transactions_indexed_data.md) for an overview of what data from the block is gathered, indexed and why.

Transactions filters apply to the messages that were included in the transaction. When a transaction message matches, the message data will be included in the indexed dataset.

Message type filters are specified in the filter file in the following way:

```json
{
    "message_type_filters": [...]
}
```

The filters are applied to the specific transaction message set.

### Filtering Rules for Transaction Messages

Before you read this section, make sure you have read the [Transactions Indexed Data - Anatomy of a Transaction and Messages](../reference/transactions_indexed_data.md#anatomy-of-a-transaction-and-messages) document so that you understand the shape of the data you will be writing filter rules for.

There are 2 types of filters currently provided by the application for transaction messages:

1. Message type filters - applies a filter to the `type_url` field of the transaction message
2. Regex message type filters - same as above but uses a regular expression instead of an exact string match

**Note**: Each filter configuration value has an associated `type` field that will identify it. This is used for loading the filter into the application at runtime and validating that it has the expected fields.

#### Message Type Filter

A message type filter applies an exact string match search to the transaction message to include or exclude it:

```json
{
    "type": "message_type",
    "message_type": "<message type to filter on>"
}
```

#### Regex Message Type Filter

A regex message type filter applies a regular expression to the message type and matches if the regex matches. This is useful to reduce the amount of filters built to fulfill your application usage requirements:

```json
{
    "type": "regex_message_type",
    "message_type_regex": "<message type regex>"
}
```

## Example Filter Configuration

Here is an example filter configuration file that includes all of the filter types:

```json
{
    "begin_block_filters": [
        {
            "type": "event_type",
            "event_type": "coin_received",
            "inclusive": true
        },
        {
            "type": "regex_event_type",
            "event_type_regex": "coin_.*", // matches coin_received and coin_spent
            "inclusive": false
        },
        {
            "type": "event_type_and_attribute_value",
            "event_type": "coin_received",
            "attribute_key": "receiver",
            "attribute_value": "cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q",
            "inclusive": true
        },
        {
            "type": "rolling_window",
            "subfilters": [
                {
                    "type": "event_type",
                    "event_type": "coin_received",
                    "inclusive": true
                },
                {
                    "type": "event_type_and_attribute_value",
                    "event_type": "coin_received",
                    "attribute_key": "receiver",
                    "attribute_value": "cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q",
                    "inclusive": true
                }
            ],
            "inclusive": true
        }
    ],
    "end_block_filters": [
        {
            "type": "event_type",
            "event_type": "coin_received",
            "inclusive": true
        },
        {
            "type": "regex_event_type",
            "event_type_regex": ".*",
            "inclusive": false
        },
        {
            "type": "event_type_and_attribute_value",
            "event_type": "coin_received",
            "attribute_key": "receiver",
            "attribute_value": "cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q",
            "inclusive": true
        },
        {
            "type": "rolling_window",
            "subfilters": [
                {
                    "type": "event_type",
                    "event_type": "coin_received",
                    "inclusive": true
                },
                {
                    "type": "event_type_and_attribute_value",
                    "event_type": "coin_received",
                    "attribute_key": "receiver",
                    "attribute_value": "cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q",
                    "inclusive": true
                }
            ],
            "inclusive": true
        }
    ],
    "message_type_filters": [
        {
            "type": "message_type",
            "message_type": "/cosmos.bank.v1beta1.MsgSend"
        },
        {
            "type": "regex_message_type",
            "message_type_regex": "/cosmos\\.gov.*" // matches all gov messages
        }
    ]
}
```
                    
