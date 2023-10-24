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
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/authz"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/bank"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/distribution"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/gov"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/ibc"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/slashing"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/staking"
	txtypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/vesting"
	"github.com/DefiantLabs/cosmos-indexer/cosmwasm/modules/wasm"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/tendermint/modules/liquidity"
	"github.com/DefiantLabs/cosmos-indexer/util"
	"github.com/DefiantLabs/lens/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	cosmosTx "github.com/cosmos/cosmos-sdk/types/tx"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"gorm.io/gorm"
)

// Unmarshal JSON to a particular type. There can be more than one handler for each type.
var messageTypeHandler = map[string][]func() txtypes.CosmosMessage{
	bank.MsgSend:                                {func() txtypes.CosmosMessage { return &bank.WrapperMsgSend{} }},
	bank.MsgMultiSend:                           {func() txtypes.CosmosMessage { return &bank.WrapperMsgMultiSend{} }},
	distribution.MsgWithdrawDelegatorReward:     {func() txtypes.CosmosMessage { return &distribution.WrapperMsgWithdrawDelegatorReward{} }},
	distribution.MsgWithdrawValidatorCommission: {func() txtypes.CosmosMessage { return &distribution.WrapperMsgWithdrawValidatorCommission{} }},
	distribution.MsgFundCommunityPool:           {func() txtypes.CosmosMessage { return &distribution.WrapperMsgFundCommunityPool{} }},
	gov.MsgDeposit:                              {func() txtypes.CosmosMessage { return &gov.WrapperMsgDeposit{} }},
	gov.MsgSubmitProposal:                       {func() txtypes.CosmosMessage { return &gov.WrapperMsgSubmitProposal{} }},
	staking.MsgDelegate:                         {func() txtypes.CosmosMessage { return &staking.WrapperMsgDelegate{} }},
	staking.MsgUndelegate:                       {func() txtypes.CosmosMessage { return &staking.WrapperMsgUndelegate{} }},
	staking.MsgBeginRedelegate:                  {func() txtypes.CosmosMessage { return &staking.WrapperMsgBeginRedelegate{} }},
	ibc.MsgRecvPacket:                           {func() txtypes.CosmosMessage { return &ibc.WrapperMsgRecvPacket{} }},
	ibc.MsgAcknowledgement:                      {func() txtypes.CosmosMessage { return &ibc.WrapperMsgAcknowledgement{} }},
}

// These messages are ignored for tax purposes.
// Fees will still be tracked, there is just not need to parse the msg body.
var messageTypeIgnorer = map[string]interface{}{
	/////////////////////////////////
	/////// Nontaxable Events ///////
	/////////////////////////////////
	// Authz module actions are not taxable
	authz.MsgExec:   nil,
	authz.MsgGrant:  nil,
	authz.MsgRevoke: nil,
	// Making a config change is not taxable
	distribution.MsgSetWithdrawAddress: nil,
	// Making a stableswap config change is not taxable
	// Voting is not taxable
	gov.MsgVote:         nil,
	gov.MsgVoteWeighted: nil,
	// The IBC msgs below do not create taxable events
	ibc.MsgTransfer:              nil,
	ibc.MsgUpdateClient:          nil,
	ibc.MsgTimeout:               nil,
	ibc.MsgTimeoutOnClose:        nil,
	ibc.MsgCreateClient:          nil,
	ibc.MsgConnectionOpenTry:     nil,
	ibc.MsgConnectionOpenConfirm: nil,
	ibc.MsgChannelOpenTry:        nil,
	ibc.MsgChannelOpenConfirm:    nil,
	ibc.MsgConnectionOpenInit:    nil,
	ibc.MsgConnectionOpenAck:     nil,
	ibc.MsgChannelOpenInit:       nil,
	ibc.MsgChannelOpenAck:        nil,
	ibc.MsgChannelCloseConfirm:   nil,
	ibc.MsgChannelCloseInit:      nil,
	// Unjailing and updating params is not taxable
	slashing.MsgUnjail:       nil,
	slashing.MsgUpdateParams: nil,
	// Creating and editing validator is not taxable
	staking.MsgCreateValidator: nil,
	staking.MsgEditValidator:   nil,

	// Create account is not taxable
	vesting.MsgCreateVestingAccount: nil,

	// Tendermint Liquidity messages are actually executed in batches during periodic EndBlocker events
	// We ignore the Message types since the actual taxable events happen later, and the messages can fail/be refunded
	liquidity.MsgCreatePool:          nil,
	liquidity.MsgDepositWithinBatch:  nil,
	liquidity.MsgWithdrawWithinBatch: nil,
	liquidity.MsgSwapWithinBatch:     nil,

	////////////////////////////////////////////////////
	/////// Possible Taxable Events, future work ///////
	////////////////////////////////////////////////////
	// CosmWasm
	wasm.MsgExecuteContract:                 nil,
	wasm.MsgInstantiateContract:             nil,
	wasm.MsgInstantiateContract2:            nil,
	wasm.MsgStoreCode:                       nil,
	wasm.MsgMigrateContract:                 nil,
	wasm.MsgUpdateAdmin:                     nil,
	wasm.MsgClearAdmin:                      nil,
	wasm.MsgUpdateInstantiationAdmin:        nil,
	wasm.MsgUpdateParams:                    nil,
	wasm.MsgSudoContract:                    nil,
	wasm.MsgPinCodes:                        nil,
	wasm.MsgUnpinCodes:                      nil,
	wasm.MsgStoreAndInstantiateContract:     nil,
	wasm.MsgRemoveCodeUploadParamsAddresses: nil,
	wasm.MsgAddCodeUploadParamsAddresses:    nil,
	wasm.MsgStoreAndMigrateContract:         nil,
}

// Merge the chain specific message type handlers into the core message type handler map.
// Chain specific handlers will be registered BEFORE any generic handlers.
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

// ParseCosmosMessageJSON - Parse a SINGLE Cosmos Message into the appropriate type.
func ParseCosmosMessage(message types.Msg, log txtypes.LogMessage) (txtypes.CosmosMessage, string, error) {
	var ok bool
	var err error
	var msgHandler txtypes.CosmosMessage
	var handlerList []func() txtypes.CosmosMessage

	// Figure out what type of Message this is based on the '@type' field that is included
	// in every Cosmos Message (can be seen in raw JSON for any cosmos transaction).
	cosmosMessage := txtypes.Message{}
	cosmosMessage.Type = types.MsgTypeURL(message)

	// So far we only parsed the '@type' field. Now we get a struct for that specific type.
	if handlerList, ok = messageTypeHandler[cosmosMessage.Type]; !ok {
		return nil, cosmosMessage.Type, txtypes.ErrUnknownMessage
	}

	for _, handlerFunc := range handlerList {
		// Unmarshal the rest of the JSON now that we know the specific type.
		// Note that depending on the type, it may or may not care about logs.
		msgHandler = handlerFunc()
		err = msgHandler.HandleMsg(cosmosMessage.Type, message, &log)

		// We're finished when a working handler is found
		if err == nil {
			break
		}
	}

	return msgHandler, cosmosMessage.Type, err
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

		processedTx.SignerAddress = dbTypes.Address{Address: txFull.FeePayer().String()}
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

		processedTx.SignerAddress = dbTypes.Address{Address: currTx.FeePayer().String()}
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
	if code == 0 {
		for messageIndex, message := range tx.Tx.Body.Messages {
			var currMessage dbTypes.Message
			var currMessageType dbTypes.MessageType
			currMessage.MessageIndex = messageIndex

			// Get the message log that corresponds to the current message
			var currMessageDBWrapper dbTypes.MessageDBWrapper
			messageLog := txtypes.GetMessageLogForIndex(tx.TxResponse.Log, messageIndex)
			cosmosMessage, msgType, err := ParseCosmosMessage(message, *messageLog)
			if err != nil {
				currMessageType.MessageType = msgType
				currMessage.MessageType = currMessageType
				currMessageDBWrapper.Message = currMessage
				if err != txtypes.ErrUnknownMessage {
					// What should we do here? This is an actual error during parsing
					config.Log.Error(fmt.Sprintf("[Block: %v] ParseCosmosMessage failed for msg of type '%v'.", tx.TxResponse.Height, msgType), err)
					config.Log.Error(fmt.Sprint(messageLog))
					config.Log.Error(tx.TxResponse.TxHash)
					config.Log.Error("Issue parsing a cosmos msg that we DO have a parser for! PLEASE INVESTIGATE")
					return txDBWapper, txTime, fmt.Errorf("error parsing message we have a parser for: '%v'", msgType)
				}
				// if this msg isn't include in our list of those we are explicitly ignoring, do something about it.
				// we have decided to throw the error back up the call stack, which will prevent any indexing from happening on this block and add this to the failed block table
				if _, ok := messageTypeIgnorer[msgType]; !ok {
					config.Log.Error(fmt.Sprintf("[Block: %v] ParseCosmosMessage failed for msg of type '%v'. Missing parser and ignore list entry.", tx.TxResponse.Height, msgType))
					return txDBWapper, txTime, fmt.Errorf("missing parser and ignore list entry for msg type '%v'", msgType)
				}
			} else {
				config.Log.Debug(fmt.Sprintf("[Block: %v] Cosmos message of known type: %s", tx.TxResponse.Height, cosmosMessage))
				currMessageType.MessageType = cosmosMessage.GetType()
				currMessage.MessageType = currMessageType
				currMessageDBWrapper.Message = currMessage

				relevantData := cosmosMessage.ParseRelevantData()

				if len(relevantData) > 0 {
					taxableTxs := make([]dbTypes.TaxableTxDBWrapper, len(relevantData))
					for i, v := range relevantData {
						if v.AmountSent != nil {
							taxableTxs[i].TaxableTx.AmountSent = util.ToNumeric(v.AmountSent)
						}
						if v.AmountReceived != nil {
							taxableTxs[i].TaxableTx.AmountReceived = util.ToNumeric(v.AmountReceived)
						}

						if v.DenominationSent != "" {
							denomSent, err := getDenom(v.DenominationSent)
							if err != nil {
								// attempt to add missing denoms to the database
								config.Log.Warnf("Denom lookup failed. Will be inserted as UNKNOWN. Denom Sent: %v. Err: %v", denomSent.Base, err)
								denomSent, err = dbTypes.AddUnknownDenom(db, denomSent.Base)
								if err != nil {
									config.Log.Error(fmt.Sprintf("There was an error adding a missing denom. Denom sent: %v", denomSent.Base), err)
									return txDBWapper, txTime, err
								}
							}

							taxableTxs[i].TaxableTx.DenominationSent = denomSent
						}

						if v.DenominationReceived != "" {
							denomReceived, err := getDenom(v.DenominationReceived)
							if err != nil {
								// attempt to add missing denoms to the database
								config.Log.Warnf("Denom lookup failed. Will be inserted as UNKNOWN. Denom Received: %v. Err: %v", denomReceived.Base, err)
								denomReceived, err = dbTypes.AddUnknownDenom(db, denomReceived.Base)
								if err != nil {
									config.Log.Error(fmt.Sprintf("There was an error adding a missing denom. Denom received: %v", denomReceived.Base), err)
									return txDBWapper, txTime, err
								}
							}

							taxableTxs[i].TaxableTx.DenominationReceived = denomReceived
						}

						taxableTxs[i].SenderAddress = dbTypes.Address{Address: strings.ToLower(v.SenderAddress)}
						taxableTxs[i].ReceiverAddress = dbTypes.Address{Address: strings.ToLower(v.ReceiverAddress)}
					}
					currMessageDBWrapper.TaxableTxs = taxableTxs
				} else {
					currMessageDBWrapper.TaxableTxs = []dbTypes.TaxableTxDBWrapper{}
				}
			}

			messages = append(messages, currMessageDBWrapper)
		}
	}

	fees, err := ProcessFees(db, tx.Tx.AuthInfo, tx.Tx.Signers)
	if err != nil {
		return txDBWapper, txTime, err
	}

	txDBWapper.Tx = dbTypes.Tx{Hash: tx.TxResponse.TxHash, Fees: fees, Code: code}
	txDBWapper.Messages = messages

	return txDBWapper, txTime, nil
}

// ProcessFees returns a comma delimited list of fee amount/denoms
func ProcessFees(db *gorm.DB, authInfo cosmosTx.AuthInfo, signers []types.AccAddress) ([]dbTypes.Fee, error) {
	feeCoins := authInfo.Fee.Amount
	payer := authInfo.Fee.GetPayer()
	fees := []dbTypes.Fee{}

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
			payerAddr := dbTypes.Address{}
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

			fees = append(fees, dbTypes.Fee{Amount: amount, Denomination: denom, PayerAddress: payerAddr})
		}
	}

	return fees, nil
}

// getDenom handles denom processing for both IBC denoms and native denoms.
// If the denom begins with ibc/ we know this is an IBC denom trace, and it's not guaranteed there is an entry in
// the Denom table.
func getDenom(denom string) (dbTypes.Denom, error) {
	var (
		denomSent dbTypes.Denom
		err       error
	)

	// if this is an ibc denom trace, get the ibc denom then use the base denom to get the Denom from the db
	if strings.HasPrefix(denom, "ibc/") {
		ibcDenom, err := dbTypes.GetIBCDenom(denom)
		if err != nil {
			config.Log.Warnf("IBC Denom lookup failed for  %s, err: %v", denom, err)
		} else {
			denomSent, err = dbTypes.GetDenomForBase(ibcDenom.BaseDenom)
			if err != nil {
				config.Log.Warnf("Denom lookup failed for IBC base denom %s, err: %v", ibcDenom.BaseDenom, err)
				return dbTypes.Denom{Base: ibcDenom.BaseDenom}, err
			}
		}
	}

	// if this is not an ibc denom trace or there was an issue querying the ibc denom trace in the other table,
	// attempt to look up this denom in the regular Denom table
	if denomSent.Base == "" {
		denomSent, err = dbTypes.GetDenomForBase(denom)
		if err != nil {
			return dbTypes.Denom{Base: denom}, err
		}
	}

	return denomSent, nil
}
