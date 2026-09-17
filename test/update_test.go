package gmk_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/msdevbytes/gomicrokit/cmd"
)

func TestParseModuleLatest(t *testing.T) {
	got, err := ParseModuleLatest([]byte(`{"Version":"v1.3.0","Time":"2026-02-02T00:00:00Z"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got != "v1.3.0" {
		t.Fatalf("got %q", got)
	}
	if _, err := ParseModuleLatest([]byte(`{}`)); err == nil {
		t.Fatal("expected error for empty Version")
	}
	if _, err := ParseModuleLatest([]byte(`not-json`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSameKitVersion(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
	}{
		{current: "1.3.0", latest: "v1.3.0", want: true},
		{current: "v1.3.0", latest: "v1.3.0", want: true},
		{current: "v1.4.1", latest: "v1.4.0", want: false},
		{current: "dev", latest: "v1.3.0", want: false},
		{current: "unknown", latest: "v1.3.0", want: false},
		{current: "", latest: "v1.3.0", want: false},
	}
	for _, tt := range tests {
		if got := SameKitVersion(tt.current, tt.latest); got != tt.want {
			t.Fatalf("SameKitVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestCompareKitVersion(t *testing.T) {
	tests := []struct {
		current, latest string
		want            int
	}{
		{current: "v1.4.1", latest: "v1.4.0", want: 1},
		{current: "v1.4.0", latest: "v1.4.1", want: -1},
		{current: "1.4.1", latest: "v1.4.1", want: 0},
		{current: "dev", latest: "v1.4.1", want: -1},
	}
	for _, tt := range tests {
		if got := CompareKitVersion(tt.current, tt.latest); got != tt.want {
			t.Fatalf("CompareKitVersion(%q, %q) = %d, want %d", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestRunUpdateSkipsDowngrade(t *testing.T) {
	origVersion := Version
	origFetch := FetchLatest
	origLook := GoLookPath
	origInstall := RunGoInstall
	t.Cleanup(func() {
		Version = origVersion
		FetchLatest = origFetch
		GoLookPath = origLook
		RunGoInstall = origInstall
	})
	Version = "v1.4.1"
	FetchLatest = func() (string, error) { return "v1.4.0", nil }
	GoLookPath = func(file string) (string, error) { return "/usr/bin/" + file, nil }
	var installed string
	RunGoInstall = func(mod string) error {
		installed = mod
		return nil
	}

	if err := RunUpdate(false, ""); err != nil {
		t.Fatal(err)
	}
	if installed != "" {
		t.Fatalf("would have downgraded to %q", installed)
	}
}

func TestRunUpdateCheckAlreadyCurrent(t *testing.T) {
	origVersion := Version
	origFetch := FetchLatest
	t.Cleanup(func() {
		Version = origVersion
		FetchLatest = origFetch
	})
	Version = "v1.3.0"
	FetchLatest = func() (string, error) { return "v1.3.0", nil }

	if err := RunUpdate(true, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRunUpdateCheckBehind(t *testing.T) {
	origVersion := Version
	origFetch := FetchLatest
	t.Cleanup(func() {
		Version = origVersion
		FetchLatest = origFetch
	})
	Version = "v1.2.0"
	FetchLatest = func() (string, error) { return "v1.3.0", nil }

	if err := RunUpdate(true, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRunUpdateInstallsLatest(t *testing.T) {
	origVersion := Version
	origFetch := FetchLatest
	origLook := GoLookPath
	origInstall := RunGoInstall
	t.Cleanup(func() {
		Version = origVersion
		FetchLatest = origFetch
		GoLookPath = origLook
		RunGoInstall = origInstall
	})
	Version = "dev"
	FetchLatest = func() (string, error) { return "v1.3.0", nil }
	GoLookPath = func(file string) (string, error) { return "/usr/bin/" + file, nil }
	var installed string
	RunGoInstall = func(mod string) error {
		installed = mod
		return nil
	}

	if err := RunUpdate(false, ""); err != nil {
		t.Fatal(err)
	}
	if installed != GMKInstallModule+"@v1.3.0" {
		t.Fatalf("installed %q", installed)
	}
}

func TestRunUpdateMissingGo(t *testing.T) {
	origLook := GoLookPath
	origFetch := FetchLatest
	origVersion := Version
	t.Cleanup(func() {
		GoLookPath = origLook
		FetchLatest = origFetch
		Version = origVersion
	})
	Version = "dev"
	FetchLatest = func() (string, error) { return "v1.3.0", nil }
	GoLookPath = func(string) (string, error) { return "", errors.New("not found") }

	err := RunUpdate(false, "")
	if err == nil || !strings.Contains(err.Error(), "Go is not on PATH") {
		t.Fatalf("error = %v", err)
	}
}

func TestFetchLatestFromProxyHTTP(t *testing.T) {
	origClient := UpdateHTTPClient
	t.Cleanup(func() { UpdateHTTPClient = origClient })

	UpdateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != GMKProxyLatest {
				t.Fatalf("url = %s", req.URL)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"Version":"v9.9.9"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	got, err := FetchLatestFromProxy()
	if err != nil {
		t.Fatal(err)
	}
	if got != "v9.9.9" {
		t.Fatalf("got %q", got)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
