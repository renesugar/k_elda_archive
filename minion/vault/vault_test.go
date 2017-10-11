package vault

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	tlsIO "github.com/kelda/kelda/connection/tls/io"
	"github.com/kelda/kelda/minion/docker"
	"github.com/kelda/kelda/minion/vault/mocks"
	"github.com/kelda/kelda/util"

	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartVaultContainer(t *testing.T) {
	util.AppFs = afero.NewMemMapFs()
	md, dk := docker.NewMock()

	caCert := "caCert"
	serverCert := "serverCert"
	serverKey := "serverKey"

	// Test the errors from the TLS credentials not being on the filesystem.
	err := startVaultContainer(dk, "")
	assert.Contains(t, err.Error(), "failed to read Vault CA")

	err = util.WriteFile(tlsIO.CACertPath(tlsIO.MinionTLSDir), []byte(caCert), 0644)
	assert.NoError(t, err)
	err = startVaultContainer(dk, "")
	assert.Contains(t, err.Error(), "failed to read Vault server certificate")

	err = util.WriteFile(tlsIO.SignedCertPath(tlsIO.MinionTLSDir),
		[]byte(serverCert), 0644)
	assert.NoError(t, err)
	err = startVaultContainer(dk, "")
	assert.Contains(t, err.Error(), "failed to read Vault server key")

	// Write the final credential file. We should now boot the Vault container.
	err = util.WriteFile(tlsIO.SignedKeyPath(tlsIO.MinionTLSDir),
		[]byte(serverKey), 0644)
	assert.NoError(t, err)
	err = startVaultContainer(dk, "")
	assert.NoError(t, err)

	dkcs, err := dk.List(nil)
	assert.NoError(t, err)
	assert.Len(t, dkcs, 1)

	// Ensure the expected image was booted.
	assert.True(t, strings.HasPrefix(dkcs[0].Image, "vault"))

	// Ensure that the Vault config is parseable.
	config := dkcs[0].Env["VAULT_LOCAL_CONFIG"]
	parsedConfig := map[string]interface{}{}
	err = json.Unmarshal([]byte(config), &parsedConfig)
	assert.NoError(t, err)

	// Ensure that the referenced TLS credentials were properly added to the
	// config and that the corresponding files were uploaded to the container.
	listenerConfig := parsedConfig["listener"].(map[string]interface{})
	tcpConfig := listenerConfig["tcp"].(map[string]interface{})
	serverCertPath := tcpConfig["tls_cert_file"].(string)
	serverKeyPath := tcpConfig["tls_key_file"].(string)
	caCertPath := tcpConfig["tls_client_ca_file"].(string)

	assert.Contains(t, md.Uploads, docker.UploadToContainerOptions{
		ContainerID: dkcs[0].ID,
		UploadPath:  filepath.Dir(serverCertPath),
		TarPath:     filepath.Base(serverCertPath),
		Contents:    serverCert,
	})

	assert.Contains(t, md.Uploads, docker.UploadToContainerOptions{
		ContainerID: dkcs[0].ID,
		UploadPath:  filepath.Dir(serverKeyPath),
		TarPath:     filepath.Base(serverKeyPath),
		Contents:    serverKey,
	})

	assert.Contains(t, md.Uploads, docker.UploadToContainerOptions{
		ContainerID: dkcs[0].ID,
		UploadPath:  filepath.Dir(caCertPath),
		TarPath:     filepath.Base(caCertPath),
		Contents:    caCert,
	})
}

func TestStartAndBootstrapVault(t *testing.T) {
	_, dk := docker.NewMock()
	mockAPIClient := &mocks.APIClient{}

	unsealKey := "unsealKey"
	rootToken := "rootToken"

	mockAPIClient.On("InitStatus").Return(false, nil)
	newVaultAPIClient = func(_ string) (APIClient, error) {
		return mockAPIClient, nil
	}

	// Ensure the certificate authentication endpoint is enabled.
	mockAPIClient.On("EnableAuth", certMountName, "cert", mock.Anything).
		Return(nil, nil)

	// Ensure that the returned unseal key is properly used to unseal Vault,
	// and the root token is set in the returned client.
	mockAPIClient.On("Init", &vaultAPI.InitRequest{
		SecretShares:    1,
		SecretThreshold: 1,
	}).Return(&vaultAPI.InitResponse{
		Keys:      []string{unsealKey},
		RootToken: rootToken,
	}, nil)

	mockAPIClient.On("Unseal", unsealKey).Return(nil, nil)
	mockAPIClient.On("SetToken", rootToken).Return()

	startAndBootstrapVault(dk, "")
	mockAPIClient.AssertExpectations(t)
}

func TestExistingContainerRemoved(t *testing.T) {
	t.Parallel()

	// Test the error case where we are unable to get the Vault container's
	// running status.
	md, dk := docker.NewMock()
	md.ListError = true
	_, ok := startAndBootstrapVault(dk, "")
	assert.False(t, ok)

	// Check that no containers were booted.
	md.ListError = false
	runningContainers, err := dk.List(nil)
	assert.NoError(t, err)
	assert.Empty(t, runningContainers)

	// Test that if the Vault container is already running, we remove it and
	// try to start a new one.
	_, err = dk.Run(docker.RunOptions{Name: vaultContainerName, Image: "vault"})
	assert.NoError(t, err)

	// Mock a start error to avoid entering the code path of configuring Vault.
	md.StartError = true
	_, ok = startAndBootstrapVault(dk, "")
	assert.False(t, ok)

	// There should be no running containers because the previous Vault
	// container was removed.
	runningContainers, err = dk.List(nil)
	assert.NoError(t, err)
	assert.Empty(t, runningContainers)
}
