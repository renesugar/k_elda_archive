package vault

import (
	"testing"

	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/vault/mocks"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDesiredRoles(t *testing.T) {
	t.Parallel()
	conn := db.New()

	minion1 := "minion1"
	minion1Cert := "minion1Cert"
	minion2 := "minion2"
	minion2Cert := "minion2Cert"
	conn.Txn(db.MinionTable).Run(func(view db.Database) error {
		minion := view.InsertMinion()
		minion.Self = true
		minion.MinionIPToPublicKey = map[string]string{
			minion1: minion1Cert,
			minion2: minion2Cert,
		}
		view.Commit(minion)
		return nil
	})

	roles := getDesiredRoles(conn)
	assert.Len(t, roles, 2)
	assert.Contains(t, roles, certRole{
		name:     minion1,
		policies: []string{minion1},
		cert:     minion1Cert,
	})
	assert.Contains(t, roles, certRole{
		name:     minion2,
		policies: []string{minion2},
		cert:     minion2Cert,
	})
}

func TestGetCurrentRolesError(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	// A failed List.
	mockClient.On("List", certListEndpoint).Return(nil, assert.AnError).Once()
	_, err := getCurrentRoles(mockClient)
	assert.NotNil(t, err)

	// A successful List, but failed Read.
	mockClient.On("List", certListEndpoint).Return(
		&vaultAPI.Secret{
			Data: map[string]interface{}{
				"keys": []interface{}{"aRole"},
			},
		}, nil,
	).Once()
	mockClient.On("Read", mock.Anything).Return(nil, assert.AnError).Once()
	_, err = getCurrentRoles(mockClient)
	assert.NotNil(t, err)
	mockClient.AssertExpectations(t)
}

// Test that we properly return an empty slice when no roles have been created.
func TestGetCurrentRolesNotInit(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	mockClient.On("List", certListEndpoint).Return(nil, nil).Once()
	roles, err := getCurrentRoles(mockClient)
	assert.NoError(t, err)
	assert.Empty(t, roles)
}

func TestGetCurrentRoles(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	roleName := "role"
	policyName := "policy"
	certificate := "cert"

	mockClient.On("List", certListEndpoint).Return(
		&vaultAPI.Secret{
			Data: map[string]interface{}{
				"keys": []interface{}{roleName},
			},
		}, nil,
	).Once()
	mockClient.On("Read", pathForRole(roleName)).Return(
		&vaultAPI.Secret{
			Data: map[string]interface{}{
				"policies":    []interface{}{policyName},
				"certificate": certificate,
			},
		}, nil,
	).Once()
	roles, err := getCurrentRoles(mockClient)
	assert.NoError(t, err)
	assert.Equal(t, []certRole{
		{
			name:     roleName,
			policies: []string{policyName},
			cert:     certificate,
		},
	}, roles)
}

func TestJoinRoles(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	noChangeName := "noChangeName"
	noChangePolicy := "noChangePolicy"
	noChangeCert := "noChangedCert"

	toAddName := "toAddName"
	toAddPolicy := "toAddPolicy"
	toAddCert := "toAddCert"

	toDelName := "toDelName"
	toDelPolicy := "toDelPolicy"
	toDelCert := "toDelCert"

	desiredRoles := []certRole{
		{noChangeName, noChangeCert, []string{noChangePolicy}},
		{toAddName, toAddCert, []string{toAddPolicy}},
	}

	currentRoles := []certRole{
		{noChangeName, noChangeCert, []string{noChangePolicy}},
		{toDelName, toDelCert, []string{toDelPolicy}},
	}

	mockClient.On("Write", pathForRole(toAddName),
		map[string]interface{}{
			"certificate": toAddCert,
			"policies":    toAddPolicy,
		}).Return(nil, nil)
	mockClient.On("Delete", pathForRole(toDelName)).Return(nil, nil)
	joinRoles(mockClient, desiredRoles, currentRoles)
	mockClient.AssertExpectations(t)
}
