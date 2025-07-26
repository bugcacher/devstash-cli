# Devstash CLI

This project is currently work in progress.

## Installation

1. Download the binary for your OS from the [releases page](../../releases)
2. Add the binary to your PATH
3. Create an alias for easy access:
   ```bash
   alias devstash='/path/to/devstash-cli'
   ```

Now you can run `devstash` from anywhere in your terminal.

## Configuration

1. Set your n8n workflow webhook URL:
   ```bash
   devstash config set webhookUrl https://your-n8n-instance.com/webhook/stash
   ```

1. Set your n8n workflow webhook Auth Token (Optional):
   ```bash
   devstash config set authToken YOUR_TOKEN
   ```


## Available Commands

```bash
> devstash --help
DevStash CLI allows you to quickly save code snippets, commands, or any text
from your terminal to your self-hosted DevStash instance via a webhook.

Examples:
  # Save from a file
  devstash --file /path/to/code.js --tags "refactor,api"

  # Save from stdin
  cat file.txt | devstash --note "A useful note"

  # Configure your webhook URL
  devstash config set webhookUrl <your-url>

Usage:
  devstash [flags]
  devstash [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Manage CLI configuration.
  help        Help about any command
  new         Create a new snippet using your default text editor

Flags:
      --auth-token string    Authentication token for the webhook
  -f, --file string          Path to a file to save as a snippet
  -h, --help                 help for devstash
  -n, --note string          A note or description for the snippet
  -t, --tags string          Comma-separated tags to add to the snippet
      --webhook-url string   Webhook URL to send data to

Use "devstash [command] --help" for more information about a command.
```