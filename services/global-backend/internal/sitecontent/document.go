package sitecontent

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// The document is structured content, never arbitrary HTML or executable styling.
type DocumentNode struct {
	Type    string         `json:"type"`
	Text    string         `json:"text,omitempty"`
	Attrs   DocumentAttrs  `json:"attrs,omitempty"`
	Marks   []DocumentMark `json:"marks,omitempty"`
	Content []DocumentNode `json:"content,omitempty"`
}
type DocumentMark struct {
	Type  string        `json:"type"`
	Attrs DocumentAttrs `json:"attrs,omitempty"`
}
type DocumentAttrs struct {
	Level           int    `json:"level,omitempty"`
	Start           int    `json:"start,omitempty"`
	Language        string `json:"language,omitempty"`
	TextAlign       string `json:"textAlign,omitempty"`
	Align           string `json:"align,omitempty"`
	ListType        string `json:"type,omitempty"`
	Color           string `json:"color,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	FontFamily      string `json:"fontFamily,omitempty"`
	FontSize        string `json:"fontSize,omitempty"`
	LineHeight      string `json:"lineHeight,omitempty"`
	Href            string `json:"href,omitempty"`
	Target          string `json:"target,omitempty"`
	Rel             string `json:"rel,omitempty"`
	Class           string `json:"class,omitempty"`
	Src             string `json:"src,omitempty"`
	Alt             string `json:"alt,omitempty"`
	Title           string `json:"title,omitempty"`
	Width           *int   `json:"width,omitempty"`
	Height          *int   `json:"height,omitempty"`
	Colspan         int    `json:"colspan,omitempty"`
	Rowspan         int    `json:"rowspan,omitempty"`
	Colwidth        []int  `json:"colwidth,omitempty"`
	Checked         bool   `json:"checked,omitempty"`
}

var documentColor = regexp.MustCompile(`^(?:#[a-fA-F0-9]{3}|#[a-fA-F0-9]{6}|#[a-fA-F0-9]{8})$`)
var documentSize = regexp.MustCompile(`^(?:[89]|[1-6][0-9]|7[0-2])(?:px|pt)$`)
var documentLanguage = regexp.MustCompile(`^[a-zA-Z0-9_+-]{0,40}$`)
var documentFonts = map[string]bool{"Google Sans Flex": true, "FG Futurist": true, "Arial": true, "Helvetica": true, "Georgia": true, "Times New Roman": true, "Courier New": true, "Verdana": true, "Tahoma": true, "Trebuchet MS": true, "monospace": true, "sans-serif": true, "serif": true}

func SafeDocumentURL(raw string, image bool) bool {
	if raw == "" || len(raw) > 2048 || strings.ContainsAny(raw, "\x00\r\n\t") || strings.HasPrefix(raw, "//") {
		return false
	}
	parsed, e := url.Parse(raw)
	if e != nil || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	if image {
		if parsed.Scheme != "https" && !(parsed.Scheme == "" && strings.HasPrefix(raw, "/")) {
			return false
		}
		for key := range parsed.Query() {
			if strings.EqualFold(key, "sig") {
				return false
			}
		}
	} else if parsed.Scheme != "https" && parsed.Scheme != "http" && parsed.Scheme != "mailto" && parsed.Scheme != "tel" && !(parsed.Scheme == "" && strings.HasPrefix(raw, "/")) {
		return false
	}
	return parsed.Scheme == "" || parsed.Scheme == "mailto" || parsed.Scheme == "tel" || parsed.Hostname() != ""
}
func validateDocumentAttrs(a *DocumentAttrs) error {
	if a.Level < 0 || a.Level > 6 || a.Start < 0 || a.Start > 100000 || a.Colspan < 0 || a.Colspan > 20 || a.Rowspan < 0 || a.Rowspan > 100 || len(a.Colwidth) > 20 {
		return errors.New("Invalid document layout")
	}
	if (a.TextAlign != "" && a.TextAlign != "left" && a.TextAlign != "center" && a.TextAlign != "right" && a.TextAlign != "justify") || (a.Align != "" && a.Align != "left" && a.Align != "center" && a.Align != "right") || (a.ListType != "" && a.ListType != "1" && a.ListType != "a" && a.ListType != "A" && a.ListType != "i" && a.ListType != "I") {
		return errors.New("Invalid text alignment")
	}
	for _, value := range []string{a.TextAlign, a.Color, a.BackgroundColor, a.FontFamily, a.FontSize, a.LineHeight, a.Href, a.Target, a.Rel, a.Class, a.Src, a.Alt, a.Title, a.Language} {
		if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
			return errors.New("Invalid document text")
		}
	}
	for _, color := range []*string{&a.Color, &a.BackgroundColor} {
		if strings.HasPrefix(*color, "rgb(") {
			var red, green, blue int
			if count, e := fmt.Sscanf(*color, "rgb(%d, %d, %d)", &red, &green, &blue); e == nil && count == 3 && red >= 0 && red <= 255 && green >= 0 && green <= 255 && blue >= 0 && blue <= 255 {
				*color = fmt.Sprintf("#%02x%02x%02x", red, green, blue)
			}
		}
		if *color != "" && !documentColor.MatchString(*color) {
			return errors.New("Use a supported text colour")
		}
	}
	if a.FontFamily != "" {
		a.FontFamily = strings.Trim(strings.Split(a.FontFamily, ",")[0], "\"' ")
		if !documentFonts[a.FontFamily] {
			return errors.New("Use a supported document font")
		}
	}
	if a.FontSize != "" && !documentSize.MatchString(a.FontSize) {
		return errors.New("Use a font size from 8 to 72 pixels or points")
	}
	if a.LineHeight != "" {
		value, e := strconv.ParseFloat(a.LineHeight, 64)
		if e != nil || value < 1 || value > 3 {
			return errors.New("Invalid line spacing")
		}
	}
	for _, value := range []*int{a.Width, a.Height} {
		if value != nil && (*value < 1 || *value > 20000) {
			return errors.New("Invalid image dimensions")
		}
	}
	for _, value := range a.Colwidth {
		if value < 0 || value > 20000 {
			return errors.New("Invalid table width")
		}
	}
	if !documentLanguage.MatchString(a.Language) || utf8.RuneCountInString(a.Alt) > 500 || utf8.RuneCountInString(a.Title) > 500 || len(a.Class) > 200 || len(a.Rel) > 200 {
		return errors.New("Document attributes exceed their limits")
	}
	if a.Target != "" && a.Target != "_blank" && a.Target != "_self" {
		return errors.New("Invalid link target")
	}
	return nil
}

// ValidateDocument returns a plain Markdown fallback derived from the actual
// validated document, rather than trusting a separate client-supplied body.
func ValidateDocument(root *DocumentNode) (string, bool, error) {
	if root == nil || root.Type != "doc" {
		return "", false, errors.New("Use a valid notice document")
	}
	nodes, characters := 0, 0
	hasContent := false
	var walk func(*DocumentNode, int, string) (string, error)
	walk = func(node *DocumentNode, depth int, parent string) (string, error) {
		nodes++
		characters += utf8.RuneCountInString(node.Text)
		if nodes > 10000 || characters > 100000 || depth > 32 || !utf8.ValidString(node.Text) || strings.ContainsRune(node.Text, 0) {
			return "", errors.New("The notice document exceeds its supported size")
		}
		if e := validateDocumentAttrs(&node.Attrs); e != nil {
			return "", e
		}
		inline := node.Type == "text" || node.Type == "hardBreak"
		switch parent {
		case "paragraph", "heading", "codeBlock":
			if !inline {
				return "", errors.New("Invalid paragraph content")
			}
		case "bulletList", "orderedList":
			if node.Type != "listItem" {
				return "", errors.New("Invalid list content")
			}
		case "table":
			if node.Type != "tableRow" {
				return "", errors.New("Invalid table content")
			}
		case "tableRow":
			if node.Type != "tableCell" && node.Type != "tableHeader" {
				return "", errors.New("Invalid table row")
			}
		default:
			if inline && parent != "" {
				return "", errors.New("Text must belong to a paragraph")
			}
		}
		switch node.Type {
		case "doc":
			if parent != "" {
				return "", errors.New("Invalid nested document")
			}
		case "paragraph", "heading", "blockquote", "bulletList", "orderedList", "listItem", "codeBlock", "table", "tableRow", "tableCell", "tableHeader":
		case "text":
			if node.Text != "" {
				hasContent = hasContent || strings.TrimSpace(node.Text) != ""
			}
		case "hardBreak", "horizontalRule":
		case "image":
			if !SafeDocumentURL(node.Attrs.Src, true) {
				return "", errors.New("Use an uploaded public image URL")
			}
			hasContent = true
		default:
			return "", errors.New("Unsupported document content")
		}
		if (node.Type == "text" || node.Type == "image" || node.Type == "hardBreak" || node.Type == "horizontalRule") && len(node.Content) > 0 {
			return "", errors.New("Invalid leaf content")
		}
		if node.Type == "heading" && (node.Attrs.Level < 1 || node.Attrs.Level > 6) {
			return "", errors.New("Invalid heading level")
		}
		for i := range node.Marks {
			mark := &node.Marks[i]
			if e := validateDocumentAttrs(&mark.Attrs); e != nil {
				return "", e
			}
			switch mark.Type {
			case "bold", "italic", "underline", "strike", "code", "textStyle":
			case "link":
				if !SafeDocumentURL(mark.Attrs.Href, false) {
					return "", errors.New("Use a valid public link without a fragment")
				}
			default:
				return "", errors.New("Unsupported text formatting")
			}
		}
		if len(node.Marks) > 12 {
			return "", errors.New("Too many text styles")
		}
		var parts []string
		for i := range node.Content {
			part, e := walk(&node.Content[i], depth+1, node.Type)
			if e != nil {
				return "", e
			}
			parts = append(parts, part)
		}
		content := strings.Join(parts, "")
		switch node.Type {
		case "text":
			return node.Text, nil
		case "hardBreak":
			return "\n", nil
		case "image":
			return fmt.Sprintf("\n\n![%s](%s)\n\n", strings.ReplaceAll(node.Attrs.Alt, "]", "\\]"), node.Attrs.Src), nil
		case "paragraph", "heading", "codeBlock", "listItem", "blockquote", "tableRow":
			return content + "\n\n", nil
		case "horizontalRule":
			return "\n---\n", nil
		}
		return content, nil
	}
	body, e := walk(root, 0, "")
	if e != nil {
		return "", false, e
	}
	if len(body) > 500000 {
		return "", false, errors.New("The notice document exceeds its supported size")
	}
	return strings.TrimSpace(body), hasContent, nil
}

// DocumentSummary is extracted from authored text; no placeholder copy is stored.
func DocumentSummary(root *DocumentNode, fallback string) string {
	var words []string
	var walk func(*DocumentNode)
	walk = func(node *DocumentNode) {
		if node.Type == "text" {
			words = append(words, node.Text)
		}
		if node.Type == "image" && node.Attrs.Alt != "" {
			words = append(words, node.Attrs.Alt)
		}
		for i := range node.Content {
			walk(&node.Content[i])
		}
	}
	if root != nil {
		walk(root)
	}
	result := strings.Join(strings.Fields(strings.Join(words, " ")), " ")
	if result == "" {
		result = fallback
	}
	runes := []rune(result)
	if len(runes) > 180 {
		result = string(runes[:177]) + "..."
	}
	return result
}
