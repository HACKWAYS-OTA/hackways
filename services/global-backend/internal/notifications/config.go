package notifications

import (
	"net/url"
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

// A scoped preview worker must not advertise delivery for templates it cannot send.
func EmailConfiguredFor(template string) bool {
	if !EmailConfigured() {
		return false
	}
	filter := strings.TrimSpace(os.Getenv("EMAIL_OUTBOX_TEMPLATE_FILTER"))
	if filter == "" {
		return true
	}
	for _, value := range strings.Split(filter, ",") {
		if strings.TrimSpace(value) == template {
			return true
		}
	}
	return false
}

func accountOrigin() string {
	if os.Getenv("APP_ENV") == "development" {
		u, err := url.Parse(strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/"))
		if err == nil && u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1") && u.User == nil && u.Path == "" && u.RawQuery == "" && u.Fragment == "" {
			return u.String()
		}
	}
	return "https://hackways.in"
}

// Production reset links remain canonical. Local previews reset only their own DB.
func PasswordResetURL(secret string) string {
	return accountOrigin() + "/auth/reset-password?key=" + url.QueryEscape(secret)
}

func programApplicationURL(slug string) string {
	if slug == "" || len(slug) > 100 || strings.IndexFunc(slug, func(r rune) bool { return r != '-' && (r < 'a' || r > 'z') && (r < '0' || r > '9') }) >= 0 {
		return accountOrigin() + "/programs"
	}
	return accountOrigin() + "/programs/" + url.PathEscape(slug) + "/application"
}

func safeResetURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.User != nil || u.Fragment != "" || u.Path != "/auth/reset-password" {
		return ""
	}
	if os.Getenv("APP_ENV") == "development" && u.Scheme == "https" && u.Host == "hackways.in" {
		// Existing local drafts can predate APP_PUBLIC_URL configuration.
		local, _ := url.Parse(accountOrigin())
		u.Scheme, u.Host = local.Scheme, local.Host
	}
	if secure := safeHTTPSURL(u.String()); secure != "" {
		return secure
	}
	if os.Getenv("APP_ENV") != "development" {
		return ""
	}
	if u.Scheme+"://"+u.Host != accountOrigin() {
		return ""
	}
	return u.String()
}
