package main

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/connection"
	"github.com/quilt/quilt/connection/credentials"
	"github.com/quilt/quilt/connection/credentials/tls"
	tlsIO "github.com/quilt/quilt/connection/credentials/tls/io"
	"github.com/quilt/quilt/connection/credentials/tls/rsa"
	"github.com/quilt/quilt/db"
)

func TestCredentials(t *testing.T) {
	localClient, err := client.New(api.DefaultSocket, credentials.Insecure{})
	if err != nil {
		t.Fatalf("Failed to get local client: %s", err)
	}
	defer localClient.Close()

	machines, err := localClient.QueryMachines()
	if err != nil {
		t.Fatalf("Failed to query machines: %s", err)
	}

	tlsDir := os.Getenv("TLS_DIR")
	if tlsDir == "" {
		testInsecure(t, machines)
	} else {
		testTLS(t, machines, tlsDir)
	}
}

func testInsecure(t *testing.T, machines []db.Machine) {
	// Test that an insecure connection succeeds.
	fmt.Println("Connecting insecurely. It should succeed.")
	assert.NoError(t, tryCredentials(machines, credentials.Insecure{}))

	// Test that connecting using TLS fails.
	fmt.Println("Connecting with TLS. It should fail.")
	tlsCreds, err := randomTLSCredentials()
	assert.NoError(t, err)
	err = tryCredentials(machines, tlsCreds)
	assert.EqualError(t, err, "tls: first record does not look like a TLS handshake")
}

func testTLS(t *testing.T, machines []db.Machine, tlsDir string) {
	// Test that an insecure connection fails.
	fmt.Println("Connecting insecurely. It should fail.")
	err := tryCredentials(machines, credentials.Insecure{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(),
		"rpc error: code = Unavailable desc = transport is closing")

	// Test that connecting with the TLS directory credentials succeeds.
	fmt.Println("Connecting with the daemon's TLS directory. It should succeed.")
	tlsCreds, err := tlsIO.ReadCredentials(tlsDir)
	assert.NoError(t, err)
	err = tryCredentials(machines, tlsCreds)
	assert.NoError(t, err)

	// Test that connecting with random TLS credentials fails.
	fmt.Println("Connecting with the random TLS credentials. It should fail.")
	tlsCreds, err = randomTLSCredentials()
	assert.NoError(t, err)
	err = tryCredentials(machines, tlsCreds)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "x509: certificate signed by unknown authority")
}

// Try querying the containers from each of the machines using the given credentials.
func tryCredentials(machines []db.Machine, creds connection.Credentials) error {
	for _, m := range machines {
		c, err := client.New(api.RemoteAddress(m.PublicIP), creds)
		if err != nil {
			return err
		}
		defer c.Close()

		if _, err = c.QueryContainers(); err != nil {
			return err
		}
	}
	return nil
}

// Generate random TLS credentials.
func randomTLSCredentials() (tls.TLS, error) {
	ca, err := rsa.NewCertificateAuthority()
	if err != nil {
		return tls.TLS{}, err
	}

	signed, err := rsa.NewSigned(ca, net.IP{0, 0, 0, 0})
	if err != nil {
		return tls.TLS{}, err
	}

	return tls.New(ca.CertString(), signed.CertString(), signed.PrivateKeyString())
}
