package model

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
