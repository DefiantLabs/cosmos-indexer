package model

import "github.com/shopspring/decimal"

type Tx struct {
	Messages []string
	Memo     string
	AuthInfo TxAuthInfo
}

type TxAuthInfo struct {
	PublicKeys []string
	Fee        TxFee
	Signatures []string
}

type TxFee struct {
	Amount   Denom
	GasLimit string
	Payer    string
	Granter  string
}

type Denom struct {
	Denom  string
	Amount string
}

type TotalTransactions struct {
	Total     int64
	Total24H  int64
	Total30D  int64
	Volume24H decimal.Decimal
	Volume30D decimal.Decimal
}

type TxSenderReceiver struct {
	MessageType string `json:"message_type,omitempty"`
	Sender      string `json:"sender,omitempty"`
	Receiver    string `json:"receiver,omitempty"`
	Amount      string `json:"amount,omitempty"`
}
