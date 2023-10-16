package parsers

import (
	"fmt"
	"time"

	"github.com/preichenberger/go-coinbasepro/v2"
)

func GetRate(cbClient *coinbasepro.Client, coin string, transactionTime time.Time) (float64, error) {
	histRate, err := cbClient.GetHistoricRates(fmt.Sprintf("%v-USD", coin), coinbasepro.GetHistoricRatesParams{
		Start:       transactionTime.Add(-1 * time.Minute),
		End:         transactionTime,
		Granularity: 60,
	})
	if err != nil {
		return 0.0, fmt.Errorf("unable to get price for coin '%v' at time '%v'. Err: %v", coin, transactionTime, err)
	}
	if len(histRate) == 0 {
		return 0.0, fmt.Errorf("unable to get price for coin '%v' at time '%v'", coin, transactionTime)
	}

	return histRate[0].Close, nil
}
