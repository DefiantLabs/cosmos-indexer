package utils

import (
	"fmt"
	"log"
	"math/rand"

	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/ory/dockertest/v3"
	"gorm.io/gorm"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randResourceNameSuffix(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func SetupTestDatabase() (*TestDockerDBConfig, error) {
	// TODO: allow environment overrides to skip creating mock database?
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, err
	}

	databaseName := "test"
	user := "test"
	password := "test"

	connectUserEnv := fmt.Sprintf("POSTGRES_USER=%s", user)
	connectPasswordEnv := fmt.Sprintf("POSTGRES_PASSWORD=%s", password)
	connectDbEnv := fmt.Sprintf("POSTGRES_DB=%s", databaseName)

	// create a docker network and attach to the resource
	networkName := fmt.Sprintf("test-network-%s", randResourceNameSuffix(10))
	network, err := pool.CreateNetwork(networkName)

	if err != nil {
		return nil, err
	}

	resourceName := fmt.Sprintf("postgres-%s", randResourceNameSuffix(10))

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       resourceName,
		Repository: "postgres",
		Tag:        "15-alpine",
		Env:        []string{connectUserEnv, connectPasswordEnv, connectDbEnv},
		Networks:   []*dockertest.Network{network},
	})
	if err != nil {
		return nil, err
	}

	var db *gorm.DB
	host := resource.GetBoundIP("5432/tcp")
	port := resource.GetPort("5432/tcp")

	if err := pool.Retry(func() error {
		var err error
		db, err = dbTypes.PostgresDbConnect(host, port, databaseName, user, password, "debug")
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	clean := func() {
		if err := pool.RemoveNetwork(network); err != nil {
			log.Fatalf("Could not remove network: %s", err)
		}

		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	conf := TestDockerDBConfig{
		DockerResourceName: resourceName,
		DockerNetwork:      networkName,
		GormDB:             db,
		Host:               host,
		Port:               port,
		Database:           databaseName,
		User:               user,
		Password:           password,
		LogLevel:           "silent",
		Clean:              clean,
	}

	return &conf, nil
}

type TestDockerDBConfig struct {
	DockerResourceName string
	DockerNetwork      string
	GormDB             *gorm.DB
	Host               string
	Port               string
	Database           string
	User               string
	Password           string
	LogLevel           string
	Clean              func()
}
