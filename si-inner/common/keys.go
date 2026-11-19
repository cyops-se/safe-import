package common

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// The CA used to sign the certificates presented to inner clients. The CA certificate must be
// distributed to (and trusted by) the inner clients. It is generated at installation, see
// install/generate-ca.sh, and must NEVER be committed to version control.
//
//	SI_INNER_CA_CERT        path to the CA certificate (default myCA.pem)
//	SI_INNER_CA_KEY         path to the CA private key (default myCA.key)
//	SI_INNER_CA_PASSPHRASE  passphrase of a legacy encrypted PEM key (optional)
var (
	caOnce sync.Once
	ca     *x509.Certificate
	caPriv crypto.Signer
	caErr  error
)

// IssSerial is a struct that holds the issuer and serial for a certificate for checking if it already exists so you don't make duplicates
type IssSerial struct {
	issuer string
	serial string
}

var (
	// Map containing all already created certificates
	certsMu sync.Mutex
	certs   = make(map[IssSerial]tls.Certificate)
)

func getenv(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}

// LoadCA loads the CA certificate and key. It is called once; later calls return the same result.
func LoadCA() (*x509.Certificate, crypto.Signer, error) {
	caOnce.Do(func() {
		ca, caPriv, caErr = loadCA(getenv("SI_INNER_CA_CERT", "myCA.pem"), getenv("SI_INNER_CA_KEY", "myCA.key"), os.Getenv("SI_INNER_CA_PASSPHRASE"))
		if caErr == nil {
			log.Println("Loaded CA", ca.Subject.String())
		}
	})
	return ca, caPriv, caErr
}

func loadCA(certFile, keyFile, passphrase string) (*x509.Certificate, crypto.Signer, error) {
	certBytes, err := os.ReadFile(certFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA certificate (generate one with install/generate-ca.sh): %w", err)
	}
	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("no PEM data found in %s", certFile)
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}
	if !cert.IsCA {
		return nil, nil, fmt.Errorf("%s is not a CA certificate", certFile)
	}

	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA key: %w", err)
	}
	if info, err := os.Stat(keyFile); err == nil && info.Mode().Perm()&0077 != 0 {
		log.Printf("WARNING: CA key %s is readable by group/others (mode %v), use chmod 600", keyFile, info.Mode().Perm())
	}

	key, err := parsePrivateKey(keyBytes, passphrase)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA key %s: %w", keyFile, err)
	}

	return cert, key, nil
}

func parsePrivateKey(data []byte, passphrase string) (crypto.Signer, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM data found")
	}

	der := block.Bytes
	//lint:ignore SA1019 legacy encrypted PEM keys are still supported for existing installations
	if x509.IsEncryptedPEMBlock(block) { //nolint:staticcheck
		if passphrase == "" {
			return nil, errors.New("key is encrypted, set SI_INNER_CA_PASSPHRASE")
		}
		var err error
		//lint:ignore SA1019 see above
		if der, err = x509.DecryptPEMBlock(block, []byte(passphrase)); err != nil { //nolint:staticcheck
			return nil, err
		}
	}

	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, errors.New("unsupported private key format")
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New("unsupported private key type")
	}
	return signer, nil
}

// GetCertificateFunc copies the certificate from the given server and returns one signed by the CA
func GetCertificateFunc(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if isIP(helloInfo.ServerName) || helloInfo.ServerName == "localhost" || helloInfo.ServerName == "" {
		return nil, errors.New("Tried to connect to IP directly")
	}

	caCert, caKey, err := LoadCA()
	if err != nil {
		return nil, err
	}

	// Dial the given server to get its certificate
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", helloInfo.ServerName+":443", &tls.Config{ServerName: helloInfo.ServerName})
	if err != nil {
		return nil, err
	}
	state := conn.ConnectionState()
	conn.Close()

	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no certificate received from %s", helloInfo.ServerName)
	}

	c := state.PeerCertificates[0]
	iss := IssSerial{c.Issuer.String(), c.SerialNumber.String()}

	certsMu.Lock()
	gotCert, ok := certs[iss]
	certsMu.Unlock()
	if ok {
		return &gotCert, nil
	}

	// Generate a new certificate with data from the server
	template := &x509.Certificate{
		SerialNumber: c.SerialNumber,
		Subject:      c.Subject,
		NotBefore:    c.NotBefore,
		NotAfter:     c.NotAfter,
		SubjectKeyId: c.SubjectKeyId,
		ExtKeyUsage:  c.ExtKeyUsage,
		KeyUsage:     c.KeyUsage,
		DNSNames:     c.DNSNames,
		URIs:         c.URIs,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	newCert, err := x509.CreateCertificate(rand.Reader, template, caCert, &certPrivKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: newCert})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey)})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	certsMu.Lock()
	certs[iss] = tlsCert
	certsMu.Unlock()

	return &tlsCert, nil
}

// Check if argument is IP address
func isIP(x string) bool {
	return net.ParseIP(x) != nil
}
