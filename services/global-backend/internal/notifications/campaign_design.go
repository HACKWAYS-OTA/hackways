package notifications

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// CampaignDesign is bounded presentation data, never executable HTML or CSS.
type CampaignDesign struct {
	Preheader  string          `json:"preheader"`
	Heading    string          `json:"heading"`
	Brand      string          `json:"brand"`
	BannerURL  string          `json:"banner_url"`
	BannerAlt  string          `json:"banner_alt"`
	Background string          `json:"background"`
	Surface    string          `json:"surface"`
	Text       string          `json:"text"`
	Accent     string          `json:"accent"`
	ButtonText string          `json:"button_text"`
	Font       string          `json:"font"`
	Align      string          `json:"align"`
	Layout     string          `json:"layout"`
	Blocks     []CampaignBlock `json:"blocks"`
}

type CampaignBlock struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
	URL  string `json:"url"`
	Alt  string `json:"alt"`
}

var campaignColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func campaignHTTPS(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && len(value) <= 2000
}

func (d *CampaignDesign) Normalize() error {
	if d.Blocks == nil {
		d.Blocks = []CampaignBlock{}
	}
	defaults := []struct {
		value    *string
		fallback string
	}{
		{&d.Brand, "Hackways"}, {&d.Background, "#f5f2f8"}, {&d.Surface, "#ffffff"},
		{&d.Text, "#29222f"}, {&d.Accent, "#604176"}, {&d.ButtonText, "#ffffff"},
		{&d.Font, "sans"}, {&d.Align, "left"}, {&d.Layout, "card"},
	}
	for _, item := range defaults {
		*item.value = strings.TrimSpace(*item.value)
		if *item.value == "" {
			*item.value = item.fallback
		}
	}
	for _, color := range []string{d.Background, d.Surface, d.Text, d.Accent, d.ButtonText} {
		if !campaignColor.MatchString(color) {
			return fmt.Errorf("Choose six-digit hex colors")
		}
	}
	if d.Font != "sans" && d.Font != "serif" && d.Font != "mono" {
		return fmt.Errorf("Choose a supported email font")
	}
	if d.Align != "left" && d.Align != "center" {
		return fmt.Errorf("Choose left or center alignment")
	}
	if d.Layout != "card" && d.Layout != "plain" {
		return fmt.Errorf("Choose a supported email layout")
	}
	for _, item := range []struct {
		value string
		max   int
	}{{d.Brand, 80}, {d.Preheader, 200}, {d.Heading, 180}, {d.BannerAlt, 180}} {
		if !utf8.ValidString(item.value) || strings.ContainsRune(item.value, 0) || utf8.RuneCountInString(item.value) > item.max {
			return fmt.Errorf("Keep campaign headings and preview text within their limits")
		}
	}
	if d.BannerURL != "" && !campaignHTTPS(d.BannerURL) {
		return fmt.Errorf("Use a full HTTPS banner image URL")
	}
	if len(d.Blocks) > 12 {
		return fmt.Errorf("Use up to twelve content blocks")
	}
	total := 0
	for _, b := range d.Blocks {
		total += len(b.Text) + len(b.Alt)
		if !utf8.ValidString(b.Text+b.Alt) || strings.ContainsRune(b.Text+b.Alt, 0) || utf8.RuneCountInString(b.Text) > 4000 || utf8.RuneCountInString(b.Alt) > 180 {
			return fmt.Errorf("Keep each block within 4,000 characters")
		}
		switch b.Kind {
		case "heading":
			if strings.TrimSpace(b.Text) == "" || utf8.RuneCountInString(b.Text) > 180 {
				return fmt.Errorf("Add a heading within 180 characters")
			}
		case "text":
			if strings.TrimSpace(b.Text) == "" {
				return fmt.Errorf("Add text to each paragraph block")
			}
		case "image":
			if !campaignHTTPS(b.URL) || strings.TrimSpace(b.Alt) == "" {
				return fmt.Errorf("Images need an HTTPS URL and descriptive alternative text")
			}
		case "button":
			if !campaignHTTPS(b.URL) || strings.TrimSpace(b.Text) == "" || utf8.RuneCountInString(b.Text) > 80 {
				return fmt.Errorf("Buttons need a label and HTTPS link")
			}
		case "divider":
		default:
			return fmt.Errorf("Choose a supported content block")
		}
	}
	if total > 24000 {
		return fmt.Errorf("Keep additional campaign content within 24,000 bytes")
	}
	return nil
}

func renderCampaignDesign(subject, body, label, actionURL, unsubscribe string, test bool, d CampaignDesign, promotional bool) (string, string, error) {
	if err := d.Normalize(); err != nil {
		return "", "", err
	}
	escape := html.EscapeString
	font := map[string]string{"sans": "Arial,Helvetica,sans-serif", "serif": "Georgia,Times New Roman,serif", "mono": "Courier New,Courier,monospace"}[d.Font]
	button := func(text, link string) string {
		if link == "" {
			return ""
		}
		return `<p style="margin:24px 0;"><a href="` + escape(link) + `" style="display:inline-block;padding:13px 22px;background:` + d.Accent + `;color:` + d.ButtonText + `;border-radius:8px;text-decoration:none;font-weight:600;">` + escape(text) + `</a></p>`
	}
	image := func(link, alt string) string {
		return `<img src="` + escape(link) + `" alt="` + escape(alt) + `" width="560" style="display:block;width:100%;max-width:560px;height:auto;border:0;margin:0 auto 24px;border-radius:8px;" />`
	}
	heading := d.Heading
	if heading == "" {
		heading = subject
	}
	var content strings.Builder
	if d.BannerURL != "" {
		content.WriteString(image(d.BannerURL, d.BannerAlt))
	}
	content.WriteString(`<h1 style="font-size:30px;line-height:1.2;margin:0 0 24px;">` + escape(heading) + `</h1>`)
	content.WriteString(`<p style="margin:0 0 24px;white-space:pre-wrap;">` + strings.ReplaceAll(escape(body), "\n", "<br />") + `</p>`)
	plain := body
	for _, b := range d.Blocks {
		switch b.Kind {
		case "heading":
			content.WriteString(`<h2 style="font-size:23px;line-height:1.3;margin:28px 0 12px;">` + escape(b.Text) + `</h2>`)
			plain += "\n\n" + b.Text
		case "text":
			content.WriteString(`<p style="margin:0 0 22px;">` + strings.ReplaceAll(escape(b.Text), "\n", "<br />") + `</p>`)
			plain += "\n\n" + b.Text
		case "image":
			content.WriteString(image(b.URL, b.Alt))
			plain += "\n\n" + b.Alt + ": " + b.URL
		case "button":
			content.WriteString(button(b.Text, b.URL))
			plain += "\n\n" + b.Text + ": " + b.URL
		case "divider":
			content.WriteString(`<hr style="border:0;border-top:1px solid ` + d.Text + `;opacity:.2;margin:28px 0;" />`)
		}
	}
	content.WriteString(button(label, actionURL))
	if actionURL != "" {
		plain += "\n\n" + label + ": " + actionURL
	}
	footer := "You receive Hackways news because promotional email is enabled for your account or contact."
	if !promotional {
		footer = "This is a service message from Hackways. Reply to hello@hackways.in for help."
		unsubscribe = ""
	}
	if test {
		footer = "This test was requested by an administrator. No subscription was added."
	}
	if unsubscribe != "" {
		content.WriteString(`<p style="margin:32px 0 8px;font-size:13px;"><a href="` + escape(unsubscribe) + `" style="color:` + d.Text + `;">Unsubscribe</a></p>`)
		plain += "\n\nUnsubscribe: " + unsubscribe
	}
	content.WriteString(`<p style="font-size:12px;margin:12px 0 0;">` + footer + `</p>`)
	padding, radius := "28px", "12px"
	if d.Layout == "plain" {
		padding = "0"
		radius = "0"
	}
	htmlBody := `<div style="display:none;max-height:0;overflow:hidden;">` + escape(d.Preheader) + `</div><div style="padding:28px 16px;background:` + d.Background + `;color:` + d.Text + `;font:16px/1.65 ` + font + `;text-align:` + d.Align + `;"><div style="max-width:616px;margin:0 auto;"><p style="font-weight:bold;margin:0 0 24px;color:` + d.Text + `;">` + escape(d.Brand) + `</p><div style="padding:` + padding + `;border-radius:` + radius + `;background:` + d.Surface + `;">` + content.String() + `</div></div></div>`
	return htmlBody, plain + "\n\n" + footer, nil
}
