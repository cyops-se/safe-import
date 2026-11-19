package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestCA(t *testing.T, dir string, isCA bool) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  isCA,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	certFile := filepath.Join(dir, "ca.pem")
	keyFile := filepath.Join(dir, "ca.key")
	os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0644)
	os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}), 0600)
	return certFile, keyFile
}

func TestLoadCA(t *testing.T) {
	certFile, keyFile := writeTestCA(t, t.TempDir(), true)
	cert, key, err := loadCA(certFile, keyFile, "")
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != "test CA" || key == nil {
		t.Errorf("unexpected CA %v / %v", cert.Subject, key)
	}
}

func TestLoadCAErrors(t *testing.T) {
	dir := t.TempDir()

	if _, _, err := loadCA(filepath.Join(dir, "missing.pem"), filepath.Join(dir, "missing.key"), ""); err == nil {
		t.Error("expected error for missing files")
	}

	certFile, keyFile := writeTestCA(t, dir, false)
	if _, _, err := loadCA(certFile, keyFile, ""); err == nil {
		t.Error("expected error for a certificate that is not a CA")
	}

	garbage := filepath.Join(dir, "garbage.key")
	os.WriteFile(garbage, []byte("not a key"), 0600)
	caCert, _ := writeTestCA(t, t.TempDir(), true)
	if _, _, err := loadCA(caCert, garbage, ""); err == nil {
		t.Error("expected error for an invalid key")
	}
}

func TestGetCertificateRejectsIPAndEmptyNames(t *testing.T) {
	for _, name := range []string{"", "localhost", "10.0.0.1", "::1"} {
		if _, err := GetCertificateFunc(&tls.ClientHelloInfo{ServerName: name}); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
}
