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
	Total     int64           `json:"total"`
	Total24H  int64           `json:"total_24h"`
	Total48H  int64           `json:"total_48h"`
	Total30D  int64           `json:"total_30d"`
	Volume24H decimal.Decimal `json:"volume_24h"`
	Volume30D decimal.Decimal `json:"volume_30d"`
}

type TxSenderReceiver struct {
	MessageType string `json:"message_type,omitempty"`
	Sender      string `json:"sender,omitempty"`
	Receiver    string `json:"receiver,omitempty"`
	Amount      string `json:"amount,omitempty"`
}
