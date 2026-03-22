package urlpolicy

import "testing"

func TestValidateAllowsExactHostForAvailableService(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"youtube"})

	got, ok := validator.Validate("https://youtube.com/watch?v=abc")
	if !ok {
		t.Fatal("expected url to be allowed")
	}
	if got != "https://youtube.com/watch?v=abc" {
		t.Fatalf("unexpected normalized url: %q", got)
	}
}

func TestValidateAllowsSubdomainForAvailableService(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"youtube"})

	got, ok := validator.Validate("https://m.youtube.com/watch?v=abc")
	if !ok {
		t.Fatal("expected subdomain url to be allowed")
	}
	if got != "https://m.youtube.com/watch?v=abc" {
		t.Fatalf("unexpected normalized url: %q", got)
	}
}

func TestValidateRejectsURLWithoutScheme(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"youtube"})

	if _, ok := validator.Validate("m.youtube.com/watch?v=abc"); ok {
		t.Fatal("expected url without scheme to be rejected")
	}
}

func TestValidateRejectsHostForUnavailableService(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"reddit"})

	if _, ok := validator.Validate("https://youtube.com/watch?v=abc"); ok {
		t.Fatal("expected url outside derived allowlist to be rejected")
	}
}

func TestValidateRejectsLookalikeHost(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"youtube"})

	if _, ok := validator.Validate("https://youtube.com.evil.example/video"); ok {
		t.Fatal("expected lookalike host to be rejected")
	}
}

func TestValidateRejectsIPAddress(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"youtube"})

	if _, ok := validator.Validate("http://127.0.0.1/video"); ok {
		t.Fatal("expected ip address to be rejected")
	}
}

func TestValidateRejectsUserInfo(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"youtube"})

	if _, ok := validator.Validate("https://user:pass@youtube.com/video"); ok {
		t.Fatal("expected url with userinfo to be rejected")
	}
}

func TestValidateRejectsWhenNoKnownServicesAvailable(t *testing.T) {
	t.Parallel()

	validator := NewURLValidator([]string{"unknown-service"})

	if _, ok := validator.Validate("https://youtube.com/watch?v=abc"); ok {
		t.Fatal("expected validation to fail with empty derived allowlist")
	}
}

func TestBuildAllowlistIncludesExpectedDomains(t *testing.T) {
	t.Parallel()

	allowlist := buildAllowlist([]string{"youtube", "twitter"})

	expected := []string{"youtu.be", "youtube.com", "m.youtube.com", "music.youtube.com", "twitter.com", "x.com"}
	for _, domain := range expected {
		found := false
		for _, candidate := range allowlist {
			if candidate == domain {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected domain %q to be present in allowlist %#v", domain, allowlist)
		}
	}
}
