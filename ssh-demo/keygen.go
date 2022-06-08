package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

func GenerateKeys() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

	pubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return pem.EncodeToMemory(privateKeyPEM), ssh.MarshalAuthorizedKey(pubKey), nil
}
