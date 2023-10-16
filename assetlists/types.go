package assetlists

type AssetList struct {
	ChainName string  `json:"chain_name"`
	Assets    []Asset `json:"assets"`
}

type Asset struct {
	Description string `json:"description"`
	Base        string `json:"base"`
	Symbol      string `json:"symbol"`
	KoinlyID    string `json:"koinly_id"` // currently stored as a string in our assetlist
}
