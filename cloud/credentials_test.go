package cloud

import (
	"crypto/rand"
	goRSA "crypto/rsa"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"

	tlsIO "github.com/kelda/kelda/connection/tls/io"
	"github.com/kelda/kelda/connection/tls/rsa"
	"github.com/kelda/kelda/db"
)

// Test the success path when generating and installing credentials on a new
// machine.
func TestSyncCredentials(t *testing.T) {
	key, err := goRSA.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	expSigner, err := ssh.NewSignerFromKey(key)
	assert.NoError(t, err)

	expHost := "8.8.8.8"
	mockFs := afero.NewMemMapFs()
	conn := db.New()

	getSftpFs = func(host string, signer ssh.Signer) (sftpFs, error) {
		assert.Equal(t, expSigner, signer)
		assert.Equal(t, expHost, host)
		return mockSFTPFs{mockFs}, nil
	}

	ca, err := rsa.NewCertificateAuthority()
	assert.NoError(t, err)

	conn.Txn(db.MachineTable).Run(func(view db.Database) error {
		dbm := view.InsertMachine()
		dbm.PublicIP = expHost
		dbm.PrivateIP = "9.9.9.9"
		view.Commit(dbm)
		return nil
	})
	syncCredentialsOnce(conn, expSigner, ca)

	aferoFs := afero.Afero{Fs: mockFs}
	certBytes, err := aferoFs.ReadFile(tlsIO.SignedCertPath(tlsIO.MinionTLSDir))
	assert.NoError(t, err)
	assert.NotEmpty(t, certBytes)

	keyBytes, err := aferoFs.ReadFile(tlsIO.SignedKeyPath(tlsIO.MinionTLSDir))
	assert.NoError(t, err)
	assert.NotEmpty(t, keyBytes)

	caBytes, err := aferoFs.ReadFile(tlsIO.CACertPath(tlsIO.MinionTLSDir))
	assert.NoError(t, err)
	assert.NotEmpty(t, caBytes)

	// Ensure that the machine's public key got written to the database.
	dbm := conn.SelectFromMachine(nil)[0]
	assert.Equal(t, string(certBytes), dbm.PublicKey)
}

func TestFailedToSSH(t *testing.T) {
	ca, err := rsa.NewCertificateAuthority()
	assert.NoError(t, err)

	conn := db.New()
	conn.Txn(db.MachineTable).Run(func(view db.Database) error {
		dbm := view.InsertMachine()
		dbm.PublicIP = "8.8.8.8"
		dbm.PrivateIP = "9.9.9.9"
		view.Commit(dbm)
		return nil
	})

	getSftpFs = func(host string, _ ssh.Signer) (sftpFs, error) {
		return nil, assert.AnError
	}

	syncCredentialsOnce(conn, nil, ca)
	// The machine's PublicKey should not be set, because Kelda should have
	// given up and not set the public key when getting an SFTP client failed.
	assert.Empty(t, conn.SelectFromMachine(nil)[0].PublicKey)
}

type mockSFTPFs struct {
	afero.Fs
}

func (fs mockSFTPFs) Close() error {
	return nil
}
