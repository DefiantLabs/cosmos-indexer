package utils

import (
	"fmt"

	dockertest "github.com/ory/dockertest/v3"
)

func findOrCreateDockerNetworkByID(pool *dockertest.Pool, optionalDockerNetworkID string) (bool, string, *dockertest.Network, error) {
	var network *dockertest.Network
	var networkName string
	created := true
	if optionalDockerNetworkID == "" {

		// create a docker network and attach to the resource
		networkName = fmt.Sprintf("test-network-%s", randResourceNameSuffix(10))
		n, err := pool.CreateNetwork(networkName)

		if err != nil {
			return created, "", nil, err
		}
		network = n

	} else {
		created = false
		externalNetworks, err := pool.Client.ListNetworks()
		if err != nil {
			return created, "", nil, err
		}

		var foundNetwork int = -1
		for i, externalNetwork := range externalNetworks {
			if externalNetwork.ID == optionalDockerNetworkID {
				foundNetwork = i
				break
			}
		}

		if foundNetwork < 0 {
			return created, "", nil, fmt.Errorf("could not find network by ID: %s", optionalDockerNetworkID)
		}

		networkCast, err := pool.NetworksByName(externalNetworks[foundNetwork].Name)

		if err != nil {
			return created, "", nil, err
		}

		if len(networkCast) == 0 {
			return created, "", nil, fmt.Errorf("could not find network with ID %s by name: %s", optionalDockerNetworkID, externalNetworks[foundNetwork].Name)
		}

		networkName = networkCast[0].Network.Name
		network = &networkCast[0]
	}

	return created, networkName, network, nil
}
