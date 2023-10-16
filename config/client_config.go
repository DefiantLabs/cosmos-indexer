package config

import (
	"errors"
	"flag"
	"io"
	lg "log"

	"github.com/BurntSushi/toml"
	"github.com/DefiantLabs/cosmos-indexer/util"
	"github.com/imdario/mergo"
)

type ClientConfig struct {
	ConfigFileLocation string
	Database           Database
	Client             client
	Log                log
}

type client struct {
	Model string
}

func ParseClientArgs(w io.Writer, args []string) (ClientConfig, *flag.FlagSet, int, error) {
	c := ClientConfig{}
	fs := flag.NewFlagSet("config", flag.ContinueOnError)

	fs.SetOutput(w)
	fs.StringVar(&c.ConfigFileLocation, "config", "", "The file to load for configuration variables")

	// Database
	fs.StringVar(&c.Database.Host, "db.host", "", "The PostgreSQL hostname for the indexer db")
	fs.StringVar(&c.Database.Database, "db.database", "", "The PostgreSQL database for the indexer db")
	fs.StringVar(&c.Database.Port, "db.port", "5432", "The PostgreSQL port for the indexer db")
	fs.StringVar(&c.Database.Password, "db.password", "", "The PostgreSQL user password for the indexer db")
	fs.StringVar(&c.Database.User, "db.user", "", "The PostgreSQL user for the indexer db")

	// Client
	fs.StringVar(&c.Client.Model, "client.model", "", "The client deployment model (commercial or not)")

	// Service
	var svcPort int
	fs.IntVar(&svcPort, "port", 8080, "the port the UI client will be served from")

	err := fs.Parse(args)
	if err != nil {
		return c, fs, svcPort, err
	}

	return c, fs, svcPort, nil
}

func GetClientConfig(configFileLocation string) (ClientConfig, error) {
	var conf ClientConfig
	_, err := toml.DecodeFile(configFileLocation, &conf)
	return conf, err
}

func MergeClientConfigs(def ClientConfig, overide ClientConfig) ClientConfig {
	err := mergo.Merge(&overide, def)
	if err != nil {
		lg.Panicf("Config merge failed. Err: %v", err)
	}

	return overide
}

// ValidateClientConfig will validate the config for fields required by the client
func (conf *ClientConfig) ValidateClientConfig() error {
	// Database Checks
	if util.StrNotSet(conf.Database.Host) {
		return errors.New("database host must be set")
	}
	if util.StrNotSet(conf.Database.Port) {
		return errors.New("database port must be set")
	}
	if util.StrNotSet(conf.Database.Database) {
		return errors.New("database name (i.e. Database) must be set")
	}
	if util.StrNotSet(conf.Database.User) {
		return errors.New("database user must be set")
	}
	if util.StrNotSet(conf.Database.Password) {
		return errors.New("database password must be set")
	}

	return nil
}
