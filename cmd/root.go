package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.booking.com/infra/nestor/server"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "nestor",
	Short: "Nestor is a bridge between ServerDB and Kea",
	Long: `Nestor is a bridge between ServerDB and Kea. Nestor will query the ServerDB and store
the data using lru cache to ensure we don't slow down the dhcp handling.
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		server.Serve()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/bmc-toolbox/nestor.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".nestor" (without extension).
		viper.SetConfigName("nestor")
		viper.AddConfigPath(fmt.Sprintf("%s/.bmc-toolbox", home))
		viper.AddConfigPath("/etc/bmc-toolbox")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetDefault("notify_url", "https://serverdb.booking.com")
	viper.SetDefault("socket_path", "/run/nestor/nestor.sock")
	viper.SetDefault("cache_file", "/var/lib/nestor/nestor.state")
	viper.SetDefault("debug", false)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
