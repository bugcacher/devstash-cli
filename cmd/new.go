package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
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
			fmt.Fprintln(os.Stderr, errSnippetIsEmpty)
			os.Exit(1)
		}

		// Get flags
		tags, _ := cmd.Flags().GetString("tags")
		note, _ := cmd.Flags().GetString("note")

		if err := sendToWebhook(string(content), tags, note); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	newCmd.Flags().StringP("tags", "t", "", "Comma-separated tags for the snippet")
	newCmd.Flags().StringP("note", "n", "", "A note or description for the snippet")
	rootCmd.AddCommand(newCmd)
}
