// Package tls provides utilities for generating self-signed TLS certificates.
package tls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

// GetOrGenerateTLS tries to load a TLS certificate and key from the given paths, and if that fails,
// it generates a new self-signed certificate and key and saves it.
func GetOrGenerateTLS(certPath, certKeyPath string, validFor time.Duration, hosts []string) (cert, key []byte, err error) {
	// Check if the certificate and key files exist
	_, err1 := os.Stat(certPath)
	_, err2 := os.Stat(certKeyPath)
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		// If either file does not exist, generate a new certificate and key
		cert, key, err = GenerateTLS(validFor, hosts)
		if err != nil {
			return nil, nil, err
		}
		// Save the generated certificate and key to the specified paths
		err = os.WriteFile(certPath, cert, 0644)
		if err != nil {
			return nil, nil, err
		}
		err = os.WriteFile(certKeyPath, key, 0600)
		if err != nil {
			return nil, nil, err
		}
		return cert, key, nil
	}

	// The files exist, read them
	cert, err = os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}
	key, err = os.ReadFile(certKeyPath)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

// GenerateTLS generated a TLS certificate and key.
// based on https://go.dev/src/crypto/tls/generate_cert.go
// - `hosts`: a list of ip / dns names to include in the certificate
func GenerateTLS(validFor time.Duration, hosts []string) (cert, key []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	keyUsage := x509.KeyUsageDigitalSignature

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// certificate is its own CA
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	var certOut bytes.Buffer
	if err = pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, nil, err
	}
	cert = certOut.Bytes()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	var keyOut bytes.Buffer
	err = pem.Encode(&keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return nil, nil, err
	}
	key = keyOut.Bytes()
	return cert, key, nil
}
