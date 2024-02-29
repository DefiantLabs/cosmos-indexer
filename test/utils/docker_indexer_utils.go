package utils

import (
	"fmt"
	"log"
	"strings"

	dockertest "github.com/ory/dockertest/v3"
)

func SetupTestIndexer(optionalDockerNetworkID string) (*TestDockerIndexerConfig, error) {
	pool, err := dockertest.NewPool("")

	if err != nil {
		return nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, err
	}

	var network *dockertest.Network
	var networks []*dockertest.Network
	var networkName string
	var networkCreated bool
	networkCreated, networkName, network, err = findOrCreateDockerNetworkByID(pool, optionalDockerNetworkID)

	if err != nil {
		return nil, err
	}

	networks = append(networks, network)

	// BuildAndRunWithBuildOptions sets built repo name to the resourceName. It must be all lower case
	resourceName := fmt.Sprintf("cosmos-indexer-%s", strings.ToLower(randResourceNameSuffix(10)))

	runOpts := &dockertest.RunOptions{
		Name:     resourceName,
		Networks: networks,
		Cmd:      []string{"/bin/bash", "-c", "sleep infinity"},
	}

	buildOpts := &dockertest.BuildOptions{
		ContextDir: "../..",
		Dockerfile: "Dockerfile.test",
	}

	resource, err := pool.BuildAndRunWithBuildOptions(buildOpts, runOpts)
	if err != nil {
		return nil, err
	}

	clean := func() {
		if network != nil && networkCreated {
			if err := pool.RemoveNetwork(network); err != nil {
				log.Fatalf("Could not remove network: %s", err)
			}
		}

		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	conf := TestDockerIndexerConfig{
		DockerResourceName: resource.Container.Name,
		DockerNetwork:      networkName,
		DockerResource:     resource,
		Clean:              clean,
	}

	return &conf, nil

}

type TestDockerIndexerConfig struct {
	DockerResourceName string
	DockerNetwork      string
	DockerResource     *dockertest.Resource
	Clean              func()
}
