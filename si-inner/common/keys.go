package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"net"
)

var (
	// Load the CA
	ca, caPriv = GetCa()
)

// IssSerial is a struct that holds the issuer and serial for a certificate for checking if it already exists so you don't make duplicates
type IssSerial struct {
	issuer string
	serial string
}

var (
	// Map containing all already created certificates
	certs map[IssSerial]tls.Certificate = make(map[IssSerial]tls.Certificate)
)

// GetCa loads the Certificate Authority certificate and private key
func GetCa() (*x509.Certificate, *rsa.PrivateKey) {
	// Load certificate PEMBlock
	certKeyBytes, _ := ioutil.ReadFile("myCA.pem")
	certPEMBlock, _ := pem.Decode(certKeyBytes)

	// Load and doced privateKey PEMBlock
	privateKeyBytes, _ := ioutil.ReadFile("myCA.key")
	privateKeyPEMBlock, _ := pem.Decode(privateKeyBytes)
	privateKeyDecodedBytes, err := x509.DecryptPEMBlock(privateKeyPEMBlock, []byte("aaaa"))
	if err != nil {
		log.Println(err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certPEMBlock.Bytes)
	if err != nil {
		log.Println(err)
	}

	// Parse the private key
	keyA, err := x509.ParsePKCS1PrivateKey(privateKeyDecodedBytes)
	if err != nil {
		log.Println(err)
	}

	log.Println("Loaded CA")

	return cert, keyA
}

// GetCertificateFunc copies the certificate from the given server and returns one signed by the CA
func GetCertificateFunc(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if isIP(helloInfo.ServerName) || helloInfo.ServerName == "localhost" {
		return nil, errors.New("Tried to connect to IP directly")
	}

	// Dial the given server
	conn, err := tls.Dial("tcp", helloInfo.ServerName+":443", &tls.Config{})

	// Check if there is a certificate
	if err == nil && len(conn.ConnectionState().PeerCertificates) > 0 {
		// Get the certificate
		c := conn.ConnectionState().PeerCertificates[0]
		iss := IssSerial{
			c.Issuer.String(),
			c.SerialNumber.String(),
		}

		// Check if certificate for issuer and serial number already exists
		gotCert, ok := certs[iss]
		if !ok {
			conn.Close()

			// Generate a new certificate with data from the servers
			cert := &x509.Certificate{
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

			// Generate a key
			certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
			if err != nil {
				log.Println(err)
				return &tls.Certificate{}, err
			}

			// Create the certificate
			newCert, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPriv)
			if err != nil {
				log.Println(err)
			}

			// Encode it
			certPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: newCert,
			})

			keyPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
			})

			// Load it into a tls.Certificate
			certs[iss], err = tls.X509KeyPair(certPEM, keyPEM)
			c := certs[iss]

			return &c, err
		}

		return &gotCert, nil
	}

	return nil, err
}

// Check if argument is IP address
func isIP(x string) bool {
	return net.ParseIP(x) != nil
}
