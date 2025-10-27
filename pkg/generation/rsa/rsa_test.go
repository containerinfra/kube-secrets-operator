package rsa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	spec := SSHKeyTemplate{
		Size: 2048,
		Type: RSAKey,
	}

	publicKey, privateKey, err := GenerateRSAKeyPair(spec)
	require.NoError(t, err)
	assert.NotEmpty(t, publicKey, "public key should not be empty")
	assert.NotEmpty(t, privateKey, "private key should not be empty")

	// validate the ssh key
	_, _, _, _, err = ssh.ParseAuthorizedKey([]byte(publicKey))
	require.NoError(t, err, "failed to parse public key")

	_, err = ssh.ParsePrivateKey([]byte(privateKey))
	require.NoError(t, err, "failed to parse private key")

	// test with invalid key size
	spec.Size = 0
	_, _, err = GenerateRSAKeyPair(spec)
	require.Error(t, err, "expected error for invalid key size")
}
