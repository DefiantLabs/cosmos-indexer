package probe

import (
	probeClient "github.com/DefiantLabs/probe/client"
	"github.com/nodersteam/cosmos-indexer/config"
)

func GetProbeClient(conf config.Probe) *probeClient.ChainClient {
	// IMPORTANT: the actual keyring-test will be searched for at the path {homepath}/keys/{ChainID}/keyring-test.
	// You can use probe default settings to generate that directory appropriately then move it to the desired path.
	// For example, 'probe keys restore default' will restore the key to the default keyring (e.g. /home/kyle/.probe/...)
	// and you can move all of the necessary keys to whatever homepath you want to use. Or you can use --home flag.
	cl, err := probeClient.NewChainClient(GetProbeConfig(conf, true), "", nil, nil)
	if err != nil {
		config.Log.Fatalf("Error connecting to chain. Err: %v", err)
	}
	return cl
}

// Will include the protos provided by the Probe package for Osmosis module interfaces
func IncludeOsmosisInterfaces(client *probeClient.ChainClient) {
	probeClient.RegisterOsmosisInterfaces(client.Codec.InterfaceRegistry)
}

// Will include the protos provided by the Probe package for Tendermint Liquidity module interfaces
func IncludeTendermintInterfaces(client *probeClient.ChainClient) {
	probeClient.RegisterTendermintLiquidityInterfaces(client.Codec.Amino, client.Codec.InterfaceRegistry)
}

func GetProbeConfig(conf config.Probe, debug bool) *probeClient.ChainClientConfig {
	return &probeClient.ChainClientConfig{
		Key:            "default",
		ChainID:        conf.ChainID,
		RPCAddr:        conf.RPC,
		AccountPrefix:  conf.AccountPrefix,
		KeyringBackend: "test",
		Debug:          debug,
		Timeout:        "30s",
		OutputFormat:   "json",
		Modules:        probeClient.DefaultModuleBasics,
	}
}
