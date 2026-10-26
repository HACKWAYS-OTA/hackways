package programs

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type Question struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	Help      string   `json:"help"`
	Kind      string   `json:"kind"`
	Required  bool     `json:"required"`
	MaxLength int      `json:"max_length"`
	Options   []string `json:"options"`
	Domain    string   `json:"domain"`
}
type Program struct {
	ID               pgtype.UUID        `json:"id"`
	Slug             string             `json:"slug"`
	Title            string             `json:"title"`
	Summary          string             `json:"summary"`
	Overview         string             `json:"overview"`
	Responsibilities string             `json:"responsibilities"`
	Recognition      string             `json:"recognition"`
	Status           string             `json:"status"`
	MinimumAge       int32              `json:"minimum_age"`
	Questions        json.RawMessage    `json:"questions"`
	OpensAt          pgtype.Timestamptz `json:"opens_at"`
	ClosesAt         pgtype.Timestamptz `json:"closes_at"`
	Revision         int64              `json:"revision"`
}

func view(p db.Programs) Program {
	return Program{p.ID, p.Slug, p.Title, p.Summary, p.Overview, p.Responsibilities, p.Recognition, p.Status, p.MinimumAge, json.RawMessage(p.Questions), p.OpensAt, p.ClosesAt, p.Revision}
}

type programBody struct {
	Slug             string     `json:"slug"`
	Title            string     `json:"title"`
	Summary          string     `json:"summary"`
	Overview         string     `json:"overview"`
	Responsibilities string     `json:"responsibilities"`
	Recognition      string     `json:"recognition"`
	Status           string     `json:"status"`
	MinimumAge       int32      `json:"minimum_age"`
	Questions        []Question `json:"questions"`
	OpensAt          *time.Time `json:"opens_at"`
	ClosesAt         *time.Time `json:"closes_at"`
	Revision         int64      `json:"revision"`
}
type profileBody struct {
	FullName        string            `json:"full_name"`
	City            string            `json:"city"`
	Phone           string            `json:"phone"`
	Biography       string            `json:"biography"`
	ProfileURL      string            `json:"profile_url"`
	Proposal        string            `json:"proposal"`
	Age             int32             `json:"age"`
	AgeConfirmed    bool              `json:"age_confirmed"`
	Answers         map[string]string `json:"answers"`
	Revision        int64             `json:"revision"`
	ProgramRevision int64             `json:"program_revision"`
}

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}[a-z0-9]$`)
var questionPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`)
var indianPhonePattern = regexp.MustCompile(`^[6-9][0-9]{9}$`)
var indianPrefixedPhonePattern = regexp.MustCompile(`^91[6-9][0-9]{9}$`)
var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func length(s string, min, max int) bool {
	return utf8.ValidString(s) && !strings.ContainsRune(s, 0) && utf8.RuneCountInString(s) >= min && utf8.RuneCountInString(s) <= max
}
func validURL(raw, domain string) bool {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	return domain == "" || host == domain || strings.HasSuffix(host, "."+domain)
}
func validateQuestions(qs []Question) error {
	if len(qs) > 10 {
		return errors.New("Use at most 10 additional questions")
	}
	seen := map[string]bool{}
	for _, q := range qs {
		if !questionPattern.MatchString(q.ID) || seen[q.ID] || !length(q.Label, 3, 200) || !length(q.Help, 0, 500) || q.MaxLength < 1 || q.MaxLength > 5000 {
			return errors.New("Use unique question identifiers, clear labels and answer limits between 1 and 5,000")
		}
		seen[q.ID] = true
		if q.Kind != "text" && q.Kind != "long_text" && q.Kind != "url" && q.Kind != "choice" {
			return errors.New("Choose a supported question type")
		}
		if q.Kind == "choice" {
			if len(q.Options) < 2 || len(q.Options) > 20 {
				return errors.New("Choice questions need 2 to 20 options")
			}
			options := map[string]bool{}
			for _, v := range q.Options {
				if !length(v, 1, 150) || options[v] {
					return errors.New("Use distinct nonempty choices")
				}
				options[v] = true
			}
		} else if len(q.Options) > 0 {
			return errors.New("Only choice questions can have options")
		}
		if q.Domain != "" {
			u, e := url.Parse("https://" + q.Domain)
			if q.Kind != "url" || e != nil || u.Hostname() != q.Domain || u.Path != "" || u.Port() != "" || !strings.Contains(q.Domain, ".") {
				return errors.New("Use a hostname such as linkedin.com for the website domain")
			}
		}
	}
	return nil
}
func (b *programBody) validate(create bool) error {
	b.Slug = strings.ToLower(strings.TrimSpace(b.Slug))
	b.Title = strings.TrimSpace(b.Title)
	b.Summary = strings.TrimSpace(b.Summary)
	b.Overview = strings.TrimSpace(b.Overview)
	b.Responsibilities = strings.TrimSpace(b.Responsibilities)
	b.Recognition = strings.TrimSpace(b.Recognition)
	if create && b.Slug == "new" {
		return errors.New("Use a program URL name other than 'new'")
	}
	if create && !slugPattern.MatchString(b.Slug) {
		return errors.New("Use a URL name of 3 to 64 lowercase letters, numbers and hyphens")
	}
	if !length(b.Title, 3, 120) || !length(b.Summary, 10, 400) || !length(b.Overview, 10, 6000) || !length(b.Responsibilities, 10, 6000) || !length(b.Recognition, 10, 6000) {
		return errors.New("Complete the program title, summary, overview, responsibilities and recognition within their character limits")
	}
	if b.Status != "draft" && b.Status != "open" && b.Status != "closed" && b.Status != "archived" {
		return errors.New("Choose a valid program status")
	}
	if b.MinimumAge < 18 || b.MinimumAge > 100 {
		return errors.New("Minimum age must be between 18 and 100")
	}
	if b.OpensAt != nil && b.ClosesAt != nil && !b.ClosesAt.After(*b.OpensAt) {
		return errors.New("Applications must close after they open")
	}
	for i := range b.Questions {
		b.Questions[i].Label = strings.TrimSpace(b.Questions[i].Label)
		b.Questions[i].Help = strings.TrimSpace(b.Questions[i].Help)
		b.Questions[i].Domain = strings.ToLower(strings.TrimSpace(b.Questions[i].Domain))
		for j := range b.Questions[i].Options {
			b.Questions[i].Options[j] = strings.TrimSpace(b.Questions[i].Options[j])
		}
	}
	return validateQuestions(b.Questions)
}
func (b *profileBody) validate(p db.Programs, submit bool) error {
	b.FullName = strings.TrimSpace(b.FullName)
	b.City = strings.TrimSpace(b.City)
	b.Phone = strings.TrimSpace(b.Phone)
	b.Phone = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "").Replace(b.Phone)
	if indianPhonePattern.MatchString(b.Phone) {
		b.Phone = "+91" + b.Phone
	} else if indianPrefixedPhonePattern.MatchString(b.Phone) {
		b.Phone = "+" + b.Phone
	}
	b.ProfileURL = strings.TrimSpace(b.ProfileURL)
	if !length(b.FullName, 0, 120) || !length(b.City, 0, 150) || !length(b.Phone, 0, 32) || !length(b.Biography, 0, 2000) || !length(b.ProfileURL, 0, 2048) || !length(b.Proposal, 0, 5000) || b.Age < 0 || b.Age > 120 {
		return errors.New("Some profile answers exceed their character limits")
	}
	if b.ProfileURL != "" && !validURL(b.ProfileURL, "") {
		return errors.New("Use a complete HTTPS profile URL")
	}
	if submit && (!length(b.FullName, 2, 120) || !length(b.City, 2, 150) || !length(strings.TrimSpace(b.Proposal), 30, 5000) || !b.AgeConfirmed || b.Age < p.MinimumAge) {
		return errors.New("Add your name, city, age confirmation and an event proposal of at least 30 characters; you must meet the program's minimum age")
	}
	if submit && p.Slug == "ambassadors" {
		if !validURL(b.ProfileURL, "linkedin.com") {
			return errors.New("Add your LinkedIn profile using linkedin.com or one of its subdomains")
		}
		if !phonePattern.MatchString(b.Phone) {
			return errors.New("Enter a valid 10-digit Indian WhatsApp number, or include the country code for an international number")
		}
		if !length(strings.TrimSpace(b.Biography), 10, 2000) {
			return errors.New("Add a short introduction of at least 10 characters")
		}
	}
	var qs []Question
	if json.Unmarshal(p.Questions, &qs) != nil {
		return errors.New("Program questions could not be read")
	}
	if len(b.Answers) > 10 {
		return errors.New("Too many answers")
	}
	known := map[string]bool{}
	for _, q := range qs {
		known[q.ID] = true
		a := strings.TrimSpace(b.Answers[q.ID])
		if !length(a, 0, q.MaxLength) {
			return errors.New(q.Label + ": answer exceeds its limit")
		}
		if submit && q.Required && a == "" {
			return errors.New(q.Label + ": answer is required")
		}
		if a != "" && q.Kind == "url" && !validURL(a, q.Domain) {
			return errors.New(q.Label + ": use a valid HTTPS URL on the allowed domain")
		}
		if a != "" && q.Kind == "choice" {
			found := false
			for _, option := range q.Options {
				if option == a {
					found = true
				}
			}
			if !found {
				return errors.New(q.Label + ": select a listed option")
			}
		}
	}
	for key := range b.Answers {
		if !known[key] {
			return errors.New("The application questions changed. Reload the current form before saving")
		}
	}
	return nil
}
func timestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}
func profile(a db.ProgramApplications) profileBody {
	var answers map[string]string
	_ = json.Unmarshal(a.Answers, &answers)
	return profileBody{FullName: a.FullName, City: a.City, Phone: a.Phone, Biography: a.Biography, ProfileURL: a.ProfileUrl, Proposal: a.Proposal, Age: a.Age, AgeConfirmed: a.AgeConfirmed, Answers: answers, Revision: a.Revision}
}
