package denoms

import transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

type Pagination struct {
	NextKey string `json:"next_key"`
	Total   string `json:"total"`
}

type GetDenomTracesResponse struct {
	DenomTraces transfertypes.Traces `json:"denom_traces"`
	Pagination  Pagination           `json:"pagination"`
}
