package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"golang.org/x/crypto/ssh"
)

type SSHKeyTemplate struct {
	Size int `json:"size"`

	// Type of SSH key generated
	// +optional
	Type SSHKeyType `json:"type,omitempty"`
}

type SSHKeyType string

const (
	RSAKey SSHKeyType = "rsa"
)

// GenerateRSAKeyPair makes a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func GenerateRSAKeyPair(spec SSHKeyTemplate) (string, string, error) {
	keySize := spec.Size
	if keySize <= 0 { // 4096
		return "", "", errors.New("invalid key size")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return "", "", err
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	var private bytes.Buffer
	if err := pem.Encode(&private, privateKeyPEM); err != nil {
		return "", "", err
	}

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	public := ssh.MarshalAuthorizedKey(pub)
	return string(public), private.String(), nil
}
