package gmk_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	. "github.com/msdevbytes/gomicrokit/generator"
)

func TestGenerateAPIKey(t *testing.T) {
	key, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}
	pattern := `^svc_test_[0-9A-Za-z]{43}$`
	if matched, _ := regexp.MatchString(pattern, key); !matched {
		t.Fatalf("GenerateAPIKey() = %q, want %s", key, pattern)
	}
	other, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() second error = %v", err)
	}
	if key == other {
		t.Fatal("GenerateAPIKey() returned the same value twice")
	}
}

func TestGenerateAPIKeyLivePrefix(t *testing.T) {
	key, err := GenerateAPIKeyWithPrefix(APIKeyPrefixLive)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "svc_live_") {
		t.Fatalf("live key = %q", key)
	}
}

func TestAPIKeyPrefixForEnv(t *testing.T) {
	if got := APIKeyPrefixForEnv("dev"); got != APIKeyPrefixTest {
		t.Fatalf("dev prefix = %q", got)
	}
	if got := APIKeyPrefixForEnv("prod"); got != APIKeyPrefixLive {
		t.Fatalf("prod prefix = %q", got)
	}
	if got := APIKeyPrefixForEnv("production"); got != APIKeyPrefixLive {
		t.Fatalf("production prefix = %q", got)
	}
}

func TestGenerateProjectAPIKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-apikey")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	envPath := filepath.Join(tmpDir, ".env")
	initial := "APP_NAME=demo\nAPI_KEY=change-me-in-prod\nENABLE_AUTH=yes\n"
	if err := os.WriteFile(envPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	key, err := GenerateProjectAPIKey(tmpDir, false)
	if err != nil {
		t.Fatalf("GenerateProjectAPIKey() error = %v", err)
	}
	body, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if !strings.Contains(got, "API_KEY="+key) {
		t.Fatalf("env missing generated key, got:\n%s", got)
	}
	if !strings.Contains(got, "ENABLE_AUTH=yes") {
		t.Fatal("env lost unrelated keys")
	}
	if !strings.HasPrefix(key, "svc_test_") {
		t.Fatalf("expected svc_test_ key when APP_ENV is unset, got %q", key)
	}

	if _, err := GenerateProjectAPIKey(tmpDir, false); err == nil {
		t.Fatal("expected error when replacing existing key without --force")
	}

	next, err := GenerateProjectAPIKey(tmpDir, true)
	if err != nil {
		t.Fatalf("GenerateProjectAPIKey(--force) error = %v", err)
	}
	if next == key {
		t.Fatal("force generate returned the same key")
	}
	body, err = os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "API_KEY="+next) {
		t.Fatalf("force generate did not replace key, got:\n%s", body)
	}
}

func TestGenerateProjectAPIKeyMissingEnv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-apikey-missing")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	if _, err := GenerateProjectAPIKey(tmpDir, false); err == nil {
		t.Fatal("expected error when .env is missing")
	}
}

func TestGenerateProjectAPIKeyAppendsMissingLine(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-apikey-append")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("APP_NAME=demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	key, err := GenerateProjectAPIKey(tmpDir, false)
	if err != nil {
		t.Fatalf("GenerateProjectAPIKey() append error = %v", err)
	}
	body, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "API_KEY="+key) {
		t.Fatalf("expected API_KEY to be appended, got:\n%s", body)
	}
}

func TestGenerateProjectAPIKeyLivePrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmk-apikey-live")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte("APP_ENV=prod\nAPI_KEY=\n"), 0644); err != nil {
		t.Fatal(err)
	}
	key, err := GenerateProjectAPIKey(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "svc_live_") {
		t.Fatalf("expected svc_live_ for APP_ENV=prod, got %q", key)
	}
}
