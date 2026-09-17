package generator

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	envFileName       = ".env"
	apiKeyEnvName     = "API_KEY"
	placeholderAPIKey = "change-me-in-prod"
	apiKeySecretLen   = 43 // ceil(256 bits / log2(62)) ≈ 43 alphanumeric chars
	apiKeyAlphabet    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	APIKeyPrefixTest  = "svc_test_"
	APIKeyPrefixLive  = "svc_live_"
)

// GenerateAPIKey returns a Stripe-style secret: svc_test_ plus 256 bits of
// alphanumeric material from crypto/rand.
func GenerateAPIKey() (string, error) {
	return GenerateAPIKeyWithPrefix(APIKeyPrefixTest)
}

func GenerateAPIKeyWithPrefix(prefix string) (string, error) {
	secret, err := randomAlphanumeric(apiKeySecretLen)
	if err != nil {
		return "", err
	}
	return prefix + secret, nil
}

func randomAlphanumeric(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("cannot generate API key: invalid length")
	}
	const maxUnbiased = 248 // 62 * 4; bytes at or above this are rejected
	out := make([]byte, n)
	var buf [1]byte
	for i := 0; i < n; {
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf("cannot generate API key: %w", err)
		}
		if buf[0] >= maxUnbiased {
			continue
		}
		out[i] = apiKeyAlphabet[int(buf[0])%len(apiKeyAlphabet)]
		i++
	}
	return string(out), nil
}

func APIKeyPrefixForEnv(appEnv string) string {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "prod", "production", "live":
		return APIKeyPrefixLive
	default:
		return APIKeyPrefixTest
	}
}

func appEnvFromDotEnv(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		name, value, found := strings.Cut(trimmed, "=")
		if !found || strings.TrimSpace(name) != "APP_ENV" {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			quote := value[0]
			if (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
				value = value[1 : len(value)-1]
			}
		}
		return value
	}
	return ""
}

// GenerateProjectAPIKey writes a new API_KEY into the project's .env.
// Existing non-placeholder keys are left unchanged unless force is true.
func GenerateProjectAPIKey(root string, force bool) (string, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	envPath := filepath.Join(root, envFileName)
	data, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s not found in %s; run this from a generated project", envFileName, root)
		}
		return "", fmt.Errorf("cannot read %s: %w", envFileName, err)
	}

	content := string(data)
	nl := "\n"
	if strings.Contains(content, "\r\n") {
		nl = "\r\n"
		content = strings.ReplaceAll(content, "\r\n", "\n")
	}

	key, err := GenerateAPIKeyWithPrefix(APIKeyPrefixForEnv(appEnvFromDotEnv(content)))
	if err != nil {
		return "", err
	}

	lines := strings.Split(content, "\n")
	replaced := false
	for i, line := range lines {
		current, ok := apiKeyAssignment(line)
		if !ok {
			continue
		}
		if !force && !isPlaceholderAPIKey(current) {
			return "", fmt.Errorf("API_KEY already set; use --force to replace it")
		}
		lines[i] = apiKeyEnvName + "=" + key
		replaced = true
		break
	}
	if !replaced {
		assignment := apiKeyEnvName + "=" + key
		switch {
		case len(lines) == 0:
			lines = []string{assignment}
		case lines[len(lines)-1] == "":
			lines[len(lines)-1] = assignment
			lines = append(lines, "")
		default:
			lines = append(lines, assignment)
		}
	}

	output := strings.Join(lines, nl)
	if !strings.HasSuffix(output, nl) {
		output += nl
	}
	if err := os.WriteFile(envPath, []byte(output), 0644); err != nil {
		return "", fmt.Errorf("cannot write %s: %w", envFileName, err)
	}
	return key, nil
}

func apiKeyAssignment(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	name, value, found := strings.Cut(trimmed, "=")
	if !found || strings.TrimSpace(name) != apiKeyEnvName {
		return "", false
	}
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		quote := value[0]
		if (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
			value = value[1 : len(value)-1]
		}
	}
	return value, true
}

func isPlaceholderAPIKey(value string) bool {
	switch strings.TrimSpace(value) {
	case "", placeholderAPIKey:
		return true
	default:
		return false
	}
}
