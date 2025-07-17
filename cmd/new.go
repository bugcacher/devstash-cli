package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new snippet using your default text editor",
	Long: `Opens your default editor (or the one specified by $EDITOR) to create a new snippet.

The snippet is saved to DevStash when you save and close the editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			// Fallback to a common default if $EDITOR is not set
			editor = "vim"
		}

		// Create a temporary file
		tempFile, err := os.CreateTemp("", "devstash-*.md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temporary file: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(tempFile.Name())

		// Open the editor with the temporary file
		editorCmd := exec.Command(editor, tempFile.Name())
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
			os.Exit(1)
		}

		// Read the content from the temp file
		content, err := os.ReadFile(tempFile.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading snippet from temp file: %v\n", err)
			os.Exit(1)
		}

		if strings.TrimSpace(string(content)) == "" {
			fmt.Fprintln(os.Stderr, "Snippet is empty, aborting.")
			os.Exit(1)
		}

		webhookURL := viper.GetString("webhookUrl")
		if webhookURL == "" {
			fmt.Fprintln(os.Stderr, "Error: Webhook URL is not configured. Use 'devstash config set webhookUrl <url>'")
			os.Exit(1)
		}

		// Get flags
		tags, _ := cmd.Flags().GetString("tags")
		note, _ := cmd.Flags().GetString("note")

		// Create and send payload
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

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating JSON payload: %v\n", err)
			os.Exit(1)
		}

		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(1)
		}

		req.Header.Set("Content-Type", "application/json")
		if authToken := viper.GetString("authToken"); authToken != "" {
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
			fmt.Fprintln(os.Stderr, "Error saving snippet to DevStash.")
			os.Exit(1)
		}

		fmt.Println("Successfully saved to DevStash!")
	},
}

func init() {
	newCmd.Flags().StringP("tags", "t", "", "Comma-separated tags for the snippet")
	newCmd.Flags().StringP("note", "n", "", "A note or description for the snippet")
	rootCmd.AddCommand(newCmd)
}
