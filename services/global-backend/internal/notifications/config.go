package notifications

import (
	"os"
	"strings"
)

// EmailConfiguration supports explicit variables and the previously documented
// ACS connection-string environment contract without logging either form.
func EmailConfiguration() (endpoint, key, sender string) {
	endpoint, key, sender = os.Getenv("ACS_ENDPOINT"), os.Getenv("ACS_ACCESS_KEY"), os.Getenv("ACS_SENDER_ADDRESS")
	for _, field := range strings.Split(os.Getenv("AZURE_ACS_CONNECTION_STRING"), ";") {
		name, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "endpoint":
			if endpoint == "" {
				endpoint = strings.TrimSpace(value)
			}
		case "accesskey":
			if key == "" {
				key = strings.TrimSpace(value)
			}
		}
	}
	if sender == "" {
		sender = os.Getenv("AZURE_ACS_SENDER_ADDRESS")
	}
	return
}
func EmailConfigured() bool {
	endpoint, key, sender := EmailConfiguration()
	return endpoint != "" && key != "" && sender != ""
}
