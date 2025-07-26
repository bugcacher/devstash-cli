package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tags     string
	note     string
	filePath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devstash",
	Short: "A CLI to save code snippets and notes to DevStash",
	Long: `DevStash CLI allows you to quickly save code snippets, commands, or any text
from your terminal to your self-hosted DevStash instance via a webhook.

Examples:
  # Save from a file
  devstash --file /path/to/code.js --tags "refactor,api"

  # Save from stdin
  cat file.txt | devstash --note "A useful note"

  # Configure your webhook URL
  devstash config set webhookUrl <your-url>`,
	Run: func(cmd *cobra.Command, args []string) {
		var content []byte
		var err error

		// Prioritize file input if the flag is provided
		if filePath != "" {
			content, err = os.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
				os.Exit(1)
			}
		} else {
			// Otherwise, check for piped input from stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				fmt.Fprintln(os.Stderr, "Usage: Provide input via --file flag or pipe. e.g., 'devstash -f file.txt' or 'cat file.txt | devstash'")
				cmd.Help()
				return
			}
			content, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
		}

		if strings.TrimSpace(string(content)) == "" {
			fmt.Fprintln(os.Stderr, "Error: Input content is empty.")
			os.Exit(1)
		}

		if err := sendToWebhook(string(content), tags, note); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Local flags for the root command
	rootCmd.Flags().String("webhook-url", "", "Webhook URL to send data to")
	rootCmd.Flags().String("auth-token", "", "Authentication token for the webhook")
	rootCmd.Flags().StringVarP(&tags, "tags", "t", "", "Comma-separated tags to add to the snippet")
	rootCmd.Flags().StringVarP(&note, "note", "n", "", "A note or description for the snippet")
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to a file to save as a snippet")

	viper.BindPFlag("webhookUrl", rootCmd.Flags().Lookup("webhook-url"))
	viper.BindPFlag("authToken", rootCmd.Flags().Lookup("auth-token"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in the default location: $HOME/.config/devstash/config.yaml
	viper.AddConfigPath(filepath.Join(home, ".config", "devstash"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Read in environment variables that match
	viper.SetEnvPrefix("DEVSTASH")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
