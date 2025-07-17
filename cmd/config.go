package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration.",
	Long:  `Manage CLI configuration settings, such as the webhook URL and auth token.`,
}

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value.",
	Long:  `Set a configuration value. Valid keys are 'webhookUrl' and 'authToken'.`,
	Args:  cobra.ExactArgs(2), // Expects exactly two arguments: key and value
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		if key != "webhookUrl" && key != "authToken" {
			fmt.Fprintf(os.Stderr, "Error: Invalid configuration key '%s'. Valid keys are 'webhookUrl', 'authToken'.\n", key)
			os.Exit(1)
		}

		// Set the value in viper
		viper.Set(key, value)

		// Get the config file path
		cfgFile := viper.ConfigFileUsed()
		if cfgFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
				os.Exit(1)
			}
			cfgFile = filepath.Join(home, ".config", "devstash", "config.yaml")
		}

		// Ensure the directory exists
		if err := os.MkdirAll(filepath.Dir(cfgFile), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		// Write the config file
		if err := viper.WriteConfigAs(cfgFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully set '%s' in %s\n", key, cfgFile)
	},
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value.",
	Long:  `Get a configuration value. Valid keys are 'webhookUrl' and 'authToken'.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := viper.GetString(key)
		if value != "" {
			fmt.Println(value)
		} else {
			fmt.Fprintf(os.Stderr, "No value set for key '%s'\n", key)
		}
	},
}

func init() {
	// Add subcommands to the config command
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(getCmd)

	// Add the config command to the root command
	rootCmd.AddCommand(configCmd)
}
