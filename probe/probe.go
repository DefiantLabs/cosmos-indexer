package probe

import (
	"github.com/DefiantLabs/cosmos-indexer/config"
	probeClient "github.com/DefiantLabs/probe/client"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func GetProbeClient(conf config.Probe, appModuleBasicsExtensions []module.AppModuleBasic, customMsgTypeRegistry map[string]sdkTypes.Msg) (*probeClient.ChainClient, error) {
	return probeClient.NewChainClient(GetProbeConfig(conf, true, appModuleBasicsExtensions, customMsgTypeRegistry), "", nil, nil)
}

// Will include the protos provided by the Probe package for Osmosis module interfaces
func IncludeOsmosisInterfaces(client *probeClient.ChainClient) {
	probeClient.RegisterOsmosisInterfaces(client.Codec.InterfaceRegistry)
}

// Will include the protos provided by the Probe package for Tendermint Liquidity module interfaces
func IncludeTendermintInterfaces(client *probeClient.ChainClient) {
	probeClient.RegisterTendermintLiquidityInterfaces(client.Codec.Amino, client.Codec.InterfaceRegistry)
}

func GetProbeConfig(conf config.Probe, debug bool, appModuleBasicsExtensions []module.AppModuleBasic, customMsgTypeRegistry map[string]sdkTypes.Msg) *probeClient.ChainClientConfig {
	moduleBasics := []module.AppModuleBasic{}
	moduleBasics = append(moduleBasics, probeClient.DefaultModuleBasics...)
	moduleBasics = append(moduleBasics, appModuleBasicsExtensions...)

	return &probeClient.ChainClientConfig{
		Key:                   "default",
		ChainID:               conf.ChainID,
		RPCAddr:               conf.RPC,
		AccountPrefix:         conf.AccountPrefix,
		KeyringBackend:        "test",
		Debug:                 debug,
		Timeout:               "30s",
		OutputFormat:          "json",
		Modules:               moduleBasics,
		CustomMsgTypeRegistry: customMsgTypeRegistry,
	}
}
