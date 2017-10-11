//go:generate mockery -name=SecretStore

package vault

import (
	"errors"
	"path"
)

const (
	// secretStorePath is the root Vault path under which all Kelda secrets are
	// stored.
	secretStorePath = "/secret/kelda"

	// secretKey is the key used to store the secret's value. A key is necessary
	// because a Vault path is a map of key-value pairs -- not just a single
	// value.
	secretKey = "value"
)

var (
	// ErrSecretDoesNotExist is the error returned when a queried secret is not
	// in Vault.
	ErrSecretDoesNotExist = errors.New("secret does not exist")
)

// SecretStore is an interface for minions to read and write key-value pairs
// into Vault. Both the key and value are encrypted at rest and in transit.
type SecretStore interface {
	Read(string) (string, error)
	Write(string, string) error
}

// secretStoreImpl is an implementation of SecretStore that stores each secret
// in a path based on the secret's name in Vault. Each secret is stored in a
// unique path so that policies can be created to restrict access on a
// per-secret basis.
type secretStoreImpl struct {
	vaultClient APIClient
}

// New creates a client connected to Vault that can be used to read and write
// secrets. It uses the minion's TLS certificates to authenticate, which it
// reads from the minion's filesystem.
func New(ip string) (SecretStore, error) {
	vaultClient, err := newVaultAPIClient(ip)
	if err != nil {
		return nil, err
	}

	// Authenticate with the Vault server with the client's TLS credentials.
	// The authentication credentials are the same as those used for transport
	// security.
	loginResp, err := vaultClient.Write(certLoginEndpoint, nil)
	if err != nil {
		return nil, err
	}

	vaultClient.SetToken(loginResp.Auth.ClientToken)
	return secretStoreImpl{vaultClient}, nil
}

func (client secretStoreImpl) Read(name string) (string, error) {
	secretStore, err := client.vaultClient.Read(pathForSecret(name))
	if err != nil {
		return "", err
	}

	if secretStore == nil {
		return "", ErrSecretDoesNotExist
	}

	secret, ok := secretStore.Data[secretKey]
	if !ok {
		return "", errors.New("malformed secret")
	}

	return secret.(string), nil
}

func (client secretStoreImpl) Write(name, value string) error {
	_, err := client.vaultClient.Write(pathForSecret(name),
		map[string]interface{}{secretKey: value})
	return err
}

func pathForSecret(name string) string {
	return path.Join(secretStorePath, name)
}
