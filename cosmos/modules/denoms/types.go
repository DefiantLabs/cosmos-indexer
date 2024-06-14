package denoms

type Pagination struct {
	NextKey string `json:"next_key"`
	Total   string `json:"total"`
}

/*
type GetDenomTracesResponse struct {
	DenomTraces transfertypes.Traces `json:"denom_traces"`
	Pagination  Pagination           `json:"pagination"`
}
*/
