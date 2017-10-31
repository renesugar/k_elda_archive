package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/kelda/kelda/api"
	"github.com/kelda/kelda/api/client"
	cliPath "github.com/kelda/kelda/cli/path"
	"github.com/kelda/kelda/connection/tls"
	tlsIO "github.com/kelda/kelda/connection/tls/io"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/integration-tester/util"

	"github.com/stretchr/testify/assert"
)

func TestSecretValues(t *testing.T) {
	clnt, err := util.GetDefaultDaemonClient()
	if err != nil {
		t.Fatalf("couldn't get api client: %s", err)
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		t.Fatalf("couldn't query containers: %s", err)
	}

	assert.Len(t, containers, 1)
	containerID := containers[0].BlueprintID

	// The secret values are defined in the 70-secret-setup test.
	// The non-secret values are defined in the blueprint accompanying this test.
	expEnv := map[string]string{
		"NOT_A_SECRET": "plaintext",
		"MY_SECRET":    "env secret",
	}

	expFiles := map[string]string{
		"/notASecret": "plaintext",
		"/mySecret":   "file secret",
	}

	for key, expVal := range expEnv {
		actualVal, err := getEnv(containerID, key)
		assert.NoError(t, err)

		assert.Equal(t, expVal, actualVal)
		fmt.Printf("Expected environment variable %q to have value %q. Got %q.\n",
			key, expVal, actualVal)
	}

	for path, expVal := range expFiles {
		actualVal, err := getFile(containerID, path)
		assert.NoError(t, err)

		assert.Equal(t, expVal, actualVal)
		fmt.Printf("Expected file %q to have value %q. Got %q.\n",
			path, expVal, actualVal)
	}
}

func getEnv(containerID, key string) (string, error) {
	envBytes, err := exec.Command("kelda", "ssh", containerID,
		"printenv", key).Output()
	if err != nil {
		return "", err
	}

	// Trim the newline appended by printenv.
	return strings.TrimRight(string(envBytes), "\n"), nil
}

func getFile(containerID, path string) (string, error) {
	fileBytes, err := exec.Command("kelda", "ssh", containerID, "cat", path).Output()
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

func TestVaultACLs(t *testing.T) {
	clnt, err := util.GetDefaultDaemonClient()
	if err != nil {
		t.Fatalf("couldn't get api client: %s", err)
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		t.Fatalf("couldn't query containers: %s", err)
	}

	machines, err := clnt.QueryMachines()
	if err != nil {
		t.Fatalf("couldn't query machines: %s", err)
	}

	tlsCreds, err := tlsIO.ReadCredentials(cliPath.DefaultTLSDir)
	if err != nil {
		t.Fatalf("couldn't read TLS credentials: %s", err)
	}

	leaderIP, err := getLeaderIP(machines, tlsCreds)
	if err != nil {
		t.Fatalf("couldn't get leader's IP: %s", err)
	}
	vaultAddr := fmt.Sprintf("https://%s:8200", leaderIP)

	secretNameToAllowedWorkers := map[string][]string{}
	for _, c := range containers {
		for _, secretName := range c.GetReferencedSecrets() {
			secretNameToAllowedWorkers[secretName] = append(
				secretNameToAllowedWorkers[secretName], c.Minion)
		}
	}

	for _, m := range machines {
		for secretName, allowedWorkers := range secretNameToAllowedWorkers {
			fmt.Printf("Attempting to fetch secret %q from %v.\n",
				secretName, m)
			output, err := tryFetchSecret(m.CloudID, vaultAddr, secretName)

			shouldSucceed := m.Role == db.Master ||
				containsString(allowedWorkers, m.PrivateIP)
			if shouldSucceed {
				fmt.Println(".. It should succeed.")
				assert.NoError(t, err)
			} else {
				fmt.Println(".. It should fail.")
				assert.NotNil(t, err)
			}

			fmt.Println(".... Error:", err)
			fmt.Println(".... Output:", string(output))
		}
	}
}

// tryFetchSecret attempts to fetch the given secretName using the credentials
// on the given machine.
func tryFetchSecret(machineID, vaultAddr, secretName string) ([]byte, error) {
	if exec.Command("kelda", "ssh", machineID, "which", "vault").Run() != nil {
		if err := setUpVault(machineID); err != nil {
			return nil, fmt.Errorf("set up vault: %s", err)
		}
	}

	vaultOpts := fmt.Sprintf("-address=%s -ca-cert=%s -client-cert=%s -client-key=%s",
		vaultAddr,
		tlsIO.CACertPath(tlsIO.MinionTLSDir),
		tlsIO.SignedCertPath(tlsIO.MinionTLSDir),
		tlsIO.SignedKeyPath(tlsIO.MinionTLSDir))
	vaultCmd := fmt.Sprintf("vault auth -method=cert %[1]s && "+
		"vault read %[1]s /secret/kelda/%[2]s",
		vaultOpts, secretName)
	return exec.Command("kelda", "ssh", machineID, "bash", "-c",
		fmt.Sprintf("%q", vaultCmd)).CombinedOutput()
}

const vaultReleaseAddress = "https://releases.hashicorp.com/vault/0.8.3/" +
	"vault_0.8.3_linux_amd64.zip"

// setUpVault installs Vault onto the given machine.
func setUpVault(machineID string) error {
	setupVaultCmd := fmt.Sprintf("apt-get install -y unzip && "+
		"workdir=$(mktemp -d) && "+
		"cd ${workdir} && "+
		"curl -s -o vault.zip %s && "+
		"unzip vault.zip && "+
		"cp vault /usr/local/bin && "+
		"rm -rf ${workdir}", vaultReleaseAddress)
	return exec.Command("kelda", "ssh", machineID, "sudo", "bash", "-c",
		fmt.Sprintf("%q", setupVaultCmd)).Run()
}

// Get the private IP of the leader of the cluster by querying each of the
// given machines.
func getLeaderIP(machines []db.Machine, tlsCreds tls.TLS) (string, error) {
	var errors []string
	for _, m := range machines {
		c, err := client.New(api.RemoteAddress(m.PublicIP), tlsCreds)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		etcds, err := c.QueryEtcd()
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		if len(etcds) != 1 || etcds[0].LeaderIP == "" {
			errors = append(errors,
				"no leader information on machine "+m.PublicIP)
			continue
		}

		return etcds[0].LeaderIP, nil
	}

	return "", fmt.Errorf(strings.Join(errors, ", "))
}

func containsString(slc []string, target string) bool {
	for _, item := range slc {
		if item == target {
			return true
		}
	}
	return false
}
