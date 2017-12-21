package vault

import (
	"fmt"
	"time"

	tlsIO "github.com/kelda/kelda/connection/tls/io"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/docker"
	"github.com/kelda/kelda/util"

	vaultAPI "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

/**
* This submodule configures the Vault container on the masters. It ensures
* that the Vault container is running, and configures the access policies to
* Vault.
*
* All Vault clients authenticate using TLS certificates. Workers are allowed
* to read the secrets necessary for their scheduled containers to run, and
* Masters are allowed to write secrets.
*
* Vault access is configured in two parts:  authentication, and policies.
* Policies dictate the actions clients can perform on the given paths. For
* example, A client may be allowed to READ to /secret/kelda/foo. Authentication
* is used to map an identity to a set of policies.
*
* In our design, there is a policy for each minion, and exactly one
* authentication role that maps to the minion's policy. Authentication is
* performed using TLS certificates. The daemon reports the minion to public key
* mappings to the masters, and the masters create authentication rules that
* require clients to possess the associated private key.
*
* Note that although a Vault container is started on each master, no data is
* actually shared between them. Each master runs a separate in-memory version
* of Vault. Furthermore, only the version of Vault running on the leader is
* used -- the `kelda secret` command publishes the secret to the leader, and
* the workers query secrets from the leader. Therefore, if leadership transfers
* to another master (or if the Vault container on the leader crashes), `kelda
* secret` will need to be run again in order for secrets to be populated on the
* new masters.
* Although the Vault instances on the standby masters aren't actively used,
* they are still booted to facilitate quicker failover, and simplify the logic
* around when to run Vault.
**/

// Run implements the logic necessary for the masters to run Vault. It boots a
// Vault container, blocks until Vault is booted, and then continuously checks
// the Kelda database to determine the policies that Vault should implement for
// secret access, and updates Vault to reflect these policies.
func Run(conn db.Conn, dk docker.Client, vaultClient APIClient) {
	for range conn.TriggerTick(30, db.MinionTable, db.ContainerTable).C {
		syncPolicies(vaultClient, conn)
		syncAuth(vaultClient, conn)
	}
}

// ContainerName is the name assigned to the Docker container running Vault.
// Assigning the container a consistent name rather than allowing Docker to
// pick a name at random makes it more obvious that the container is a Kelda
// system when running `docker ps`. Furthermore, it simplifies the logic to
// check whether Kelda has already booted a Vault container.
const ContainerName = "vault"

// Start repeatedly tries to start and bootstrap the Vault container until it
// succeeds. This way, we recover from any transient errors that might occur
// when starting Vault, such as a failure to pull the Vault Docker image.
func Start(conn db.Conn, dk docker.Client) APIClient {
	var myIP string
	for range conn.TriggerTick(5, db.MinionTable).C {
		myIP = conn.MinionSelf().PrivateIP
		if myIP != "" {
			break
		}
	}

	for {
		if client, ok := startAndBootstrapVault(dk, myIP); ok {
			return client
		}
		log.Warn("Failed to start Vault. Retrying in 30 seconds")
		time.Sleep(30 * time.Second)
	}
}

func startAndBootstrapVault(dk docker.Client, listenAddr string) (APIClient, bool) {
	// If there is a Vault container already running, we remove it and start a
	// new one. This simplifies the logic for configuring Vault because it does
	// not have to account for whether previous calls (such as Init) had
	// succeeded.
	// It is safe to remove the already running Vault container because if this
	// function is called, it means that Vault has not yet come up
	// successfully, and so there is not yet any important state stored.
	isRunning, err := dk.IsRunning(ContainerName)
	if err != nil {
		log.WithError(err).Error(
			"Failed to get running status for the Vault container")
		return nil, false
	}

	if isRunning {
		if err := dk.Remove(ContainerName); err != nil {
			log.WithError(err).Error(
				"Failed to remove currently running Vault container")
			return nil, false
		}
	}

	if err := startVaultContainer(dk, listenAddr); err != nil {
		log.WithError(err).Error("Failed to start Vault container")
		return nil, false
	}

	// Block until the Vault container responds to API requests. We may need to
	// try multiple times because the container requires some setup at boot
	// before being accessible over the network.
	var client APIClient
	var getClientError error
	err = util.BackoffWaitFor(func() bool {
		client, getClientError = newVaultAPIClient(listenAddr)
		if err != nil {
			log.WithError(err).Error("Failed to get Vault client")
			return false
		}

		_, getClientError = client.InitStatus()
		if getClientError == nil {
			return true
		}

		log.WithError(getClientError).Debug("Failed to connect to Vault. " +
			"This is expected when the container first boots.")
		return false
	}, 10*time.Second, 5*time.Minute)
	if err != nil {
		log.WithFields(log.Fields{
			"waitError":         err,
			"finalConnectError": err,
		}).Error("Failed to connect to Vault")
		return nil, false
	}

	// Ensure that Vault is unsealed by generating an unseal key using the Init
	// endpoint, and then Unsealing with the generated key. We only use a single
	// secret to unseal Vault. Although it is usually recommended to allow
	// a subset of secrets to be used to unseal Vault in case an operator is not
	// present to provide a secret, this does not apply to Kelda because the
	// unsealing is completely automated. Furthermore, increasing the number of
	// secrets would not increase security because the unseal key never leaves
	// the minion's memory.
	// Because Vault is running with the in-memory backend, Init, Unseal, and
	// EnabledAuth must be called again whenever the Vault container restarts.
	initResp, err := client.Init(&vaultAPI.InitRequest{
		// Number of unseal keys to return.
		SecretShares: 1,

		// Minimum number of unseal keys that must be provided to unseal Vault.
		SecretThreshold: 1,
	})
	if err != nil {
		log.WithError(err).Error("Failed to generate Vault unseal key")
		return nil, false
	}

	if _, err = client.Unseal(initResp.Keys[0]); err != nil {
		log.WithError(err).Error("Failed to unseal Vault")
		return nil, false
	}

	rootToken := initResp.RootToken
	client.SetToken(rootToken)
	err = client.EnableAuth(certMountName, "cert", "cert auth for minions")
	if err != nil {
		log.WithError(err).Error("Failed to enable Vault authentication")
		return nil, false
	}

	return client, true
}

// startVaultContainer reads the minion's TLS certificates, places them into
// the Vault filesystem, and boots the Vault container in server mode.
func startVaultContainer(dk docker.Client, listenAddr string) error {
	caCert, err := util.ReadFile(tlsIO.CACertPath(tlsIO.MinionTLSDir))
	if err != nil {
		return fmt.Errorf("failed to read Vault CA: %s", err)
	}

	serverCert, err := util.ReadFile(tlsIO.SignedCertPath(tlsIO.MinionTLSDir))
	if err != nil {
		return fmt.Errorf("failed to read Vault server certificate: %s", err)
	}

	serverKey, err := util.ReadFile(tlsIO.SignedKeyPath(tlsIO.MinionTLSDir))
	if err != nil {
		return fmt.Errorf("failed to read Vault server key: %s", err)
	}

	serverCertPath := "/server.crt"
	serverKeyPath := "/server.key"
	caCertPath := "/ca.crt"
	config := fmt.Sprintf(`{
		"backend": { "inmem": {} },
		"listener": { "tcp": {
			"address": "%s:%d",
			"tls_cert_file": %q,
			"tls_key_file": %q,
			"tls_client_ca_file": %q,
			"tls_require_and_verify_client_cert": "true"
		}}
	}`, listenAddr, vaultPort, serverCertPath, serverKeyPath, caCertPath)

	ro := docker.RunOptions{
		Name:        ContainerName,
		NetworkMode: "host",
		Image:       "vault:0.9.1",
		Args:        []string{"server"},
		Env:         map[string]string{"VAULT_LOCAL_CONFIG": config},
		FilepathToContent: map[string]string{
			caCertPath:     caCert,
			serverCertPath: serverCert,
			serverKeyPath:  serverKey,
		},
		// Vault requires IPC_LOCK to be enabled in order to prevent swapping
		// sensitive values to disk.
		CapAdd: []string{"IPC_LOCK"},
	}
	_, err = dk.Run(ro)
	return err
}
