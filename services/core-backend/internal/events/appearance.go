package events

import "regexp"

var colorValue = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func validAppearance(req *createEventRequest) bool {
	if req.Visibility == "" {
		req.Visibility = "public"
	}
	if req.Visibility != "public" && req.Visibility != "private" {
		return false
	}
	if req.Theme == "" {
		req.Theme = "minimal"
	}
	if req.ThemeColor == "" {
		req.ThemeColor = "#ddd4e9"
	}
	if req.ThemeFont == "" {
		req.ThemeFont = "futurist"
	}
	if req.ThemeStyle == "" {
		req.ThemeStyle = "soft"
	}
	if req.ThemeDisplay == "" {
		req.ThemeDisplay = "light"
	}
	switch req.Theme {
	case "minimal", "aurora", "ambient", "paper", "lilac", "confetti", "pattern", "playful", "seasonal":
	default:
		return false
	}
	return colorValue.MatchString(req.ThemeColor) && (req.ThemeFont == "futurist" || req.ThemeFont == "google") && (req.ThemeStyle == "soft" || req.ThemeStyle == "crisp") && (req.ThemeDisplay == "light" || req.ThemeDisplay == "dark")
}
