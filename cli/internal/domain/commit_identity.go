package domain

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxCommitTimezoneOffsetMinutes = 23*60 + 59

// CommitIdentity stores commit author or committer metadata.
type CommitIdentity struct {
	name                  string
	email                 string
	timestamp             int64
	timezoneOffsetMinutes int
}

// NewCommitIdentity validates and constructs a commit identity.
func NewCommitIdentity(
	name,
	email string,
	timestamp int64,
	timezoneOffsetMinutes int,
) (CommitIdentity, error) {
	identity := CommitIdentity{
		name:                  name,
		email:                 email,
		timestamp:             timestamp,
		timezoneOffsetMinutes: timezoneOffsetMinutes,
	}
	if err := validateCommitIdentity(identity); err != nil {
		return CommitIdentity{}, fmt.Errorf(
			"create commit identity: %w",
			err,
		)
	}
	return identity, nil
}

// Name returns the identity's display name.
func (i CommitIdentity) Name() string {
	return i.name
}

// Email returns the identity's email address.
func (i CommitIdentity) Email() string {
	return i.email
}

// Timestamp returns the identity's Unix timestamp in seconds.
func (i CommitIdentity) Timestamp() int64 {
	return i.timestamp
}

// TimezoneOffsetMinutes returns the identity's UTC offset in minutes.
func (i CommitIdentity) TimezoneOffsetMinutes() int {
	return i.timezoneOffsetMinutes
}

func validateCommitIdentity(identity CommitIdentity) error {
	if err := validateCommitIdentityName(identity.name); err != nil {
		return err
	}
	if err := validateIdentityEmail(identity.email); err != nil {
		return err
	}
	if identity.timezoneOffsetMinutes < -maxCommitTimezoneOffsetMinutes ||
		identity.timezoneOffsetMinutes > maxCommitTimezoneOffsetMinutes {
		return fmt.Errorf(
			"timezone offset %d minutes is outside the supported range",
			identity.timezoneOffsetMinutes,
		)
	}
	return nil
}

func validateCommitIdentityName(name string) error {
	if name == "" {
		return fmt.Errorf("name is empty")
	}
	if !utf8.ValidString(name) {
		return fmt.Errorf(
			"name is not valid UTF-8",
		)
	}
	if name != strings.TrimSpace(name) {
		return fmt.Errorf(
			"name has leading or trailing whitespace",
		)
	}
	if strings.ContainsAny(name, "<>") {
		return fmt.Errorf(
			"name contains reserved delimiters",
		)
	}
	if containsASCIIControl(name) {
		return fmt.Errorf(
			"name contains an ASCII control character",
		)
	}
	return nil
}

func validateIdentityEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is empty")
	}
	if !utf8.ValidString(email) {
		return fmt.Errorf("email is not valid UTF-8")
	}
	if strings.IndexFunc(email, unicode.IsSpace) >= 0 {
		return fmt.Errorf("email contains whitespace")
	}
	if strings.ContainsAny(email, "<>") {
		return fmt.Errorf(
			"email contains reserved delimiters",
		)
	}
	if containsASCIIControl(email) {
		return fmt.Errorf(
			"email contains an ASCII control character",
		)
	}
	if strings.Count(email, "@") != 1 {
		return fmt.Errorf(
			"email must contain exactly one '@'",
		)
	}

	local, domain, _ := strings.Cut(email, "@")
	if local == "" || domain == "" {
		return fmt.Errorf("email must contain non-empty local and domain parts")
	}
	return nil
}

func containsASCIIControl(value string) bool {
	for _, char := range []byte(value) {
		if char < 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}
