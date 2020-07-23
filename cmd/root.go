/*
Copyright © 2020 Thomas Mitchell

*/

//Package cmd defines actions to be run from the command line.
package cmd

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-community/merge-bosh-dns/config"
	"github.com/jhunt/go-ansi"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "merge-bosh-dns",
	Short: "Import remote BOSH DNS data",
	Long: `This program can fetch a remotely served records.json BOSH DNS file and
a locally located BOSH DNS file and merge them together into a new records.json
file for BOSH DNS to ingest.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		bailWith("%s", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "merge-bosh-dns-conf.yml", "config file location")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	setConfigDefaults()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		bailWith("Couldn't read in config: %s", err)
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())

	err = viper.Unmarshal(&cfg)
	if err != nil {
		bailWith("Couldn't unmarshal config: %s", err)
	}
}

func setConfigDefaults() {
	cfg.LocalSource.ScrapeInterval = 30
	cfg.RemoteSource.ScrapeInterval = 30
}

func bailWith(fmt string, args ...interface{}) {
	ansi.Fprintf(os.Stderr, "@R{!! "+fmt+"}\n", args...)
	os.Exit(1)
}
