package core

import (
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"reflect"
	"strings"
	"time"
	"unsafe"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	cosmosTx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/nodersteam/cosmos-indexer/config"
	txtypes "github.com/nodersteam/cosmos-indexer/cosmos/modules/tx"
	dbTypes "github.com/nodersteam/cosmos-indexer/db"
	"github.com/nodersteam/cosmos-indexer/db/models"
	"github.com/nodersteam/cosmos-indexer/filter"
	"github.com/nodersteam/cosmos-indexer/util"
	"github.com/nodersteam/probe/client"
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

func ProcessRPCBlockByHeightTXs(db *gorm.DB, cl *client.ChainClient, messageTypeFilters []filter.MessageTypeFilter, blockResults *coretypes.ResultBlock, resultBlockRes *coretypes.ResultBlockResults) ([]dbTypes.TxDBWrapper, *time.Time, error) {
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
		var txFull *cosmosTx.Tx
		if err != nil {
			txBasic, err = InAppTxDecoder(cl.Codec)(tendermintTx)
			if err != nil {
				return nil, blockTime, fmt.Errorf("ProcessRPCBlockByHeightTXs: TX cannot be parsed from block %v. This is usually a proto definition error. Err: %v", blockResults.Block.Height, err)
			}
			txFull = txBasic.(*cosmosTx.Tx)
		} else {
			// This is a hack, but as far as I can tell necessary. "wrapper" struct is private in Cosmos SDK.
			field := reflect.ValueOf(txBasic).Elem().FieldByName("tx")
			iTx := getUnexportedField(field)
			txFull = iTx.(*cosmosTx.Tx)
		}

		logs := types.ABCIMessageLogs{}

		// Failed TXs do not have proper JSON in the .Log field, causing ParseABCILogs to fail to unmarshal the logs
		// We can entirely ignore failed TXs in downstream parsers, because according to the Cosmos specification, a single failed message in a TX fails the whole TX
		if txResult.Code == 0 {
			logs, err = types.ParseABCILogs(txResult.Log)
			if err != nil {
				return nil, blockTime, fmt.Errorf("logs could not be parsed")
			}
		}

		txHash := tendermintTx.Hash()

		var messagesRaw [][]byte

		// Get the Messages and Message Logs
		for msgIdx := range txFull.Body.Messages {
			shouldIndex, err := messageTypeShouldIndex(txFull.Body.Messages[msgIdx].TypeUrl, messageTypeFilters)
			if err != nil {
				config.Log.Error("messageTypeShouldIndex", err)
				return nil, blockTime, err
			}

			if !shouldIndex {
				config.Log.Debug(fmt.Sprintf("[Block: %v] [TX: %v] Skipping msg of type '%v'.",
					blockResults.Block.Height, tendermintHashToHex(txHash), txFull.Body.Messages[msgIdx].TypeUrl))
				currMessages = append(currMessages, nil)
				currLogMsgs = append(currLogMsgs, txtypes.LogMessage{
					MessageIndex: msgIdx,
				})
				messagesRaw = append(messagesRaw, nil)
				continue
			}

			currMsg := txFull.Body.Messages[msgIdx].GetCachedValue()

			if currMsg != nil {
				msg := currMsg.(types.Msg)
				messagesRaw = append(messagesRaw, txFull.Body.Messages[msgIdx].Value)
				currMessages = append(currMessages, msg)
				msgEvents := types.StringEvents{}
				if txResult.Code == 0 {
					msgEvents = logs[msgIdx].Events
				}

				currTxLog := txtypes.LogMessage{
					MessageIndex: msgIdx,
					Events:       toEvents(msgEvents),
				}
				currLogMsgs = append(currLogMsgs, currTxLog)
			}
			// TODO understand why it's not working here
			//else {
			//	return nil, blockTime, fmt.Errorf("tx message could not be processed")
			//}
		}

		txBody.Messages = currMessages
		indexerTx.Body = txBody

		indexerTxResp := txtypes.Response{
			TxHash:    tendermintHashToHex(txHash),
			Height:    fmt.Sprintf("%d", blockResults.Block.Height),
			TimeStamp: blockTimeStr,
			RawLog:    []byte(txResult.Log),
			Log:       currLogMsgs,
			Code:      txResult.Code,
			GasUsed:   txResult.GasUsed,
			GasWanted: txResult.GasWanted,
			Codespace: txResult.Codespace,
			Info:      txResult.Info,
			//Data:      string(txResult.Data), TODO
		}

		indexerTx.AuthInfo = *txFull.AuthInfo
		indexerMergedTx.TxResponse = indexerTxResp
		indexerMergedTx.Tx = indexerTx
		indexerMergedTx.Tx.AuthInfo = *txFull.AuthInfo

		processedTx, _, err := ProcessTx(db, indexerMergedTx, messagesRaw)
		if err != nil {
			config.Log.Error("ProcessTx", err)
			return currTxDbWrappers, blockTime, err
		}

		filteredSigners := []types.AccAddress{}
		for _, filteredMessage := range txBody.Messages {
			if filteredMessage != nil {
				filteredSigners = append(filteredSigners, filteredMessage.GetSigners()...)
			}
		}

		signers, signerInfos, err := ProcessSigners(cl, txFull.AuthInfo, filteredSigners)
		if err != nil {
			config.Log.Error("ProcessSigners", err)
			return currTxDbWrappers, blockTime, err
		}

		processedTx.Tx.SignerAddresses = signers

		fees, err := ProcessFees(db, indexerTx.AuthInfo, signers)
		if err != nil {
			config.Log.Error("ProcessFees", err)
			return currTxDbWrappers, blockTime, err
		}

		processedTx.Tx.Fees = fees

		// extra fields
		processedTx.Tx.Signatures = txFull.Signatures
		processedTx.Tx.Memo = txFull.Body.Memo
		processedTx.Tx.TimeoutHeight = txFull.Body.TimeoutHeight

		extensionOptions := make([]string, 0)
		for _, opt := range txFull.Body.ExtensionOptions {
			extensionOptions = append(extensionOptions, opt.String())
		}
		processedTx.Tx.ExtensionOptions = extensionOptions

		nonExtensionOptions := make([]string, 0)
		for _, opt := range txFull.Body.NonCriticalExtensionOptions {
			extensionOptions = append(extensionOptions, opt.String())
		}
		processedTx.Tx.NonCriticalExtensionOptions = nonExtensionOptions
		processedTx.Tx.TxResponse = models.TxResponse{
			TxHash:    indexerTxResp.TxHash,
			Height:    indexerTxResp.Height,
			TimeStamp: indexerTxResp.TimeStamp,
			Code:      indexerTxResp.Code,
			RawLog:    indexerTxResp.RawLog,
			GasUsed:   indexerTxResp.GasUsed,
			GasWanted: indexerTxResp.GasWanted,
			Codespace: indexerTxResp.Codespace,
			Data:      indexerTxResp.Data,
			Info:      indexerTxResp.Info,
		}

		if txFull.AuthInfo != nil && txFull.AuthInfo.Fee != nil {
			txAuthInfo := models.AuthInfo{
				Fee: models.AuthInfoFee{
					Granter:  txFull.AuthInfo.Fee.Granter,
					Payer:    txFull.AuthInfo.Fee.Payer,
					GasLimit: txFull.AuthInfo.Fee.GasLimit,
				},
				SignerInfos: signerInfos,
			}
			if txFull.AuthInfo.Tip != nil {
				tipAmount := make([]models.TipAmount, 0)
				for _, a := range txFull.AuthInfo.Tip.Amount {
					tipAmount = append(tipAmount, models.TipAmount{
						Denom:  a.Denom,
						Amount: decimal.NewFromInt(a.Amount.Int64()),
					})
				}
				txAuthInfo.Tip = models.Tip{
					Tipper: txFull.AuthInfo.Tip.Tipper,
					Amount: tipAmount,
				}
			}

			processedTx.Tx.AuthInfo = txAuthInfo
		}

		currTxDbWrappers[txIdx] = processedTx
	}

	return currTxDbWrappers, blockTime, nil
}

func tendermintHashToHex(hash []byte) string {
	return strings.ToUpper(hex.EncodeToString(hash))
}

// ProcessRPCTXs - Given an RPC response, build out the more specific data used by the parser.
func ProcessRPCTXs(db *gorm.DB, cl *client.ChainClient, messageTypeFilters []filter.MessageTypeFilter, txEventResp *cosmosTx.GetTxsEventResponse) ([]dbTypes.TxDBWrapper, *time.Time, error) {
	currTxDbWrappers := make([]dbTypes.TxDBWrapper, len(txEventResp.Txs))
	var blockTime *time.Time

	for txIdx := range txEventResp.Txs {
		// Indexer types only used by the indexer app (similar to the cosmos types)
		var indexerMergedTx txtypes.MergedTx
		var indexerTx txtypes.IndexerTx
		var txBody txtypes.Body
		var currMessages []types.Msg
		var currLogMsgs []txtypes.LogMessage
		var messagesRaw [][]byte

		currTx := txEventResp.Txs[txIdx]
		currTxResp := txEventResp.TxResponses[txIdx]

		// Get the Messages and Message Logs
		for msgIdx := range currTx.Body.Messages {

			shouldIndex, err := messageTypeShouldIndex(currTx.Body.Messages[msgIdx].TypeUrl, messageTypeFilters)
			if err != nil {
				return nil, blockTime, err
			}

			if !shouldIndex {
				config.Log.Debug(fmt.Sprintf("[Block: %v] [TX: %v] Skipping msg of type '%v'.", currTxResp.Height, currTxResp.TxHash, currTx.Body.Messages[msgIdx].TypeUrl))
				currMessages = append(currMessages, nil)
				currLogMsgs = append(currLogMsgs, txtypes.LogMessage{
					MessageIndex: msgIdx,
				})
				messagesRaw = append(messagesRaw, nil)
				continue
			}

			currMsg := currTx.Body.Messages[msgIdx].GetCachedValue()
			messagesRaw = append(messagesRaw, currTx.Body.Messages[msgIdx].Value)

			// If we reached here, unpacking the entire TX raw was not successful
			// Attempt to unpack the message individually.
			if currMsg == nil {
				var currMsgUnpack types.Msg
				err := cl.Codec.InterfaceRegistry.UnpackAny(currTx.Body.Messages[msgIdx], &currMsgUnpack)
				if err != nil || currMsgUnpack == nil {
					return nil, blockTime, fmt.Errorf("tx message could not be processed. Unpacking protos failed and CachedValue is not present. TX Hash: %s, Msg type: %s, Msg index: %d, Code: %d",
						currTxResp.TxHash,
						currTx.Body.Messages[msgIdx].TypeUrl,
						msgIdx,
						currTxResp.Code,
					)
				}
				currMsg = currMsgUnpack
			}

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
			}
		}

		txBody.Messages = currMessages
		indexerTx.Body = txBody

		indexerTxResp := txtypes.Response{
			TxHash:    currTxResp.TxHash,
			Height:    fmt.Sprintf("%d", currTxResp.Height),
			TimeStamp: currTxResp.Timestamp,
			RawLog:    []byte(currTxResp.RawLog),
			Log:       currLogMsgs,
			Code:      currTxResp.Code,
			GasUsed:   currTxResp.GasUsed,
			GasWanted: currTxResp.GasWanted,
			Info:      currTxResp.Info,
			Data:      currTxResp.Data,
		}

		indexerTx.AuthInfo = *currTx.AuthInfo
		indexerMergedTx.TxResponse = indexerTxResp
		indexerMergedTx.Tx = indexerTx
		indexerMergedTx.Tx.AuthInfo = *currTx.AuthInfo

		processedTx, txTime, err := ProcessTx(db, indexerMergedTx, messagesRaw)
		if err != nil {
			return currTxDbWrappers, blockTime, err
		}

		if blockTime == nil {
			blockTime = &txTime
		}

		filteredSigners := make([]types.AccAddress, 0)
		for _, filteredMessage := range txBody.Messages {
			if filteredMessage != nil {
				filteredSigners = append(filteredSigners, filteredMessage.GetSigners()...)
			}
		}

		err = currTx.AuthInfo.UnpackInterfaces(cl.Codec.InterfaceRegistry)
		if err != nil {
			return currTxDbWrappers, blockTime, err
		}

		signers, signerInfos, err := ProcessSigners(cl, currTx.AuthInfo, filteredSigners)
		if err != nil {
			return currTxDbWrappers, blockTime, err
		}
		processedTx.Tx.SignerAddresses = signers

		fees, err := ProcessFees(db, indexerTx.AuthInfo, signers)
		if err != nil {
			return currTxDbWrappers, blockTime, err
		}

		processedTx.Tx.Fees = fees

		// extra fields
		processedTx.Tx.Signatures = currTx.Signatures
		processedTx.Tx.Memo = currTx.Body.Memo
		processedTx.Tx.TimeoutHeight = currTx.Body.TimeoutHeight

		extensionOptions := make([]string, 0)
		for _, opt := range currTx.Body.ExtensionOptions {
			extensionOptions = append(extensionOptions, opt.String())
		}
		processedTx.Tx.ExtensionOptions = extensionOptions

		nonExtensionOptions := make([]string, 0)
		for _, opt := range currTx.Body.NonCriticalExtensionOptions {
			extensionOptions = append(extensionOptions, opt.String())
		}
		processedTx.Tx.NonCriticalExtensionOptions = nonExtensionOptions
		processedTx.Tx.TxResponse = models.TxResponse{
			TxHash:    indexerTxResp.TxHash,
			Height:    indexerTxResp.Height,
			TimeStamp: indexerTxResp.TimeStamp,
			Code:      indexerTxResp.Code,
			RawLog:    indexerTxResp.RawLog,
			GasUsed:   indexerTxResp.GasUsed,
			GasWanted: indexerTxResp.GasWanted,
			Codespace: indexerTxResp.Codespace,
		}

		if currTx.AuthInfo != nil {
			txAuthInfo := models.AuthInfo{
				Fee: models.AuthInfoFee{
					Granter:  currTx.AuthInfo.Fee.Granter,
					Payer:    currTx.AuthInfo.Fee.Payer,
					GasLimit: currTx.AuthInfo.Fee.GasLimit,
				},
				SignerInfos: signerInfos,
			}
			if currTx.AuthInfo.Tip != nil {
				tipAmount := make([]models.TipAmount, 0)
				for _, a := range currTx.AuthInfo.Tip.Amount {
					tipAmount = append(tipAmount, models.TipAmount{
						Denom:  a.Denom,
						Amount: decimal.NewFromInt(a.Amount.Int64()),
					})
				}
				txAuthInfo.Tip = models.Tip{
					Tipper: currTx.AuthInfo.Tip.Tipper,
					Amount: tipAmount,
				}
			}

			processedTx.Tx.AuthInfo = txAuthInfo
		}

		currTxDbWrappers[txIdx] = processedTx
	}

	return currTxDbWrappers, blockTime, nil
}

func messageTypeShouldIndex(messageType string, filters []filter.MessageTypeFilter) (bool, error) {
	if len(filters) != 0 {
		filterData := filter.MessageTypeData{
			MessageType: messageType,
		}

		matches := false
		for _, messageTypeFilter := range filters {
			typeMatch, err := messageTypeFilter.MessageTypeMatches(filterData)
			if err != nil {
				return false, err
			}
			if typeMatch {
				matches = true
				break
			}
		}

		return matches, nil
	}

	return true, nil
}

func ProcessTx(db *gorm.DB, tx txtypes.MergedTx, messagesRaw [][]byte) (txDBWapper dbTypes.TxDBWrapper, txTime time.Time, err error) {
	txTime, err = time.Parse(time.RFC3339, tx.TxResponse.TimeStamp)
	if err != nil {
		config.Log.Error("Error parsing tx timestamp.", err)
		return txDBWapper, txTime, err
	}

	code := tx.TxResponse.Code

	var messages []dbTypes.MessageDBWrapper

	uniqueMessageTypes := make(map[string]models.MessageType)
	uniqueEventTypes := make(map[string]models.MessageEventType)
	uniqueEventAttributeKeys := make(map[string]models.MessageEventAttributeKey)
	// non-zero code means the Tx was unsuccessful. We will still need to account for fees in both cases though.
	if code == 0 {
		for messageIndex, message := range tx.Tx.Body.Messages {
			if message != nil {
				messageType, currMessageDBWrapper := ProcessMessage(messageIndex, message,
					tx.TxResponse.Log, uniqueEventTypes, uniqueEventAttributeKeys)
				currMessageDBWrapper.Message.MessageBytes = messagesRaw[messageIndex]
				uniqueMessageTypes[messageType] = currMessageDBWrapper.Message.MessageType
				config.Log.Debug(fmt.Sprintf("[Block: %v] [TX: %v] Found msg of type '%v'.", tx.TxResponse.Height, tx.TxResponse.TxHash, messageType))
				messages = append(messages, currMessageDBWrapper)
			}
		}
	}

	txDBWapper.Tx = models.Tx{Hash: tx.TxResponse.TxHash, Code: code, Timestamp: txTime}
	txDBWapper.Messages = messages
	txDBWapper.UniqueMessageTypes = uniqueMessageTypes
	txDBWapper.UniqueMessageAttributeKeys = uniqueEventAttributeKeys
	txDBWapper.UniqueMessageEventTypes = uniqueEventTypes

	return txDBWapper, txTime, nil
}

// ProcessSigners in a deterministic order.
// 1. Processes signers from the auth info
// 2. Processes signers from the signers array
// 3. Processes the fee payer
func ProcessSigners(cl *client.ChainClient, authInfo *cosmosTx.AuthInfo, messageSigners []types.AccAddress) ([]models.Address, []*models.SignerInfo, error) {
	// For unique checks
	signerAddressMap := make(map[string]models.Address)
	// For deterministic output of signer values
	var signerAddressArray []models.Address
	signerInfos := make([]*models.SignerInfo, 0)

	// If there is a signer info, get the addresses from the keys add it to the list of signers
	for _, signerInfo := range authInfo.SignerInfos {
		if signerInfo.PublicKey != nil {
			var info models.SignerInfo

			pubKey, err := cl.Codec.InterfaceRegistry.Resolve(signerInfo.PublicKey.TypeUrl)
			if err != nil {
				config.Log.Error("ProcessSigners cl.Codec.InterfaceRegistry.Resolve ", err)
				return nil, nil, err
			}

			err = cl.Codec.InterfaceRegistry.UnpackAny(signerInfo.PublicKey, &pubKey)
			if err != nil {
				config.Log.Error("ProcessSigners cl.Codec.InterfaceRegistry.UnpackAny ", err)
				return nil, nil, err
			}

			multisigKey, ok := pubKey.(*multisig.LegacyAminoPubKey)

			if ok {
				for _, key := range multisigKey.GetPubKeys() {
					address := types.AccAddress(key.Address().Bytes()).String()
					if _, ok := signerAddressMap[address]; !ok {
						signerAddressArray = append(signerAddressArray, models.Address{Address: address})
					}
					signerAddr := models.Address{Address: address}
					signerAddressMap[address] = signerAddr
					info.Address = &signerAddr
				}
			} else {
				castPubKey, ok := pubKey.(cryptoTypes.PubKey)
				if !ok {
					return nil, nil, err
				}

				address := types.AccAddress(castPubKey.Address().Bytes()).String()
				if _, ok := signerAddressMap[address]; !ok {
					signerAddressArray = append(signerAddressArray, models.Address{Address: address})
				}
				signerAddr := models.Address{Address: address}
				signerAddressMap[address] = signerAddr
				info.Address = &signerAddr
			}

			info.Sequence = signerInfo.Sequence
			info.ModeInfo = signerInfo.ModeInfo.String()
			signerInfos = append(signerInfos, &info)
		}
	}

	for _, signer := range messageSigners {
		addressStr := signer.String()
		if _, ok := signerAddressMap[addressStr]; !ok {
			signerAddressArray = append(signerAddressArray, models.Address{Address: addressStr})
		}
		signerAddressMap[addressStr] = models.Address{Address: addressStr}
	}

	// If there is a fee payer, add it to the list of signers
	if authInfo.Fee.GetPayer() != "" {
		if _, ok := signerAddressMap[authInfo.Fee.GetPayer()]; !ok {
			signerAddressArray = append(signerAddressArray, models.Address{Address: authInfo.Fee.GetPayer()})
		}
		signerAddressMap[authInfo.Fee.GetPayer()] = models.Address{Address: authInfo.Fee.GetPayer()}
	}

	return signerAddressArray, signerInfos, nil
}

// ProcessFees Processes fees into model form, applying denoms and addresses to them
func ProcessFees(db *gorm.DB, authInfo cosmosTx.AuthInfo, signers []models.Address) ([]models.Fee, error) {
	// TODO not the best way
	if authInfo.Fee == nil {
		fees := make([]models.Fee, 0)
		return fees, nil
	}

	feeCoins := authInfo.Fee.Amount
	payer := authInfo.Fee.GetPayer()
	fees := make([]models.Fee, 0)

	for _, coin := range feeCoins {
		zeroFee := big.NewInt(0)

		if zeroFee.Cmp(coin.Amount.BigInt()) != 0 {
			amount := util.ToNumeric(coin.Amount.BigInt())
			denom := models.Denom{Base: coin.Denom}

			payerAddr := models.Address{}
			if payer != "" {
				payerAddr.Address = payer
			} else if len(signers) > 0 {
				payerAddr = signers[0]
			}

			fees = append(fees, models.Fee{Amount: amount, Denomination: denom, PayerAddress: payerAddr})
		}
	}

	return fees, nil
}

func ProcessMessage(messageIndex int, message types.Msg, txMessageEventLogs []txtypes.LogMessage, uniqueEventTypes map[string]models.MessageEventType, uniqueEventAttributeKeys map[string]models.MessageEventAttributeKey) (string, dbTypes.MessageDBWrapper) {
	var currMessage models.Message
	var currMessageType models.MessageType
	currMessage.MessageIndex = messageIndex

	// Get the message log that corresponds to the current message
	var currMessageDBWrapper dbTypes.MessageDBWrapper
	messageLog := txtypes.GetMessageLogForIndex(txMessageEventLogs, messageIndex)

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
	return currMessageType.MessageType, currMessageDBWrapper
}
