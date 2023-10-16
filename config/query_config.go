package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type QueryConfig struct {
	Database Database
	Log      log
	Base     queryBase
}

type queryBase struct {
	Addresses []string `mapstructure:"addresses"`
	Format    string   `mapstructure:"format"`
	StartDate string   `mapstructure:"start-date"`
	EndDate   string   `mapstructure:"end-date"`
}

func SetupQuerySpecificFlags(validParserKeys []string, conf *QueryConfig, cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&conf.Base.Addresses, "address", nil, "A comma separated list of the address(s) to query. (Both '--address addr1,addr2' and '--address addr1 --address addr2' are valid)")
	cmd.Flags().StringVar(&conf.Base.StartDate, "start-date", "", "If set, tx before this date will be ignored. (Dates must be specified in the format 'YYYY-MM-DD:HH:MM:SS' in UTC)")
	cmd.Flags().StringVar(&conf.Base.EndDate, "end-date", "", "If set, tx on or after this date will be ignored. (Dates must be specified in the format 'YYYY-MM-DD:HH:MM:SS' in UTC)")
	defaultParser := ""
	if len(validParserKeys) != 0 {
		defaultParser = validParserKeys[0]
	}

	cmd.Flags().StringVar(&conf.Base.Format, "format", defaultParser, "The format to output")
}

func (conf *QueryConfig) Validate(validCsvParsers []string) error {
	found := false

	for _, v := range validCsvParsers {
		if v == conf.Base.Format {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("invalid format %s, valid formats are %s", conf.Base.Format, validCsvParsers)
	}

	// Validate addresses
	for _, address := range conf.Base.Addresses {
		if strings.Contains(address, ",") {
			return fmt.Errorf("invalid address %s, addresses cannot contain commas", address)
		} else if strings.Contains(address, " ") {
			return fmt.Errorf("invalid address '%v', addresses cannot contain spaces", address)
		}
	}

	expectedLayout := "2006-01-02:15:04:05"

	if conf.Base.StartDate != "" {
		_, err := time.Parse(expectedLayout, conf.Base.StartDate)
		if err != nil {
			return fmt.Errorf("invalid start date '%v'", conf.Base.StartDate)
		}
	}
	if conf.Base.EndDate != "" {
		_, err := time.Parse(expectedLayout, conf.Base.EndDate)
		if err != nil {
			return fmt.Errorf("invalid end date '%v'", conf.Base.EndDate)
		}
	}

	return nil
}

func CheckSuperfluousQueryKeys(keys []string) []string {
	validKeys := make(map[string]struct{})

	addDatabaseConfigKeys(validKeys)
	addLogConfigKeys(validKeys)

	// add base keys
	for _, key := range getValidConfigKeys(queryBase{}, "base") {
		validKeys[key] = struct{}{}
	}

	// Check keys
	ignoredKeys := make([]string, 0)
	for _, key := range keys {
		if _, ok := validKeys[key]; !ok {
			ignoredKeys = append(ignoredKeys, key)
		}
	}

	return ignoredKeys
}
