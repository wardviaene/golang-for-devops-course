package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/key"
)

func CreateClientCert(caCert *x509.Certificate, caKey interface{}, subject string) (bytes.Buffer, bytes.Buffer, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}
	return CreateClientCertWithSerial(caCert, caKey, subject, serialNumber)
}

func CreateCACert(ca *CACert, keyFile, caCertFile string) error {
	if ca == nil {
		return fmt.Errorf("No CA Config found")
	}
	template := x509.Certificate{
		SerialNumber: ca.Serial,
		Subject: pkix.Name{
			Organization:       []string{ca.Subject.Organization},
			OrganizationalUnit: []string{ca.Subject.OrganizationalUnit},
			Country:            []string{ca.Subject.Country},
			Province:           []string{ca.Subject.Province},
			Locality:           []string{ca.Subject.Locality},
			StreetAddress:      []string{ca.Subject.StreetAddress},
			PostalCode:         []string{ca.Subject.PostalCode},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(ca.validForYears, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	cert, key, err := createCert(template, nil, nil)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyFile, key.Bytes(), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(caCertFile, cert.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func CreateClientCertWithSerial(caCert *x509.Certificate, caKey interface{}, subject string, serialNumber *big.Int) (bytes.Buffer, bytes.Buffer, error) {
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{os.Getenv("CLIENT_CERT_ORG")},
			CommonName:   subject,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 1, 0),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	return createCert(template, caKey, caCert)
}

func createCert(template x509.Certificate, caKey interface{}, caCert *x509.Certificate) (bytes.Buffer, bytes.Buffer, error) {
	var (
		certOut  bytes.Buffer
		keyOut   bytes.Buffer
		derBytes []byte
	)
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return certOut, keyOut, err
	}
	if template.IsCA {
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			return certOut, keyOut, err
		}
	} else {
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
		if err != nil {
			return certOut, keyOut, err
		}
	}

	if err != nil {
		return certOut, keyOut, err
	}
	if err := pem.Encode(&certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return certOut, keyOut, err
	}

	if err := pem.Encode(&keyOut, key.RSAPrivateKeyToPEM(privateKey)); err != nil {
		return certOut, keyOut, err
	}

	return certOut, keyOut, nil
}
