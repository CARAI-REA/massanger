package prodguard

import (
	"fmt"
	"strings"
)

// Known weak secrets used in local/dev templates — forbidden when APP_ENV=production.
var WeakSecrets = []string{
	"rooms-auth-secret",
	"rooms-join-token-secret",
	"sfu-turn-secret",
	"CHANGE_ME",
	"changeme",
	"secret",
	"password",
	"postgres",
}

func IsProduction(appEnv string) bool {
	return strings.EqualFold(strings.TrimSpace(appEnv), "production")
}

func IsWeakSecret(value string) bool {
	v := strings.TrimSpace(value)
	if v == "" {
		return true
	}
	lower := strings.ToLower(v)
	for _, weak := range WeakSecrets {
		if lower == strings.ToLower(weak) {
			return true
		}
	}
	return false
}

func ForbidWeakSecret(name, value string) error {
	if IsWeakSecret(value) {
		return fmt.Errorf("production: %s must be a strong secret (got weak/empty value)", name)
	}
	return nil
}

func ForbidLocalhostURL(name, value string) error {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1") {
		return fmt.Errorf("production: %s must not point to localhost (got %q)", name, value)
	}
	return nil
}

func RequireNonEmpty(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("production: %s is required", name)
	}
	return nil
}

func ForbidDisabledSSL(name, mode string) error {
	if strings.EqualFold(strings.TrimSpace(mode), "disable") {
		return fmt.Errorf("production: %s must not be disable", name)
	}
	return nil
}
