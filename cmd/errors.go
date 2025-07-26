package cmd

const (
	errWebhookNotConfigured = "Error: Webhook URL is not configured. Use 'devstash config set webhookUrl <url>'"
	errEmptyInput           = "Error: Input content is empty."
	errSnippetIsEmpty       = "Snippet is empty, aborting."
	errSavingSnippet        = "Error saving snippet to DevStash."
)
