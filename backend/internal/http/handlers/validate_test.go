package handlers

import "testing"

func TestValidateOptionalHTTPURL(t *testing.T) {
	t.Parallel()

	allowed := []*string{
		nil,
		strPtr(""),
		strPtr("   "),
		strPtr("http://example.com"),
		strPtr("https://example.com/path?q=1"),
		strPtr("  https://example.com  "),
	}
	for _, v := range allowed {
		if err := validateOptionalHTTPURL(v); err != nil {
			got := "<nil>"
			if v != nil {
				got = *v
			}
			t.Errorf("expected %q to be allowed, got %v", got, err)
		}
	}

	rejected := []string{
		"javascript:alert(1)",
		"JavaScript:alert(1)",
		"data:text/html;base64,PHNjcmlwdD4=",
		"vbscript:msgbox(1)",
		"ftp://example.com",
		"mailto:user@example.com",
		"//evil.com",
		"example.com",
	}
	for _, v := range rejected {
		if err := validateOptionalHTTPURL(strPtr(v)); err == nil {
			t.Errorf("expected %q to be rejected", v)
		}
	}
}
