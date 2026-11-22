package usvc

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestNatsURL(t *testing.T) {
	t.Setenv("NATS_URL", "")
	if got := NatsURL(); got != nats.DefaultURL {
		t.Errorf("default URL: got %q", got)
	}
	t.Setenv("NATS_URL", "tls://nats.example.com:4222")
	if got := NatsURL(); got != "tls://nats.example.com:4222" {
		t.Errorf("NATS_URL: got %q", got)
	}
}

func applyOptions(t *testing.T, opts []nats.Option) nats.Options {
	t.Helper()
	o := nats.GetDefaultOptions()
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			t.Fatal(err)
		}
	}
	return o
}

func TestNatsOptionsUserPassword(t *testing.T) {
	t.Setenv("NATS_USER", "si-engine")
	t.Setenv("NATS_PASSWORD", "secret")
	opts, err := NatsOptions("test")
	if err != nil {
		t.Fatal(err)
	}
	o := applyOptions(t, opts)
	if o.User != "si-engine" || o.Password != "secret" {
		t.Errorf("user/password not applied: %q/%q", o.User, o.Password)
	}
	if o.Name != "test" || o.MaxReconnect != -1 {
		t.Errorf("unexpected defaults: name %q, max reconnect %d", o.Name, o.MaxReconnect)
	}
}

func TestNatsOptionsToken(t *testing.T) {
	t.Setenv("NATS_TOKEN", "tok")
	opts, err := NatsOptions("test")
	if err != nil {
		t.Fatal(err)
	}
	if o := applyOptions(t, opts); o.Token != "tok" {
		t.Errorf("token not applied: %q", o.Token)
	}
}

func TestNatsOptionsRequiresCertAndKey(t *testing.T) {
	t.Setenv("NATS_CERT", "client.pem")
	t.Setenv("NATS_KEY", "")
	if _, err := NatsOptions("test"); err == nil {
		t.Error("expected error when only NATS_CERT is set")
	}
}
