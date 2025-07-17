package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devstash",
	Short: "A CLI to save code snippets and notes to your DevStash vault.",
	Long: `DevStash CLI allows you to quickly save text from your terminal to your personal knowledge vault.

Pipe content into it to save it:
  cat my_script.js | devstash --tags "javascript,api"
  history | grep docker | devstash --note "Useful docker commands"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read from standard input
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Println("Usage: Pipe content into devstash. e.g., 'cat file.txt | devstash'")
			return
		}

		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}

		if strings.TrimSpace(string(content)) == "" {
			fmt.Fprintln(os.Stderr, "Error: Input is empty.")
			os.Exit(1)
		}

		// Get config values from Viper (which has read from flags, env, and file)
		webhookURL := viper.GetString("webhookUrl")
		authToken := viper.GetString("authToken")

		if webhookURL == "" {
			fmt.Fprintln(os.Stderr, "Error: Webhook URL is not set. Use 'devstash config set webhookUrl ...' or the --webhook-url flag.")
			os.Exit(1)
		}

		// Get flags for this command
		tags, _ := cmd.Flags().GetString("tags")
		note, _ := cmd.Flags().GetString("note")

		// Create payload
		type Payload struct {
			ID        string   `json:"id"`
			Content   string   `json:"content"`
			UserTags  []string `json:"userTags"`
			Note      string   `json:"note,omitempty"`
			CreatedAt string   `json:"createdAt"`
		}

		payload := Payload{
			ID:        uuid.New().String(),
			Content:   string(content),
			UserTags:  parseTags(tags),
			Note:      note,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating JSON payload: %v\n", err)
			os.Exit(1)
		}

		// Send request
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(1)
		}

		req.Header.Set("Content-Type", "application/json")
		if authToken != "" {
			req.Header.Set("Authorization", authToken)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending request to webhook: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Error: Received status code %d from webhook: %s\n", resp.StatusCode, string(body))
			os.Exit(1)
		}

		fmt.Println("Successfully saved to DevStash!")
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Local flags for the root command
	rootCmd.Flags().String("webhook-url", "", "Webhook URL to send data to")
	rootCmd.Flags().String("auth-token", "", "Authentication token for the webhook")
	rootCmd.Flags().StringP("tags", "t", "", "Comma-separated tags to add to the snippet")
	rootCmd.Flags().StringP("note", "n", "", "A note or description for the snippet")

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

func parseTags(tagsStr string) []string {
	if strings.TrimSpace(tagsStr) == "" {
		return []string{}
	}
	tags := strings.Split(tagsStr, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}
