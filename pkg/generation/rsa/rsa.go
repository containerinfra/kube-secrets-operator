package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

type SSHKeyTemplate struct {
	Size int `json:"size"`

	// Type of SSH key generated
	// +optional
	Type string `json:"type,omitempty"`

	// FileName is an optional attribute to specify the filename of the private and public key generated.
	// Public key will be FileName + .pub
	// +optional
	FileName string `json:"fileName"`
}

// GenerateRSAKeyPair makes a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func GenerateRSAKeyPair(spec *SSHKeyTemplate) (string, string, error) {
	keySize := spec.Size
	if keySize <= 0 {
		keySize = 4096
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
