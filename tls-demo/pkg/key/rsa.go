package key

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func CreateRSAPrivateKey(n int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, n)
}

func RSAPrivateKeyToPEM(privateKey *rsa.PrivateKey) *pem.Block {
	return &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
}

func CreateRSAPrivateKeyAndSave(path string, n int) error {
	privateKey, err := CreateRSAPrivateKey(n)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if err := pem.Encode(f, RSAPrivateKeyToPEM(privateKey)); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
