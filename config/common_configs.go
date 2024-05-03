package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/DefiantLabs/cosmos-indexer/util"
	"github.com/spf13/cobra"
)

// These configs are used across multiple commands, and are not specific to a single command
type log struct {
	Level  string
	Path   string
	Pretty bool
}

type Database struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	LogLevel string `mapstructure:"log-level"`
}

type Probe struct {
	RPC           string
	AccountPrefix string `mapstructure:"account-prefix"`
	ChainID       string `mapstructure:"chain-id"`
	ChainName     string `mapstructure:"chain-name"`
}

type Server struct {
	Port int
}

type RedisConf struct {
	RedisAddr string
	RedisPsw  string
}

type throttlingBase struct {
	Throttling float64 `mapstructure:"throttling"`
}

type retryBase struct {
	RequestRetryAttempts int64  `mapstructure:"request-retry-attempts"`
	RequestRetryMaxWait  uint64 `mapstructure:"request-retry-max-wait"`
}

func SetupLogFlags(logConf *log, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&logConf.Level, "log.level", "info", "log level")
	cmd.PersistentFlags().BoolVar(&logConf.Pretty, "log.pretty", false, "pretty logs")
	cmd.PersistentFlags().StringVar(&logConf.Path, "log.path", "", "log path (default is $HOME/.cosmos-indexer/logs.txt")
}

func SetupDatabaseFlags(databaseConf *Database, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&databaseConf.Host, "database.host", "", "database host")
	cmd.PersistentFlags().StringVar(&databaseConf.Port, "database.port", "5432", "database port")
	cmd.PersistentFlags().StringVar(&databaseConf.Database, "database.database", "", "database name")
	cmd.PersistentFlags().StringVar(&databaseConf.User, "database.user", "", "database user")
	cmd.PersistentFlags().StringVar(&databaseConf.Password, "database.password", "", "database password")
	cmd.PersistentFlags().StringVar(&databaseConf.LogLevel, "database.log-level", "", "database loglevel")
}

func SetupProbeFlags(probeConf *Probe, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&probeConf.RPC, "probe.rpc", "", "node rpc endpoint")
	cmd.PersistentFlags().StringVar(&probeConf.AccountPrefix, "probe.account-prefix", "", "probe account prefix")
	cmd.PersistentFlags().StringVar(&probeConf.ChainID, "probe.chain-id", "", "probe chain ID")
	cmd.PersistentFlags().StringVar(&probeConf.ChainName, "probe.chain-name", "", "probe chain name")
}

func SetupServerFlags(serverConf *Server, cmd *cobra.Command) {
	cmd.PersistentFlags().IntVar(&serverConf.Port, "server.port", 9002, "inbound grpc port")
}

func SetupRedisFlags(redisConf *RedisConf, cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&redisConf.RedisAddr, "redis.addr", "-", "redis address")
	cmd.PersistentFlags().StringVar(&redisConf.RedisPsw, "redis.psw", "-", "redis password")
}

func SetupThrottlingFlag(throttlingValue *float64, cmd *cobra.Command) {
	cmd.PersistentFlags().Float64Var(throttlingValue, "base.throttling", 0.5, "throttle delay")
}

func validateDatabaseConf(dbConf Database) error {
	if util.StrNotSet(dbConf.Host) {
		return errors.New("database host must be set")
	}
	if util.StrNotSet(dbConf.Port) {
		return errors.New("database port must be set")
	}
	if util.StrNotSet(dbConf.Database) {
		return errors.New("database name (i.e. database) must be set")
	}
	if util.StrNotSet(dbConf.User) {
		return errors.New("database user must be set")
	}
	if util.StrNotSet(dbConf.Password) {
		return errors.New("database password must be set")
	}

	return nil
}

func validateProbeConf(probeConf Probe) (Probe, error) {
	if util.StrNotSet(probeConf.RPC) {
		return probeConf, errors.New("probe rpc must be set")
	}
	// add port if not set
	if strings.Count(probeConf.RPC, ":") != 2 {
		if strings.HasPrefix(probeConf.RPC, "https:") {
			probeConf.RPC = fmt.Sprintf("%s:443", probeConf.RPC)
		} else if strings.HasPrefix(probeConf.RPC, "http:") {
			probeConf.RPC = fmt.Sprintf("%s:80", probeConf.RPC)
		}
	}

	if util.StrNotSet(probeConf.AccountPrefix) {
		return probeConf, errors.New("probe account-prefix must be set")
	}
	if util.StrNotSet(probeConf.ChainID) {
		return probeConf, errors.New("probe chain-id must be set")
	}
	if util.StrNotSet(probeConf.ChainName) {
		return probeConf, errors.New("probe chain-name must be set")
	}
	return probeConf, nil
}

func validateThrottlingConf(throttlingConf throttlingBase) error {
	if throttlingConf.Throttling < 0 {
		return errors.New("throttling must be a positive number or 0")
	}
	return nil
}

// Reads the Viper mapstructure tag to get the valid keys for a given config struct
func getValidConfigKeys(section any, baseName string) (keys []string) {
	v := reflect.ValueOf(section)
	typeOfS := v.Type()

	if baseName == "" {
		baseName = strings.ToLower(typeOfS.Name())
	}

	for i := 0; i < v.NumField(); i++ {
		field := typeOfS.Field(i)

		// Hack to get around the fact that we have embedded struct inside a struct in some of our definitions
		if !strings.HasPrefix(field.Type.String(), "config.") {
			name := field.Tag.Get("mapstructure")
			if name == "" {
				name = field.Name
			}

			key := fmt.Sprintf("%v.%v", baseName, strings.ReplaceAll(strings.ToLower(name), " ", ""))
			keys = append(keys, key)
		}
	}
	return
}

func addDatabaseConfigKeys(validKeys map[string]struct{}) {
	for _, key := range getValidConfigKeys(Database{}, "") {
		validKeys[key] = struct{}{}
	}
}

func addLogConfigKeys(validKeys map[string]struct{}) {
	for _, key := range getValidConfigKeys(log{}, "") {
		validKeys[key] = struct{}{}
	}
}

func addProbeConfigKeys(validKeys map[string]struct{}) {
	for _, key := range getValidConfigKeys(Probe{}, "") {
		validKeys[key] = struct{}{}
	}
}
