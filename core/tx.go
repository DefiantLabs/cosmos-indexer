package core

import (
	b64 "encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/DefiantLabs/cosmos-indexer/config"
	txtypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/util"
	"github.com/DefiantLabs/probe/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	cosmosTx "github.com/cosmos/cosmos-sdk/types/tx"
	"gorm.io/gorm"
)

// Unmarshal JSON to a particular type. There can be more than one handler for each type.
// TODO: Remove this map and replace with a more generic solution
var messageTypeHandler = map[string][]func() txtypes.CosmosMessage{}

// var messageTypeIgnorer = map[string]interface{}{}

// Merge the chain specific message type handlers into the core message type handler map.
// Chain specific handlers will be registered BEFORE any generic handlers.
// TODO: Remove this function and replace with a more generic solution
func ChainSpecificMessageTypeHandlerBootstrap(chainID string) {
	var chainSpecificMessageTpeHandler map[string][]func() txtypes.CosmosMessage
	for key, value := range chainSpecificMessageTpeHandler {
		if list, ok := messageTypeHandler[key]; ok {
			messageTypeHandler[key] = append(value, list...)
		} else {
			messageTypeHandler[key] = value
		}
	}
}

func toAttributes(attrs []types.Attribute) []txtypes.Attribute {
	list := []txtypes.Attribute{}
	for _, attr := range attrs {
		lma := txtypes.Attribute{Key: attr.Key, Value: attr.Value}
		list = append(list, lma)
	}

	return list
}

func toEvents(msgEvents types.StringEvents) (list []txtypes.LogMessageEvent) {
	for _, evt := range msgEvents {
		lme := txtypes.LogMessageEvent{Type: evt.Type, Attributes: toAttributes(evt.Attributes)}
		list = append(list, lme)
	}

	return list
}

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func ProcessRPCBlockByHeightTXs(db *gorm.DB, cl *client.ChainClient, blockResults *coretypes.ResultBlock, resultBlockRes *coretypes.ResultBlockResults) ([]dbTypes.TxDBWrapper, *time.Time, error) {
	if len(blockResults.Block.Txs) != len(resultBlockRes.TxsResults) {
		config.Log.Fatalf("blockResults & resultBlockRes: different length")
	}

	blockTime := &blockResults.Block.Time
	blockTimeStr := blockTime.Format(time.RFC3339)
	currTxDbWrappers := make([]dbTypes.TxDBWrapper, len(blockResults.Block.Txs))

	for txIdx, tendermintTx := range blockResults.Block.Txs {
		txResult := resultBlockRes.TxsResults[txIdx]

		// Indexer types only used by the indexer app (similar to the cosmos types)
		var indexerMergedTx txtypes.MergedTx
		var indexerTx txtypes.IndexerTx
		var txBody txtypes.Body
		var currMessages []types.Msg
		var currLogMsgs []txtypes.LogMessage

		txDecoder := cl.Codec.TxConfig.TxDecoder()
		txBasic, err := txDecoder(tendermintTx)
		if err != nil {
			return nil, blockTime, fmt.Errorf("ProcessRPCBlockByHeightTXs: TX cannot be parsed from block %v. Err: %v", blockResults.Block.Height, err)
		}

		// This is a hack, but as far as I can tell necessary. "wrapper" struct is private in Cosmos SDK.
		field := reflect.ValueOf(txBasic).Elem().FieldByName("tx")
		iTx := getUnexportedField(field)
		txFull := iTx.(*cosmosTx.Tx)
		logs := types.ABCIMessageLogs{}

		// Failed TXs do not have proper JSON in the .Log field, causing ParseABCILogs to fail to unmarshal the logs
		// We can entirely ignore failed TXs in downstream parsers, because according to the Cosmos specification, a single failed message in a TX fails the whole TX
		if txResult.Code == 0 {
			logs, err = types.ParseABCILogs(txResult.Log)
		} else {
			err = nil
		}

		if err != nil {
			return nil, blockTime, fmt.Errorf("logs could not be parsed")
		}

		// Get the Messages and Message Logs
		for msgIdx, currMsg := range txFull.GetMsgs() {
			if currMsg != nil {
				currMessages = append(currMessages, currMsg)
				msgEvents := types.StringEvents{}
				if txResult.Code == 0 {
					msgEvents = logs[msgIdx].Events
				}

				currTxLog := txtypes.LogMessage{
					MessageIndex: msgIdx,
					Events:       toEvents(msgEvents),
				}
				currLogMsgs = append(currLogMsgs, currTxLog)
			} else {
				return nil, blockTime, fmt.Errorf("tx message could not be processed")
			}
		}

		txBody.Messages = currMessages
		indexerTx.Body = txBody
		txHash := tendermintTx.Hash()
		indexerTxResp := txtypes.Response{
			TxHash:    b64.StdEncoding.EncodeToString(txHash),
			Height:    fmt.Sprintf("%d", blockResults.Block.Height),
			TimeStamp: blockTimeStr,
			RawLog:    txResult.Log,
			Log:       currLogMsgs,
			Code:      txResult.Code,
		}

		indexerTx.AuthInfo = *txFull.AuthInfo
		indexerTx.Signers = txFull.GetSigners()
		indexerMergedTx.TxResponse = indexerTxResp
		indexerMergedTx.Tx = indexerTx
		indexerMergedTx.Tx.AuthInfo = *txFull.AuthInfo

		processedTx, _, err := ProcessTx(db, indexerMergedTx)
		if err != nil {
			return currTxDbWrappers, blockTime, err
		}

		processedTx.SignerAddress = models.Address{Address: txFull.FeePayer().String()}
		currTxDbWrappers[txIdx] = processedTx
	}

	return currTxDbWrappers, blockTime, nil
}

// ProcessRPCTXs - Given an RPC response, build out the more specific data used by the parser.
func ProcessRPCTXs(db *gorm.DB, txEventResp *cosmosTx.GetTxsEventResponse) ([]dbTypes.TxDBWrapper, *time.Time, error) {
	currTxDbWrappers := make([]dbTypes.TxDBWrapper, len(txEventResp.Txs))
	var blockTime *time.Time

	for txIdx := range txEventResp.Txs {
		// Indexer types only used by the indexer app (similar to the cosmos types)
		var indexerMergedTx txtypes.MergedTx
		var indexerTx txtypes.IndexerTx
		var txBody txtypes.Body
		var currMessages []types.Msg
		var currLogMsgs []txtypes.LogMessage
		currTx := txEventResp.Txs[txIdx]
		currTxResp := txEventResp.TxResponses[txIdx]

		// Get the Messages and Message Logs
		for msgIdx := range currTx.Body.Messages {
			currMsg := currTx.Body.Messages[msgIdx].GetCachedValue()
			if currMsg != nil {
				msg := currMsg.(types.Msg)
				currMessages = append(currMessages, msg)
				if len(currTxResp.Logs) >= msgIdx+1 {
					msgEvents := currTxResp.Logs[msgIdx].Events
					currTxLog := txtypes.LogMessage{
						MessageIndex: msgIdx,
						Events:       toEvents(msgEvents),
					}
					currLogMsgs = append(currLogMsgs, currTxLog)
				}
			} else {
				return nil, blockTime, fmt.Errorf("tx message could not be processed. CachedValue is not present. TX Hash: %s, Msg type: %s, Msg index: %d, Code: %d",
					currTxResp.TxHash,
					currTx.Body.Messages[msgIdx].TypeUrl,
					msgIdx,
					currTxResp.Code,
				)
			}
		}

		txBody.Messages = currMessages
		indexerTx.Body = txBody

		indexerTxResp := txtypes.Response{
			TxHash:    currTxResp.TxHash,
			Height:    fmt.Sprintf("%d", currTxResp.Height),
			TimeStamp: currTxResp.Timestamp,
			RawLog:    currTxResp.RawLog,
			Log:       currLogMsgs,
			Code:      currTxResp.Code,
		}

		indexerTx.AuthInfo = *currTx.AuthInfo
		indexerTx.Signers = currTx.GetSigners()
		indexerMergedTx.TxResponse = indexerTxResp
		indexerMergedTx.Tx = indexerTx
		indexerMergedTx.Tx.AuthInfo = *currTx.AuthInfo

		processedTx, txTime, err := ProcessTx(db, indexerMergedTx)
		if err != nil {
			return currTxDbWrappers, blockTime, err
		}

		if blockTime == nil {
			blockTime = &txTime
		}

		processedTx.SignerAddress = models.Address{Address: currTx.FeePayer().String()}
		currTxDbWrappers[txIdx] = processedTx
	}

	return currTxDbWrappers, blockTime, nil
}

func ProcessTx(db *gorm.DB, tx txtypes.MergedTx) (txDBWapper dbTypes.TxDBWrapper, txTime time.Time, err error) {
	txTime, err = time.Parse(time.RFC3339, tx.TxResponse.TimeStamp)
	if err != nil {
		config.Log.Error("Error parsing tx timestamp.", err)
		return
	}

	code := tx.TxResponse.Code

	var messages []dbTypes.MessageDBWrapper

	// non-zero code means the Tx was unsuccessful. We will still need to account for fees in both cases though.
	uniqueEventTypes := make(map[string]models.MessageEventType)
	uniqueEventAttributeKeys := make(map[string]models.MessageEventAttributeKey)
	if code == 0 {
		for messageIndex, message := range tx.Tx.Body.Messages {
			var currMessage models.Message
			var currMessageType models.MessageType
			currMessage.MessageIndex = messageIndex

			// Get the message log that corresponds to the current message
			var currMessageDBWrapper dbTypes.MessageDBWrapper
			messageLog := txtypes.GetMessageLogForIndex(tx.TxResponse.Log, messageIndex)

			currMessageType.MessageType = types.MsgTypeURL(message)
			currMessage.MessageType = currMessageType
			currMessageDBWrapper.Message = currMessage

			for eventIndex, event := range messageLog.Events {
				uniqueEventTypes[event.Type] = models.MessageEventType{Type: event.Type}

				var currMessageEvent dbTypes.MessageEventDBWrapper
				currMessageEvent.MessageEvent = models.MessageEvent{
					MessageEventType: uniqueEventTypes[event.Type],
					Index:            uint64(eventIndex),
				}
				var currMessageEventAttributes []models.MessageEventAttribute
				for attributeIndex, attribute := range event.Attributes {
					uniqueEventAttributeKeys[attribute.Key] = models.MessageEventAttributeKey{Key: attribute.Key}

					currMessageEventAttributes = append(currMessageEventAttributes, models.MessageEventAttribute{
						Value:                    attribute.Value,
						MessageEventAttributeKey: uniqueEventAttributeKeys[attribute.Key],
						Index:                    uint64(attributeIndex),
					})
				}

				currMessageEvent.Attributes = currMessageEventAttributes
				currMessageDBWrapper.MessageEvents = append(currMessageDBWrapper.MessageEvents, currMessageEvent)
			}

			config.Log.Debug(fmt.Sprintf("[Block: %v] Found msg of type '%v'.", tx.TxResponse.Height, currMessageType.MessageType))

			messages = append(messages, currMessageDBWrapper)
		}
	}

	fees, err := ProcessFees(db, tx.Tx.AuthInfo, tx.Tx.Signers)
	if err != nil {
		return txDBWapper, txTime, err
	}

	txDBWapper.Tx = models.Tx{Hash: tx.TxResponse.TxHash, Fees: fees, Code: code}
	txDBWapper.Messages = messages
	txDBWapper.UniqueMessageAttributeKeys = uniqueEventAttributeKeys
	txDBWapper.UniqueMessageEventTypes = uniqueEventTypes

	return txDBWapper, txTime, nil
}

// ProcessFees returns a comma delimited list of fee amount/denoms
func ProcessFees(db *gorm.DB, authInfo cosmosTx.AuthInfo, signers []types.AccAddress) ([]models.Fee, error) {
	feeCoins := authInfo.Fee.Amount
	payer := authInfo.Fee.GetPayer()
	fees := []models.Fee{}

	for _, coin := range feeCoins {
		zeroFee := big.NewInt(0)

		// There are chains like Osmosis that do not require TX fees for certain TXs
		if zeroFee.Cmp(coin.Amount.BigInt()) != 0 {
			amount := util.ToNumeric(coin.Amount.BigInt())
			denom, err := dbTypes.GetDenomForBase(coin.Denom)
			if err != nil {
				// attempt to add missing denoms to the database
				config.Log.Warnf("Denom lookup failed. Will be inserted as UNKNOWN. Denom Received: %v. Err: %v", coin.Denom, err)
				denom, err = dbTypes.AddUnknownDenom(db, coin.Denom)
				if err != nil {
					config.Log.Error(fmt.Sprintf("There was an error adding a missing denom. Denom: %v", coin.Denom), err)
					return nil, err
				}
			}
			payerAddr := models.Address{}
			if payer != "" {
				payerAddr.Address = payer
			} else {
				if authInfo.SignerInfos[0].PublicKey == nil && len(signers) > 0 {
					payerAddr.Address = signers[0].String()
				} else {
					var pubKey cryptoTypes.PubKey
					cpk := authInfo.SignerInfos[0].PublicKey.GetCachedValue()

					// if this is a multisig msg, handle it specially
					if strings.Contains(authInfo.SignerInfos[0].ModeInfo.GetMulti().String(), "mode:SIGN_MODE_LEGACY_AMINO_JSON") {
						pubKey = cpk.(*multisig.LegacyAminoPubKey).GetPubKeys()[0]
					} else {
						pubKey = cpk.(cryptoTypes.PubKey)
					}
					hexPub := hex.EncodeToString(pubKey.Bytes())
					bechAddr, err := ParseSignerAddress(hexPub, "")
					if err != nil {
						config.Log.Error(fmt.Sprintf("Error parsing signer address '%v' for tx.", hexPub), err)
					} else {
						payerAddr.Address = bechAddr
					}
				}
			}

			fees = append(fees, models.Fee{Amount: amount, Denomination: denom, PayerAddress: payerAddr})
		}
	}

	return fees, nil
}

// getDenom handles denom processing for both IBC denoms and native denoms.
// If the denom begins with ibc/ we know this is an IBC denom trace, and it's not guaranteed there is an entry in
// the Denom table.
// func getDenom(denom string) (dbTypes.Denom, error) {
// 	var (
// 		denomSent dbTypes.Denom
// 		err       error
// 	)

// 	// if this is an ibc denom trace, get the ibc denom then use the base denom to get the Denom from the db
// 	if strings.HasPrefix(denom, "ibc/") {
// 		ibcDenom, err := dbTypes.GetIBCDenom(denom)
// 		if err != nil {
// 			config.Log.Warnf("IBC Denom lookup failed for  %s, err: %v", denom, err)
// 		} else {
// 			denomSent, err = dbTypes.GetDenomForBase(ibcDenom.BaseDenom)
// 			if err != nil {
// 				config.Log.Warnf("Denom lookup failed for IBC base denom %s, err: %v", ibcDenom.BaseDenom, err)
// 				return dbTypes.Denom{Base: ibcDenom.BaseDenom}, err
// 			}
// 		}
// 	}

// 	// if this is not an ibc denom trace or there was an issue querying the ibc denom trace in the other table,
// 	// attempt to look up this denom in the regular Denom table
// 	if denomSent.Base == "" {
// 		denomSent, err = dbTypes.GetDenomForBase(denom)
// 		if err != nil {
// 			return dbTypes.Denom{Base: denom}, err
// 		}
// 	}

// 	return denomSent, nil
// }
