package commands

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestNormalizeServerURL tests URL normalization with table-driven tests.
func TestNormalizeServerURL_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"http preserved", "http://localhost:8080", "http://localhost:8080"},
		{"https preserved", "https://api.shoehorn.dev", "https://api.shoehorn.dev"},
		{"adds https when no scheme", "api.shoehorn.dev", "https://api.shoehorn.dev"},
		{"strips single trailing slash", "http://localhost:8080/", "http://localhost:8080"},
		{"strips multiple trailing slashes", "http://localhost:8080///", "http://localhost:8080"},
		{"no-scheme with trailing slash", "api.shoehorn.dev/", "https://api.shoehorn.dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeServerURL(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeServerURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestHasScheme tests scheme detection.
func TestHasScheme_TableDriven(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"http://localhost", true},
		{"https://api.example.com", true},
		{"api.example.com", false},
		{"ftp://other", false},
		{"", false},
		{"http://", true},
		{"h", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := hasScheme(tt.input)
			if got != tt.want {
				t.Errorf("hasScheme(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestFormatDuration tests human-readable duration formatting.
func TestFormatDuration_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		contains string
	}{
		{"seconds", 30, "seconds"},
		{"minutes", 300, "minutes"},
		{"hours", 7200, "hours"},
		{"days", 172800, "days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := formatDuration(time.Duration(tt.seconds) * time.Second)
			if !strings.Contains(d, tt.contains) {
				t.Errorf("formatDuration(%ds) = %q, want to contain %q", tt.seconds, d, tt.contains)
			}
		})
	}
}

// ─── Security: SHOEHORN_TOKEN env var ────────────────────────────────────────

// TestResolveToken_EnvVar tests that SHOEHORN_TOKEN env var is picked up.
func TestResolveToken_EnvVar(t *testing.T) {
	t.Setenv("SHOEHORN_TOKEN", "shp_from_env")
	token, source := resolveToken("")
	if token != "shp_from_env" {
		t.Errorf("resolveToken() token = %q, want %q", token, "shp_from_env")
	}
	if source != "env" {
		t.Errorf("resolveToken() source = %q, want %q", source, "env")
	}
}

// TestResolveToken_FlagOverridesEnv tests that --token flag overrides env var.
func TestResolveToken_FlagOverridesEnv(t *testing.T) {
	t.Setenv("SHOEHORN_TOKEN", "shp_from_env")
	token, source := resolveToken("shp_from_flag")
	if token != "shp_from_flag" {
		t.Errorf("resolveToken() token = %q, want %q", token, "shp_from_flag")
	}
	if source != "flag" {
		t.Errorf("resolveToken() source = %q, want %q", source, "flag")
	}
}

// TestResolveToken_Empty tests that empty flag + no env returns empty.
func TestResolveToken_Empty(t *testing.T) {
	os.Unsetenv("SHOEHORN_TOKEN")
	token, _ := resolveToken("")
	if token != "" {
		t.Errorf("resolveToken() token = %q, want empty", token)
	}
}

// ─── Security: HTTP safety validation ────────────────────────────────────────

// TestValidateServerSecurity tests HTTP vs HTTPS validation for remote servers.
func TestValidateServerSecurity_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https remote is safe", "https://api.company.com", false},
		{"http localhost allowed", "http://localhost:8080", false},
		{"http 127.0.0.1 allowed", "http://127.0.0.1:8080", false},
		{"http [::1] allowed", "http://[::1]:8080", false},
		{"http remote blocked", "http://api.company.com", true},
		{"http remote with port blocked", "http://api.company.com:8080", true},
		{"empty string no error", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServerSecurity(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateServerSecurity(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

// TestRunLoginWithPAT_ErrorReturnsNonNil is a regression test for the error
// swallowing bug where runLoginWithPAT returned nil on API failure, causing
// exit code 0 instead of non-zero.
//
// This test verifies that when the API call fails, the function returns a
// non-nil error so Cobra sets exit code 1.
func TestRunLoginWithPAT_ErrorReturnsNonNil(t *testing.T) {
	// runLoginWithPAT calls api.NewClient then GetMe via spinner.
	// We can test the error propagation by calling with an unreachable server.
	// The function should return a non-nil error, not nil.
	err := runLoginWithPAT("http://127.0.0.1:1", "fake-token")
	if err == nil {
		t.Error("runLoginWithPAT with unreachable server returned nil error; want non-nil for correct exit code")
	}
}
