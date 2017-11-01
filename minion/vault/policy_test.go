package vault

import (
	"fmt"
	"testing"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/vault/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test that we don't unecessarily add or delete policies when no change is
// needed.
func TestJoinPoliciesNoChange(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	name1 := "name1"
	policy1 := "policy1"
	name2 := "name2"
	policy2 := "policy2"

	desiredPolicies := []vaultPolicy{
		{name1, policy1},
		{name2, policy2},
	}

	currentPolicies := []vaultPolicy{
		{name2, policy2},
		{name1, policy1},
	}

	joinPolicies(mockClient, desiredPolicies, currentPolicies)

	mockClient.AssertNotCalled(t, "PutPolicy")
	mockClient.AssertNotCalled(t, "DeletePolicy")
}

// Test that we properly add and remove policies, while leaving correct ones
// untouched.
func TestJoinPoliciesPutAndDelete(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	noChangeName := "noChangeName"
	noChangePolicy := "noChangePolicy"

	toAddName := "toAddName"
	toAddPolicy := "toAddPolicy"

	toDelName := "toDelName"
	toDelPolicy := "toDelPolicy"

	desiredPolicies := []vaultPolicy{
		{noChangeName, noChangePolicy},
		{toAddName, toAddPolicy},
	}

	currentPolicies := []vaultPolicy{
		{noChangeName, noChangePolicy},
		{toDelName, toDelPolicy},
	}

	mockClient.On("PutPolicy", toAddName, toAddPolicy).Return(nil, nil)
	mockClient.On("DeletePolicy", toDelName).Return(nil, nil)
	joinPolicies(mockClient, desiredPolicies, currentPolicies)
	mockClient.AssertExpectations(t)
}

// Test that if the desired policy has the same name, but different rules, we
// replace it.
func TestJoinPoliciesReplace(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	toReplaceName := "toReplace"
	oldPolicy := "old"
	newPolicy := "new"

	desiredPolicies := []vaultPolicy{
		{toReplaceName, newPolicy},
	}

	currentPolicies := []vaultPolicy{
		{toReplaceName, oldPolicy},
	}

	mockClient.On("PutPolicy", toReplaceName, newPolicy).Return(nil, nil)
	mockClient.On("DeletePolicy", toReplaceName).Return(nil, nil)
	joinPolicies(mockClient, desiredPolicies, currentPolicies)
	mockClient.AssertExpectations(t)
}

func TestGetDesiredPolicies(t *testing.T) {
	t.Parallel()
	conn := db.New()

	masterIP := "master"
	scheduledMinion := "minion"
	secretName := "aSecret"
	conn.Txn(db.ContainerTable, db.MinionTable).Run(func(view db.Database) error {
		dbc := view.InsertContainer()
		dbc.Minion = scheduledMinion
		dbc.Env = map[string]blueprint.ContainerValue{
			"aKey": blueprint.NewSecret(secretName),
		}
		view.Commit(dbc)

		aMaster := view.InsertMinion()
		aMaster.Role = db.Master
		aMaster.PrivateIP = masterIP
		view.Commit(aMaster)

		return nil
	})
	desiredPolicies := getDesiredPolicies(conn)
	assert.Len(t, desiredPolicies, 2)
	assert.Contains(t, desiredPolicies, vaultPolicy{
		name: scheduledMinion,
		policy: fmt.Sprintf(
			`{"path":{"/secret/kelda/%s":{"policy":"read"}}}`,
			secretName),
	})
	assert.Contains(t, desiredPolicies, vaultPolicy{
		name:   masterIP,
		policy: `{"path":{"/secret/kelda/*":{"policy":"write"}}}`,
	})
}

func TestGetCurrentPoliciesError(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	mockClient.On("ListPolicies").Return(nil, assert.AnError).Once()
	_, err := getCurrentPolicies(mockClient)
	assert.NotNil(t, err)

	mockClient.On("ListPolicies").Return([]string{"aPolicy"}, nil).Once()
	mockClient.On("GetPolicy", mock.Anything).Return("", assert.AnError).Once()
	_, err = getCurrentPolicies(mockClient)
	assert.NotNil(t, err)
}

func TestGetCurrentPoliciesSuccess(t *testing.T) {
	t.Parallel()
	mockClient := &mocks.APIClient{}

	// Test the success case, and that the default policies are ignored.
	name1 := "name1"
	name2 := "name2"
	policy1 := "policy1"
	policy2 := "policy2"
	mockClient.On("ListPolicies").Return(
		[]string{name1, name2, "default", "root"}, nil).Once()
	mockClient.On("GetPolicy", name1).Return(policy1, nil).Once()
	mockClient.On("GetPolicy", name2).Return(policy2, nil).Once()
	policies, err := getCurrentPolicies(mockClient)
	assert.NoError(t, err)
	assert.Equal(t, []vaultPolicy{
		{
			name:   name1,
			policy: policy1,
		},
		{
			name:   name2,
			policy: policy2,
		},
	}, policies)
}
