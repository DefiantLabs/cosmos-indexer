package model

type SearchResult struct {
	TxHash string `bson:"tx_hash"`
	Type   string `bson:"type"`
}
