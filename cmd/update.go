/*
Copyright © 2020 Thomas Mitchell

*/

//Package cmd defines actions to be run from the command line.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/cloudfoundry-community/merge-bosh-dns/cmd/internal"
	"github.com/cloudfoundry-community/merge-bosh-dns/merge"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Run a one-off merge and output it to the configured file",
	Long: `Runs a one-off merge using the configuration parameters found in the
configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Destination.File == "" {
			return fmt.Errorf("No destination file path was given")
		}

		localCfg, err := merge.RecordsConfigFromFilepath(cfg.LocalSource.File)
		if err != nil {
			return fmt.Errorf("Getting config from local source `%s': %s", cfg.LocalSource.File, err)
		}

		remoteClient, err := internal.RecordsClientFromConfig(&cfg)
		if err != nil {
			return fmt.Errorf("Creating client for remote config: %s", err)
		}

		remoteCfg, err := remoteClient.FetchRecordsConfig()
		if err != nil {
			return fmt.Errorf("Fetching remote config: %s", err)
		}

		err = remoteCfg.Filter(cfg.RemoteSource.Include)
		if err != nil {
			return fmt.Errorf("Applying include rules to remote config: %s", err)
		}

		err = internal.MergeAndWriteConfigs(
			localCfg,
			remoteCfg,
			cfg.Destination.File,
		)
		if err != nil {
			return fmt.Errorf("When merging local and remote configs: %s", err)
		}

		fmt.Println("success")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
