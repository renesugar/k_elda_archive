package vault

import (
	"testing"

	"github.com/kelda/kelda/minion/vault/mocks"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestReadAndWrite(t *testing.T) {
	mockClient := &mocks.APIClient{}
	store := secretStoreImpl{mockClient}

	secretName := "secretName"
	secretValue := "secretValue"

	// Test that we write with the proper parameters.
	mockClient.On("Write", pathForSecret(secretName), map[string]interface{}{
		secretKey: secretValue,
	}).Return(nil, nil)
	err := store.Write(secretName, secretValue)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)

	// Test an error reading from Vault.
	mockClient.On("Read", pathForSecret(secretName)).
		Return(nil, assert.AnError).Once()
	_, err = store.Read(secretName)
	assert.NotNil(t, err)
	mockClient.AssertExpectations(t)

	// Test when a secret is undefined.
	mockClient.On("Read", pathForSecret(secretName)).
		Return(nil, nil).Once()
	_, err = store.Read(secretName)
	assert.Equal(t, ErrSecretDoesNotExist, err)
	mockClient.AssertExpectations(t)

	// Test when the secret was not properly written.
	mockClient.On("Read", pathForSecret(secretName)).Return(
		&vaultAPI.Secret{
			Data: map[string]interface{}{},
		}, nil,
	).Once()
	_, err = store.Read(secretName)
	assert.EqualError(t, err, "malformed secret")
	mockClient.AssertExpectations(t)

	// Test a successful read.
	mockClient.On("Read", pathForSecret(secretName)).Return(
		&vaultAPI.Secret{
			Data: map[string]interface{}{
				secretKey: secretValue,
			},
		}, nil,
	).Once()
	actualValue, err := store.Read(secretName)
	assert.Nil(t, err)
	assert.Equal(t, secretValue, actualValue)
	mockClient.AssertExpectations(t)
}
