// Package registrationforms validates owner-defined questions and attendee answers.
package registrationforms

import (
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Question struct {
	PublicCode  string     `json:"public_code"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	Kind        string     `json:"kind"`
	Required    bool       `json:"required"`
	Options     []string   `json:"options"`
	Placeholder string     `json:"placeholder"`
	Validation  Validation `json:"validation"`
}
type Validation struct {
	Enabled       *bool    `json:"enabled,omitempty"`
	AllowedDomain string   `json:"allowed_domain,omitempty"`
	MinLength     *int     `json:"min_length,omitempty"`
	MaxLength     *int     `json:"max_length,omitempty"`
	Min           *float64 `json:"min,omitempty"`
	Max           *float64 `json:"max,omitempty"`
	Integer       bool     `json:"integer,omitempty"`
	Message       string   `json:"message,omitempty"`
}
type Settings struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	SubmitLabel    string `json:"submit_label"`
	SuccessMessage string `json:"success_message"`
}
type Form struct {
	Revision     int64      `json:"revision"`
	NameRequired bool       `json:"name_required"`
	Questions    []Question `json:"questions"`
	Settings     Settings   `json:"settings"`
}
type Answer struct {
	Code     string   `json:"code"`
	Text     string   `json:"text"`
	Checked  bool     `json:"checked"`
	Selected []string `json:"selected,omitempty"`
}
type Snapshot struct {
	Answer
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	Section string `json:"section,omitempty"`
}

func ValidateForm(form Form) error {
	s := form.Settings
	if utf8.RuneCountInString(s.Title) > 180 || utf8.RuneCountInString(s.Description) > 1000 || utf8.RuneCountInString(s.SubmitLabel) > 40 || utf8.RuneCountInString(s.SuccessMessage) > 500 {
		return fmt.Errorf("Keep the title within 180 characters, introduction within 1000, button label within 40 and confirmation within 500")
	}
	return ValidateQuestions(form.Questions)
}

func ValidateQuestions(items []Question) error {
	if len(items) > 50 {
		return fmt.Errorf("Use up to 40 questions and 10 section breaks")
	}
	questions, sections := 0, 0
	seen := make(map[string]bool, len(items))
	for i := range items {
		q := &items[i]
		q.Label = strings.TrimSpace(q.Label)
		q.Description = strings.TrimSpace(q.Description)
		q.Placeholder = strings.TrimSpace(q.Placeholder)
		if q.Validation.Enabled != nil && !*q.Validation.Enabled {
			q.Validation = Validation{Enabled: q.Validation.Enabled}
		}
		q.Validation.AllowedDomain = strings.ToLower(strings.TrimSpace(q.Validation.AllowedDomain))
		q.Validation.Message = strings.TrimSpace(q.Validation.Message)
		if utf8.RuneCountInString(q.Placeholder) > 180 || utf8.RuneCountInString(q.Validation.Message) > 200 {
			return fmt.Errorf("Keep hints within 180 characters and validation messages within 200")
		}
		if utf8.RuneCountInString(q.Label) < 1 || utf8.RuneCountInString(q.Label) > 180 || utf8.RuneCountInString(q.Description) > 500 {
			return fmt.Errorf("Give every question a label of 1–180 characters and help text of up to 500 characters")
		}
		if q.PublicCode != "" {
			if seen[q.PublicCode] {
				return fmt.Errorf("A question cannot appear twice")
			}
			seen[q.PublicCode] = true
		}
		switch q.Kind {
		case "text", "textarea", "url", "number", "checkbox", "select", "phone", "email", "date", "time", "multiselect", "section":
		default:
			return fmt.Errorf("Choose a supported answer type for %s", q.Label)
		}
		if q.Kind == "section" {
			sections++
			q.Required = false
			q.Validation = Validation{}
			q.Placeholder = ""
		} else {
			questions++
		}
		if questions > 40 || sections > 10 {
			return fmt.Errorf("Use up to 40 questions and 10 section breaks")
		}
		if err := validateRules(*q); err != nil {
			return err
		}
		if q.Kind != "select" && q.Kind != "multiselect" {
			q.Options = []string{}
			continue
		}
		if len(q.Options) < 2 || len(q.Options) > 20 {
			return fmt.Errorf("Give %s between 2 and 20 choices", q.Label)
		}
		options := make(map[string]bool, len(q.Options))
		for j := range q.Options {
			q.Options[j] = strings.TrimSpace(q.Options[j])
			v := q.Options[j]
			if v == "" || utf8.RuneCountInString(v) > 100 || options[v] {
				return fmt.Errorf("Choices for %s must be different and contain 1–100 characters", q.Label)
			}
			options[v] = true
		}
	}
	return nil
}

func ValidateAnswers(form Form, name string, inputs []Answer) ([]Snapshot, error) {
	if form.NameRequired && strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("Your name is required. Update your account name before registering if you are signed in")
	}
	if len(inputs) > 40 {
		return nil, fmt.Errorf("Too many answers")
	}
	incoming := make(map[string]Answer, len(inputs))
	total := 0
	for _, a := range inputs {
		a.Text = strings.TrimSpace(a.Text)
		if _, exists := incoming[a.Code]; exists {
			return nil, fmt.Errorf("An answer was submitted twice")
		}
		if utf8.RuneCountInString(a.Text) > 2000 {
			return nil, fmt.Errorf("Keep each answer within 2,000 characters")
		}
		if len(a.Selected) > 20 {
			return nil, fmt.Errorf("Too many choices selected")
		}
		for _, choice := range a.Selected {
			if utf8.RuneCountInString(choice) > 100 {
				return nil, fmt.Errorf("Invalid selected choice")
			}
			total += len(choice)
		}
		total += len(a.Text)
		incoming[a.Code] = a
	}
	if total > 40000 {
		return nil, fmt.Errorf("Keep the combined answers within 40,000 bytes")
	}
	out := make([]Snapshot, 0, len(form.Questions))
	section := ""
	for _, q := range form.Questions {
		if q.Validation.Enabled != nil && !*q.Validation.Enabled {
			q.Validation = Validation{}
		}
		if q.Kind == "section" {
			section = q.Label
			continue
		}
		a, exists := incoming[q.PublicCode]
		delete(incoming, q.PublicCode)
		a.Code = q.PublicCode
		if q.Kind == "multiselect" {
			if a.Text != "" || a.Checked {
				return nil, fmt.Errorf("Choose from the options for %s", q.Label)
			}
			if q.Required && len(a.Selected) == 0 {
				return nil, fmt.Errorf("Please answer: %s", q.Label)
			}
			seen := map[string]bool{}
			for _, choice := range a.Selected {
				valid := false
				for _, opt := range q.Options {
					if choice == opt {
						valid = true
					}
				}
				if !valid || seen[choice] {
					return nil, fmt.Errorf("Choose different available options for %s", q.Label)
				}
				seen[choice] = true
			}
		} else if len(a.Selected) > 0 {
			return nil, fmt.Errorf("This question does not accept multiple choices: %s", q.Label)
		} else if q.Kind == "checkbox" {
			if a.Text != "" {
				return nil, fmt.Errorf("Use the checkbox for %s", q.Label)
			}
			if q.Required && !a.Checked {
				return nil, fmt.Errorf("Please confirm: %s", q.Label)
			}
		} else {
			if a.Checked {
				return nil, fmt.Errorf("Provide a written answer for %s", q.Label)
			}
			if q.Required && a.Text == "" {
				return nil, fmt.Errorf("Please answer: %s", q.Label)
			}
			if a.Text != "" {
				switch q.Kind {
				case "url":
					u, err := url.Parse(a.Text)
					if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
						return nil, fmt.Errorf("Use a full https:// link for %s", q.Label)
					}
					if domain := q.Validation.AllowedDomain; domain != "" {
						host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
						if host != domain && !strings.HasSuffix(host, "."+domain) {
							return nil, answerError(q, "Use a link from "+domain)
						}
					}
				case "number":
					n, err := strconv.ParseFloat(a.Text, 64)
					if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
						return nil, fmt.Errorf("Enter a number for %s", q.Label)
					}
					if q.Validation.Min != nil && n < *q.Validation.Min || q.Validation.Max != nil && n > *q.Validation.Max || q.Validation.Integer && math.Trunc(n) != n {
						return nil, answerError(q, "Enter a number within the allowed range")
					}
				case "email":
					address, err := mail.ParseAddress(a.Text)
					if err != nil || address.Address != a.Text || len(a.Text) > 254 {
						return nil, answerError(q, "Enter a valid email address")
					}
				case "date":
					if _, err := time.Parse("2006-01-02", a.Text); err != nil {
						return nil, answerError(q, "Enter a valid calendar date")
					}
				case "time":
					if _, err := time.Parse("15:04", a.Text); err != nil {
						return nil, answerError(q, "Enter a valid time")
					}
				case "select":
					valid := false
					for _, opt := range q.Options {
						if a.Text == opt {
							valid = true
						}
					}
					if !valid {
						return nil, fmt.Errorf("Choose an available option for %s", q.Label)
					}
				case "phone":
					digits := 0
					for _, r := range a.Text {
						if r >= '0' && r <= '9' {
							digits++
						} else if !strings.ContainsRune("+()- .", r) {
							return nil, fmt.Errorf("Enter a phone number for %s", q.Label)
						}
					}
					if digits < 5 || digits > 20 {
						return nil, fmt.Errorf("Enter a phone number for %s", q.Label)
					}
				}
			}
		}
		if q.Kind != "checkbox" && q.Kind != "multiselect" && a.Text != "" {
			length := utf8.RuneCountInString(a.Text)
			if q.Validation.MinLength != nil && length < *q.Validation.MinLength || q.Validation.MaxLength != nil && length > *q.Validation.MaxLength {
				return nil, answerError(q, "Use the required answer length")
			}
		}
		if exists || q.Required {
			out = append(out, Snapshot{Answer: a, Label: q.Label, Kind: q.Kind, Section: section})
		}
	}
	if len(incoming) > 0 {
		return nil, fmt.Errorf("The questions changed. Reload the registration form before submitting")
	}
	return out, nil
}

func answerError(q Question, fallback string) error {
	message := strings.TrimSpace(q.Validation.Message)
	if message == "" {
		message = fallback
	}
	return fmt.Errorf("%s: %s", q.Label, message)
}

var websiteDomain = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+$`)

func validateRules(q Question) error {
	v := q.Validation
	if v.AllowedDomain != "" && (q.Kind != "url" || len(v.AllowedDomain) > 253 || !websiteDomain.MatchString(v.AllowedDomain)) {
		return fmt.Errorf("Use a website domain such as linkedin.com for website questions only")
	}
	if v.MinLength != nil && (*v.MinLength < 0 || *v.MinLength > 2000) || v.MaxLength != nil && (*v.MaxLength < 1 || *v.MaxLength > 2000) || v.MinLength != nil && v.MaxLength != nil && *v.MinLength > *v.MaxLength {
		return fmt.Errorf("Choose valid answer lengths for %s (up to 2000 characters)", q.Label)
	}
	if v.Min != nil && (math.IsNaN(*v.Min) || math.IsInf(*v.Min, 0)) || v.Max != nil && (math.IsNaN(*v.Max) || math.IsInf(*v.Max, 0)) || v.Min != nil && v.Max != nil && *v.Min > *v.Max {
		return fmt.Errorf("Choose a valid number range for %s", q.Label)
	}
	if q.Kind != "number" && (v.Min != nil || v.Max != nil || v.Integer) {
		return fmt.Errorf("Numeric validation only applies to number questions")
	}
	if (q.Kind == "section" || q.Kind == "checkbox" || q.Kind == "select" || q.Kind == "multiselect") && (v.MinLength != nil || v.MaxLength != nil) {
		return fmt.Errorf("Answer lengths only apply to written questions")
	}
	return nil
}
