package common

import (
	"os"
	"testing"
)

func TestLoadEnvs_JwtSecret(t *testing.T) {
	// Case 1: JWT_SECRET is set
	os.Setenv("JWT_SECRET", "my-secret-key")
	LoadEnvs()
	if JwtSecret != "my-secret-key" {
		t.Errorf("Expected JwtSecret to be 'my-secret-key', got '%s'", JwtSecret)
	}

	// Case 2: JWT_SECRET is not set
	os.Unsetenv("JWT_SECRET")
	JwtSecret = "" // Reset to ensure LoadEnvs regenerates it
	LoadEnvs()
	if JwtSecret == "" {
		t.Error("Expected JwtSecret to be generated, got empty string")
	}
	if len(JwtSecret) < 32 {
		t.Errorf("Expected generated JwtSecret to be sufficiently long, got length %d", len(JwtSecret))
	}

	// Store the generated secret to compare later
	generatedSecret := JwtSecret

	// Case 3: JWT_SECRET is not set, run again -> should generate a NEW secret
	LoadEnvs()
	if JwtSecret == "" {
		t.Error("Expected JwtSecret to be generated, got empty string")
	}
	if JwtSecret == generatedSecret {
		// Note: Theoretically possible collision but extremely unlikely
		t.Error("Expected a new random secret to be generated, got the same one")
	}
}
