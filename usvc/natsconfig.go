package usvc

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// NatsURL returns the NATS server URL(s) from NATS_URL, or the default URL
func NatsURL() string {
	if u := strings.TrimSpace(os.Getenv("NATS_URL")); u != "" {
		return u
	}
	return nats.DefaultURL
}

// NatsOptions builds the NATS connection options from the environment:
//
//	NATS_CREDS          path to a .creds file (JWT/NKey authentication)
//	NATS_NKEY_SEED      path to an NKey seed file
//	NATS_USER/NATS_PASSWORD  user name and password
//	NATS_TOKEN          authentication token
//	NATS_CA             CA certificate used to verify the server (enables TLS)
//	NATS_CERT/NATS_KEY  client certificate and key (mutual TLS)
func NatsOptions(name string) ([]nats.Option, error) {
	opts := []nats.Option{
		nats.Name(name),
		nats.MaxReconnects(-1), // Allow infinitely many reconnects
		nats.ReconnectWait(2 * time.Second),
		nats.RetryOnFailedConnect(true), // Start even if the NATS server is not (yet) available
	}

	if creds := os.Getenv("NATS_CREDS"); creds != "" {
		opts = append(opts, nats.UserCredentials(creds))
	}

	if seed := os.Getenv("NATS_NKEY_SEED"); seed != "" {
		opt, err := nats.NkeyOptionFromSeed(seed)
		if err != nil {
			return nil, fmt.Errorf("failed to load NATS NKey seed: %w", err)
		}
		opts = append(opts, opt)
	}

	if user := os.Getenv("NATS_USER"); user != "" {
		opts = append(opts, nats.UserInfo(user, os.Getenv("NATS_PASSWORD")))
	}

	if token := os.Getenv("NATS_TOKEN"); token != "" {
		opts = append(opts, nats.Token(token))
	}

	if ca := os.Getenv("NATS_CA"); ca != "" {
		opts = append(opts, nats.RootCAs(ca))
	}

	cert, key := os.Getenv("NATS_CERT"), os.Getenv("NATS_KEY")
	if cert != "" || key != "" {
		if cert == "" || key == "" {
			return nil, fmt.Errorf("both NATS_CERT and NATS_KEY must be set for mutual TLS")
		}
		opts = append(opts, nats.ClientCert(cert, key))
	}

	return opts, nil
}
