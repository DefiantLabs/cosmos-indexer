package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/denoms"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
)

var apiEndpoints = map[string]string{
	"blocks_endpoint":              "/cosmos/base/tendermint/v1beta1/blocks/%d",
	"latest_block_endpoint":        "/blocks/latest",
	"txs_by_block_height_endpoint": "/cosmos/tx/v1beta1/txs?events=tx.height=%d&pagination.limit=100&order_by=ORDER_BY_UNSPECIFIED",
	"denom-traces":                 "/ibc/apps/transfer/v1/denom_traces",
}

func GetEndpoint(key string) string {
	return apiEndpoints[key]
}

// GetBlockByHeight makes a request to the Cosmos REST API to get a block by height
func GetBlockByHeight(host string, height uint64) (result tx.GetBlockByHeightResponse, err error) {
	requestEndpoint := fmt.Sprintf(apiEndpoints["blocks_endpoint"], height)

	resp, err := http.Get(fmt.Sprintf("%s%s", host, requestEndpoint))
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	err = checkResponseErrorCode(requestEndpoint, resp)
	if err != nil {
		return result, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// GetTxsByBlockHeight makes a request to the Cosmos REST API and returns all the transactions for a specific block
func GetTxsByBlockHeight(host string, height uint64) (tx.GetTxByBlockHeightResponse, error) {
	var result tx.GetTxByBlockHeightResponse

	requestEndpoint := fmt.Sprintf(apiEndpoints["txs_by_block_height_endpoint"], height)

	resp, err := http.Get(fmt.Sprintf("%s%s", host, requestEndpoint))
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	err = checkResponseErrorCode(requestEndpoint, resp)

	if err != nil {
		return result, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func GetLatestBlock(host string) (tx.GetLatestBlockResponse, error) {
	var result tx.GetLatestBlockResponse

	requestEndpoint := apiEndpoints["latest_block_endpoint"]

	resp, err := http.Get(fmt.Sprintf("%s%s", host, requestEndpoint))
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	err = checkResponseErrorCode(requestEndpoint, resp)

	if err != nil {
		return result, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func checkResponseErrorCode(requestEndpoint string, resp *http.Response) error {
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error getting response for endpoint %s: Status %s Body %s", requestEndpoint, resp.Status, body)
	}

	return nil
}

func GetIBCDenomTraces(host string) (result denoms.GetDenomTracesResponse, err error) {
	result, err = getDenomTraces(host, "")
	if err != nil {
		return
	}
	allResults := result
	for result.Pagination.NextKey != "" {
		result, err = getDenomTraces(host, result.Pagination.NextKey)
		if err != nil {
			return
		}
		allResults.DenomTraces = append(allResults.DenomTraces, result.DenomTraces...)
	}
	result = allResults
	return
}

func getDenomTraces(host string, paginationKey string) (result denoms.GetDenomTracesResponse, err error) {
	requestEndpoint := apiEndpoints["denom-traces"]
	u := fmt.Sprintf("%s%s", host, requestEndpoint)
	if paginationKey != "" {
		u = fmt.Sprintf("%v?pagination.key=%v", u, url.QueryEscape(paginationKey))
	}

	resp, err := http.Get(u) //nolint:gosec
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	err = checkResponseErrorCode(requestEndpoint, resp)
	if err != nil {
		return result, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
