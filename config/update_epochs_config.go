package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

type UpdateEpochsConfig struct {
	Database Database
	Lens     lens
	Base     updateEpochsBase
	Log      log
}

type updateEpochsBase struct {
	throttlingBase
	EpochIdentifier string `mapstructure:"epoch-identifier"`
}

func SetupUpdateEpochsSpecificFlags(conf *UpdateEpochsConfig, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&conf.Base.EpochIdentifier, "base.epoch-identifier", "", "the epoch identifier to update")
}

func (conf *UpdateEpochsConfig) Validate() error {
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

	err = validateThrottlingConf(conf.Base.throttlingBase)

	if err != nil {
		return err
	}

	if conf.Base.EpochIdentifier == "" {
		return fmt.Errorf("epoch identifier must be set")
	}

	return nil
}

func CheckSuperfluousUpdateEpochsKeys(keys []string) []string {
	validKeys := make(map[string]struct{})

	addDatabaseConfigKeys(validKeys)
	addLogConfigKeys(validKeys)
	addLensConfigKeys(validKeys)

	// add base keys
	for _, key := range getValidConfigKeys(updateEpochsBase{}, "base") {
		validKeys[key] = struct{}{}
	}

	for _, key := range getValidConfigKeys(throttlingBase{}, "base") {
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
