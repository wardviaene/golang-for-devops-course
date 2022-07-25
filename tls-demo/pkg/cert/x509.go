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
	"time"

	"github.com/wardviaene/golang-for-devops-course/tls-demo/pkg/key"
)

func CreateCACert(ca *CACert, keyFile, caCertFile string) error {
	if ca == nil {
		return fmt.Errorf("No CA Config found")
	}
	template := x509.Certificate{
		SerialNumber: ca.Serial,
		Subject: pkix.Name{
			Organization:       checkEmptyString([]string{ca.Subject.Organization}),
			OrganizationalUnit: checkEmptyString([]string{ca.Subject.OrganizationalUnit}),
			Country:            checkEmptyString([]string{ca.Subject.Country}),
			Province:           checkEmptyString([]string{ca.Subject.Province}),
			Locality:           checkEmptyString([]string{ca.Subject.Locality}),
			StreetAddress:      checkEmptyString([]string{ca.Subject.StreetAddress}),
			PostalCode:         checkEmptyString([]string{ca.Subject.PostalCode}),
			CommonName:         ca.Subject.CommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(ca.ValidForYears, 0, 0),
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

func CreateCert(cert *Cert, caKey []byte, caCert []byte, keyFile, certFile string) error {
	template := x509.Certificate{
		SerialNumber: cert.Serial,
		Subject: pkix.Name{
			Organization:       checkEmptyString([]string{cert.Subject.Organization}),
			OrganizationalUnit: checkEmptyString([]string{cert.Subject.OrganizationalUnit}),
			Country:            checkEmptyString([]string{cert.Subject.Country}),
			Province:           checkEmptyString([]string{cert.Subject.Province}),
			Locality:           checkEmptyString([]string{cert.Subject.Locality}),
			StreetAddress:      checkEmptyString([]string{cert.Subject.StreetAddress}),
			PostalCode:         checkEmptyString([]string{cert.Subject.PostalCode}),
			CommonName:         cert.Subject.CommonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(cert.ValidForYears, 0, 0),

		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	caKeyParsed, err := key.PrivateKeyPemToRSA(caKey)
	if err != nil {
		return err
	}
	caCertParsed, err := PemToX509(caCert)
	if err != nil {
		return err
	}

	certBytes, key, err := createCert(template, caKeyParsed, caCertParsed)

	if err := ioutil.WriteFile(keyFile, key.Bytes(), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(certFile, certBytes.Bytes(), 0644); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

func createCert(template x509.Certificate, caKey *rsa.PrivateKey, caCert *x509.Certificate) (bytes.Buffer, bytes.Buffer, error) {
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

func PemToX509(input []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(input)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

func checkEmptyString(input []string) []string {
	if len(input) == 1 && input[0] == "" {
		return []string{}
	}
	return input
}
