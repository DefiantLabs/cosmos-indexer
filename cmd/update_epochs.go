package cmd

import (
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/osmosis"
	epochsTypes "github.com/DefiantLabs/cosmos-indexer/osmosis/modules/epochs"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
	"github.com/DefiantLabs/lens/client"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	updateEpochsConfig       config.UpdateEpochsConfig
	updateEpochsDbConnection *gorm.DB
)

func init() {
	config.SetupLogFlags(&updateEpochsConfig.Log, updateEpochsCmd)
	config.SetupDatabaseFlags(&updateEpochsConfig.Database, updateEpochsCmd)
	config.SetupLensFlags(&updateEpochsConfig.Lens, updateEpochsCmd)
	config.SetupThrottlingFlag(&updateEpochsConfig.Base.Throttling, updateEpochsCmd)
	config.SetupUpdateEpochsSpecificFlags(&updateEpochsConfig, updateEpochsCmd)
	rootCmd.AddCommand(updateEpochsCmd)
}

var updateEpochsCmd = &cobra.Command{
	Use:   "update-epochs",
	Short: "Gather Epoch information from the blockchain and index it. Currently only supports the Osmosis Epochs module.",
	Long: `Indexes Epoch information from the blockchain. This command currently only support the Osmosis Epochs module.
	Future versions will support the same concept of Epochs for other Cosmos chains if they exist.`,
	PreRunE: setupUpdateEpochs,
	Run:     updateEpochs,
}

func setupUpdateEpochs(cmd *cobra.Command, args []string) error {
	bindFlags(cmd, viperConf)

	err := updateEpochsConfig.Validate()
	if err != nil {
		return err
	}

	ignoredKeys := config.CheckSuperfluousUpdateDenomsKeys(viperConf.AllKeys())

	if len(ignoredKeys) > 0 {
		config.Log.Warnf("Warning, the following invalid keys will be ignored: %v", ignoredKeys)
	}

	setupLogger(updateEpochsConfig.Log.Level, updateEpochsConfig.Log.Path, updateEpochsConfig.Log.Pretty)

	db, err := connectToDBAndMigrate(updateEpochsConfig.Database)
	if err != nil {
		config.Log.Fatal("Could not establish connection to the database", err)
	}

	updateEpochsDbConnection = db

	return nil
}

func updateEpochs(cmd *cobra.Command, args []string) {
	cfg := updateEpochsConfig
	db := updateEpochsDbConnection

	cl := config.GetLensClient(cfg.Lens)

	epochIdentifier := cfg.Base.EpochIdentifier

	if cl.Config.ChainID == osmosis.ChainID {
		// Setup Chain model item
		var chain dbTypes.Chain
		chain.ChainID = cl.Config.ChainID
		chain.Name = cfg.Lens.ChainName
		res := db.FirstOrCreate(&chain)

		if res.Error != nil {
			config.Log.Fatalf("Error setting up Chain model. Err: %v", res.Error)
		}

		config.Log.Infof("Running Epoch indexer for %s and identifier %s", cl.Config.ChainID, epochIdentifier)

		// Start at latest height to get the latest Epochs and work from there
		latestHeight, err := rpc.GetLatestBlockHeight(cl)
		if err != nil {
			config.Log.Fatalf("Error getting latest block height. Err: %v", err)
		}

		config.Log.Infof("Found latest block height %d", latestHeight)

		if latestHeight > 0 {
			time.Sleep(8 * time.Second)
			currentHeight := latestHeight

			for {
				lastIndexedEpoch, foundLast := indexEpochsAtStartingHeight(db, cl, currentHeight, chain, epochIdentifier, cfg.Base.Throttling)

				if lastIndexedEpoch.EpochNumber <= 1 || foundLast {
					config.Log.Infof("Indexed earliest possible Epoch through Epoch querying method")
					break
				}

				var nextIndexedEpochs []dbTypes.Epoch

				// Get the next Epoch to index
				dbResp := db.Where("epoch_number < ? AND identifier=? AND blockchain_id=?", lastIndexedEpoch.EpochNumber, lastIndexedEpoch.Identifier, chain.ID).Find(&nextIndexedEpochs)

				if dbResp.Error != nil {
					config.Log.Fatal("Error validating all epochs have been indexed", dbResp.Error)
				}

				if len(nextIndexedEpochs) == 0 {
					currentHeight = int64(lastIndexedEpoch.StartHeight - 1)
				} else {
					currentEpochNumber := lastIndexedEpoch.EpochNumber

					for i, epoch := range nextIndexedEpochs {
						if epoch.EpochNumber != currentEpochNumber-1 {
							currentHeight = int64(epoch.StartHeight - 1)
						} else {
							currentEpochNumber = epoch.EpochNumber
						}

						if i == len(nextIndexedEpochs)-1 {
							currentHeight = int64(epoch.StartHeight - 1)
						}
					}

					if currentEpochNumber-1 > 1 && currentHeight > 0 {
						config.Log.Debugf("Next Epoch to index is %d", currentEpochNumber-1)
					} else {
						config.Log.Infof("All possible Epochs indexed through Epoch querying method")
						break
					}
				}

			}

			lastIndexedEpoch := dbTypes.Epoch{Identifier: epochIdentifier, Chain: chain}

			dbResp := db.Where(&lastIndexedEpoch).Order("epoch_number asc").First(&lastIndexedEpoch)

			if dbResp.Error != nil {
				config.Log.Fatal("Error validating all epochs have been indexed", dbResp.Error)
			}

			config.Log.Infof("Last indexed Epoch is %d at height %d", lastIndexedEpoch.EpochNumber, lastIndexedEpoch.StartHeight)

			if lastIndexedEpoch.EpochNumber > 1 {
				config.Log.Error("Last indexed Epoch is not the first, could not index full history of Epochs")
			}
		}
	} else {
		config.Log.Infof("Chain %s is not supported by this command.", cl.Config.ChainID)
	}
}

func indexEpochsAtStartingHeight(db *gorm.DB, cl *client.ChainClient, startingHeight int64, chain dbTypes.Chain, identifierToIndex string, throttling float64) (*dbTypes.Epoch, bool) {
	currentHeight := startingHeight
	var lastIndexedItem dbTypes.Epoch
	for {
		time.Sleep(time.Second * time.Duration(throttling))
		resp, err := rpc.GetEpochsAtHeight(cl, currentHeight)
		if err != nil {
			config.Log.Fatalf("Error getting epochs at height %d. Err: %v", currentHeight, err)
		}

		// Not sure if this is possible, this means that the Epochs module returned no Epochs for a height
		if len(resp.Epochs) == 0 {
			config.Log.Fatalf("No Epochs found at height %d", currentHeight)
		}

		var newItem dbTypes.Epoch
		found := false

		// Get the index of the shortest duration EpochInfo, this will be used for the querying mechanism
		for _, epoch := range resp.Epochs {
			// Make sure we have the ability to index this EpochInfo
			// This will save us trouble if Osmosis adds more Epochs in the future
			indexable, identifierExists := epochsTypes.OsmosisIndexableEpochs[epoch.Identifier]
			if identifierExists && indexable && epoch.Identifier == identifierToIndex {

				if epoch.CurrentEpochStartHeight <= 0 {
					config.Log.Debugf("Found Epoch %d that contains 0 for CurrentEpochStartHeight, cannot continue", epoch.CurrentEpoch)
					return &lastIndexedItem, true
				}

				config.Log.Infof("Found Epoch %d at height %d", epoch.CurrentEpoch, epoch.CurrentEpochStartHeight)
				currentHeight = epoch.CurrentEpochStartHeight - 1
				newItem = dbTypes.Epoch{Chain: chain, Identifier: epoch.Identifier, StartHeight: uint(epoch.CurrentEpochStartHeight), EpochNumber: uint(epoch.CurrentEpoch)}
				found = true
				break
			}
		}

		if !found {
			config.Log.Fatalf("Epoch with identifier %s not found at %d", updateEpochsConfig.Base.EpochIdentifier, currentHeight)
		}

		dbResp := db.Where(&newItem).FirstOrCreate(&newItem)

		if dbResp.Error != nil {
			config.Log.Fatal("Error creating Epoch item", dbResp.Error)
		}

		// We have reached an item we already created for
		if dbResp.RowsAffected == 0 {
			config.Log.Debugf("Epoch already exists for Epoch %d at height %d", newItem.EpochNumber, newItem.StartHeight)
			return &newItem, false
		}

		if currentHeight <= 0 || newItem.EpochNumber == 1 {
			config.Log.Infof("Reached height %d, stopping", currentHeight)
			return &newItem, true
		}

		lastIndexedItem = newItem
	}
}
