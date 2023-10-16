package config

import "github.com/spf13/cobra"

type UpdateDenomsConfig struct {
	Database Database
	Lens     lens
	Log      log
	Base     updateDenomsBase
}

type updateDenomsBase struct {
	retryBase
	UpdateAll bool `mapstructure:"update-all"`
}

func SetupUpdateDenomsSpecificFlags(conf *UpdateDenomsConfig, cmd *cobra.Command) {
	cmd.Flags().BoolVar(&conf.Base.UpdateAll, "base.update-all", false, "If provided, the update script will ignore the config chain-id and update all denoms by reaching out to all assetlists supported.")
	cmd.PersistentFlags().Int64Var(&conf.Base.RequestRetryAttempts, "base.request-retry-attempts", 0, "number of RPC query retries to make")
	cmd.PersistentFlags().Uint64Var(&conf.Base.RequestRetryMaxWait, "base.request-retry-max-wait", 30, "max retry incremental backoff wait time in seconds")
}

func (conf *UpdateDenomsConfig) Validate() error {
	err := validateDatabaseConf(conf.Database)
	if err != nil {
		return err
	}

	lensConf := conf.Lens

	lensConf, err = validateLensConf(lensConf)

	if err != nil {
		return err
	}

	conf.Lens = lensConf

	return nil
}

func CheckSuperfluousUpdateDenomsKeys(keys []string) []string {
	validKeys := make(map[string]struct{})

	addDatabaseConfigKeys(validKeys)
	addLogConfigKeys(validKeys)
	addLensConfigKeys(validKeys)

	// add base keys
	for _, key := range getValidConfigKeys(updateDenomsBase{}, "base") {
		validKeys[key] = struct{}{}
	}

	for _, key := range getValidConfigKeys(retryBase{}, "base") {
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
