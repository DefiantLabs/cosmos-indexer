package testdb

import (
	"context"
	"fmt"
	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/mongo"
	mOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // for test
	"github.com/ory/dockertest/v3/docker"
)

const (
	Image   = "mongo"
	Version = "5.0.2"

	MongoDBName = "search_indexer"
)

func NewMongoDB() (*mongo.Database, func(), error) {
	// Uses a sensible default on windows (tcp/http) and linux/osx (socket).
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to docker: %v", err)
	}
	options := &dockertest.RunOptions{
		Repository: Image,
		Tag:        Version,
		Env: []string{
			"MONGO_INITDB_DATABASE=" + MongoDBName,
			"MONGO_INITDB_ROOT_USERNAME=admin",
			"MONGO_INITDB_ROOT_PASSWORD=password",
		},
		Labels:       map[string]string{"goose_test": "1"},
		PortBindings: make(map[docker.Port][]docker.PortBinding),
	}
	options.PortBindings["27017/tcp"] = []docker.PortBinding{
		{HostPort: strconv.Itoa(27017), HostIP: "0.0.0.0"}, //nolint:gomnd
	}
	options.PortBindings["27018/tcp"] = []docker.PortBinding{
		{HostPort: strconv.Itoa(27018), HostIP: "0.0.0.0"}, //nolint:gomnd
	}
	options.ExposedPorts = []string{"27017", "27018"}

	container, err := pool.RunWithOptions(
		options,
		func(config *docker.HostConfig) {
			// Set AutoRemove to true so that stopped container goes away by itself.
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create docker container: %v", err)
	}
	cleanup := func() {
		if err = pool.Purge(container); err != nil {
			log.Printf("failed to purge resource: %v", err)
		}
	}

	time.Sleep(5 * time.Second)

	port := container.GetPort("27017/tcp")
	ctx := context.Background()
	mongoURL := fmt.Sprintf("mongodb://admin:password@localhost:%s", port)

	mongoClient, err := mongo.Connect(ctx, mOptions.Client().ApplyURI(mongoURL))
	if err != nil {
		panic(err)
	}
	err = mongoClient.Ping(ctx, &readpref.ReadPref{})
	if err != nil {
		panic(err)
	}

	db := mongoClient.Database(MongoDBName)

	return db, cleanup, err
}
