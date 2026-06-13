package handlers

import (
	"errors"
	"net/url"
	"strings"
)

// errUnsafeURL is returned when a user-provided URL is not a plain http(s) URL.
var errUnsafeURL = errors.New("url must be an http or https URL")

// validateOptionalHTTPURL rejects a non-empty URL whose scheme is not http(s).
// It blocks javascript:, data:, and similar schemes from being stored and later
// rendered into an href (stored-XSS defense). A nil or blank value is allowed.
func validateOptionalHTTPURL(value *string) error {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return errUnsafeURL
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return nil
	default:
		return errUnsafeURL
	}
}

// validateOptionalHTTPURLs validates each value, returning the first failure.
func validateOptionalHTTPURLs(values ...*string) error {
	for _, v := range values {
		if err := validateOptionalHTTPURL(v); err != nil {
			return err
		}
	}
	return nil
}
