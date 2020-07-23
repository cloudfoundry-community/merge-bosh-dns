/*
Copyright © 2020 Thomas Mitchell

*/

//Package cmd defines actions to be run from the command line.
package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudfoundry-community/merge-bosh-dns/cmd/internal"
	"github.com/cloudfoundry-community/merge-bosh-dns/merge"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously poll the local and remote sources to merge out new records configs",
	Long:  `Continuously poll the local and remote sources to merge out new records configs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Destination.File == "" {
			return fmt.Errorf("No destination file path was given")
		}

		const (
			UpdateReasonLocal  string = "local"
			UpdateReasonRemote string = "remote"
		)

		//Don't want to trigger merges if the source configs haven't changed
		var currentLocalVersion, currentRemoteVersion uint64

		internal.Log("Starting initial merge")

		//seed config values
		localCfg, err := merge.RecordsConfigFromFilepath(cfg.LocalSource.File)
		if err != nil {
			return fmt.Errorf("Getting config from local source `%s': %s", cfg.LocalSource.File, err)
		}

		currentLocalVersion = localCfg.Version

		remoteClient, err := internal.RecordsClientFromConfig(&cfg)
		if err != nil {
			return fmt.Errorf("Creating client for remote config: %s", err)
		}

		remoteCfg, err := remoteClient.FetchRecordsConfig()
		if err != nil {
			return fmt.Errorf("Fetching remote config: %s", err)
		}

		currentRemoteVersion = remoteCfg.Version

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

		internal.Log("Initial merge was successful")

		stateLock := sync.Mutex{}
		mergeChan := make(chan string)

		internal.Log("Beginning watch of local and remote files")
		//localCfg updater
		go func() {
			scrapeInterval := time.Duration(cfg.LocalSource.ScrapeInterval) * time.Second
			internal.Log("Checking local file every %s\n", scrapeInterval.String())
			for range time.Tick(scrapeInterval) {
				stateLock.Lock()
				var err error
				localCfg, err = merge.RecordsConfigFromFilepath(cfg.LocalSource.File)
				stateLock.Unlock()
				if err != nil {
					internal.LogErr("Getting config from local source `%s': %s", cfg.LocalSource.File, err)
					continue
				}

				if localCfg.Version == currentLocalVersion {
					continue
				}

				internal.Log("local source version change found")
				currentLocalVersion = localCfg.Version

				mergeChan <- UpdateReasonLocal
			}
		}()

		//remoteCfg updater
		go func() {
			scrapeInterval := time.Duration(cfg.RemoteSource.ScrapeInterval) * time.Second
			internal.Log("Checking remote endpoint every %s", scrapeInterval.String())
			for range time.Tick(scrapeInterval) {
				stateLock.Lock()
				var err error
				remoteCfg, err = remoteClient.FetchRecordsConfig()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Fetching remote config: %s", err)
					stateLock.Unlock()
					continue
				}

				if remoteCfg.Version == currentRemoteVersion {
					stateLock.Unlock()
					continue
				}

				currentRemoteVersion = remoteCfg.Version
				internal.Log("remote source version change found")

				err = remoteCfg.Filter(cfg.RemoteSource.Include)
				stateLock.Unlock()
				if err != nil {
					internal.LogErr("Applying include rules to remote config: %s", err)
					continue
				}

				mergeChan <- UpdateReasonRemote
			}
		}()

		for reason := range mergeChan {
			internal.Log("Merging to destination config due to %s change", reason)
			stateLock.Lock()
			err = internal.MergeAndWriteConfigs(
				localCfg,
				remoteCfg,
				cfg.Destination.File,
			)
			stateLock.Unlock()
			if err != nil {
				fmt.Fprintf(os.Stderr, "When merging local and remote configs: %s", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
