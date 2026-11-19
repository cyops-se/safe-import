package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds the runtime configuration of si-engine. All values can be
// overridden with environment variables, see docs/CONFIGURATION.md.
type Config struct {
	ListenAddr     string        // SI_ENGINE_LISTEN
	DBPath         string        // SI_ENGINE_DB
	JWTKey         []byte        // SI_ENGINE_JWT_KEY or SI_ENGINE_JWT_KEY_FILE
	TokenTimeout   time.Duration // SI_ENGINE_TOKEN_TIMEOUT
	AllowedOrigins []string      // SI_ENGINE_ALLOWED_ORIGINS (comma separated, empty = same origin only)
}

func getenv(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}

// LoadConfig reads the configuration from the environment
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ListenAddr: getenv("SI_ENGINE_LISTEN", ":7499"),
		DBPath:     getenv("SI_ENGINE_DB", "test.db"),
	}

	timeout, err := time.ParseDuration(getenv("SI_ENGINE_TOKEN_TIMEOUT", "8h"))
	if err != nil {
		return nil, fmt.Errorf("invalid SI_ENGINE_TOKEN_TIMEOUT: %w", err)
	}
	cfg.TokenTimeout = timeout

	if origins := getenv("SI_ENGINE_ALLOWED_ORIGINS", ""); origins != "" {
		for _, o := range strings.Split(origins, ",") {
			if o = strings.TrimSpace(o); o != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
			}
		}
	}

	key, err := loadJWTKey()
	if err != nil {
		return nil, err
	}
	cfg.JWTKey = key

	return cfg, nil
}

// loadJWTKey returns the JWT signing key. Order of precedence:
//  1. SI_ENGINE_JWT_KEY (the key itself, at least 32 characters)
//  2. the file pointed out by SI_ENGINE_JWT_KEY_FILE (default "jwt.key"), created
//     with a random key and mode 0600 if it does not exist
func loadJWTKey() ([]byte, error) {
	if k := os.Getenv("SI_ENGINE_JWT_KEY"); k != "" {
		if len(k) < 32 {
			return nil, fmt.Errorf("SI_ENGINE_JWT_KEY must be at least 32 characters")
		}
		return []byte(k), nil
	}

	filename := getenv("SI_ENGINE_JWT_KEY_FILE", "jwt.key")
	if data, err := os.ReadFile(filename); err == nil {
		key := strings.TrimSpace(string(data))
		if len(key) < 32 {
			return nil, fmt.Errorf("JWT key in %s is too short (min 32 characters)", filename)
		}
		return []byte(key), nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read JWT key file %s: %w", filename, err)
	}

	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to generate JWT key: %w", err)
	}
	key := hex.EncodeToString(buf)

	if dir := filepath.Dir(filename); dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create directory for JWT key: %w", err)
		}
	}
	if err := os.WriteFile(filename, []byte(key), 0600); err != nil {
		return nil, fmt.Errorf("failed to write JWT key file %s: %w", filename, err)
	}
	log.Printf("Generated new JWT signing key in %s", filename)

	return []byte(key), nil
}
