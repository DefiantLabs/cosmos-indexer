package assetlists

import (
	"encoding/json"
	"io"
	"net/http"
)

func GetAssetList(url string) (AssetList, error) {
	var assetList AssetList
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return assetList, err
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return assetList, err
	}

	err = json.Unmarshal(resBody, &assetList)

	return assetList, err
}
