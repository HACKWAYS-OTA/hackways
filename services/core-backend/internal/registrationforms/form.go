// Package registrationforms validates owner-defined questions and attendee answers.
package registrationforms

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Question struct {
	PublicCode  string   `json:"public_code"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Kind        string   `json:"kind"`
	Required    bool     `json:"required"`
	Options     []string `json:"options"`
}
type Form struct {
	Revision     int64      `json:"revision"`
	NameRequired bool       `json:"name_required"`
	Questions    []Question `json:"questions"`
}
type Answer struct {
	Code    string `json:"code"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}
type Snapshot struct {
	Answer
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

func ValidateQuestions(items []Question) error {
	if len(items) > 20 {
		return fmt.Errorf("Use up to 20 registration questions")
	}
	seen := make(map[string]bool, len(items))
	for i := range items {
		q := &items[i]
		q.Label = strings.TrimSpace(q.Label)
		q.Description = strings.TrimSpace(q.Description)
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
		case "text", "textarea", "url", "number", "checkbox", "select", "phone":
		default:
			return fmt.Errorf("Choose a supported answer type for %s", q.Label)
		}
		if q.Kind != "select" {
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
	if len(inputs) > 20 {
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
		total += len(a.Text)
		incoming[a.Code] = a
	}
	if total > 20000 {
		return nil, fmt.Errorf("Keep the combined answers within 20,000 bytes")
	}
	out := make([]Snapshot, 0, len(form.Questions))
	for _, q := range form.Questions {
		a, exists := incoming[q.PublicCode]
		delete(incoming, q.PublicCode)
		a.Code = q.PublicCode
		if q.Kind == "checkbox" {
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
				case "number":
					n, err := strconv.ParseFloat(a.Text, 64)
					if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
						return nil, fmt.Errorf("Enter a number for %s", q.Label)
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
		if exists || q.Required {
			out = append(out, Snapshot{a, q.Label, q.Kind})
		}
	}
	if len(incoming) > 0 {
		return nil, fmt.Errorf("The questions changed. Reload the registration form before submitting")
	}
	return out, nil
}
