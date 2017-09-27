package command

import (
	"crypto/rand"
	goRSA "crypto/rsa"
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/quilt/quilt/api/server"
	cliPath "github.com/quilt/quilt/cli/path"
	"github.com/quilt/quilt/cloud"
	tlsIO "github.com/quilt/quilt/connection/tls/io"
	"github.com/quilt/quilt/connection/tls/rsa"
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/engine"
	"github.com/quilt/quilt/util"
	"github.com/quilt/quilt/version"

	log "github.com/sirupsen/logrus"
)

// Daemon contains the options for running the Quilt daemon.
type Daemon struct {
	adminSSHPrivateKey string

	*connectionFlags
}

// NewDaemonCommand creates a new Daemon command instance.
func NewDaemonCommand() *Daemon {
	return &Daemon{
		connectionFlags: &connectionFlags{},
	}
}

var daemonCommands = "quilt daemon [OPTIONS]"
var daemonExplanation = "Start the quilt daemon, which listens for quilt API requests."

// InstallFlags sets up parsing for command line flags
func (dCmd *Daemon) InstallFlags(flags *flag.FlagSet) {
	dCmd.connectionFlags.InstallFlags(flags)
	flags.StringVar(&dCmd.adminSSHPrivateKey, "admin-ssh-private-key", "",
		"if specified, all machines will be configured to allow access from "+
			"this private SSH key")
	flags.Usage = func() {
		util.PrintUsageString(daemonCommands, daemonExplanation, flags)
	}
}

// Parse parses the command line arguments for the daemon command.
func (dCmd *Daemon) Parse(args []string) error {
	return nil
}

// BeforeRun makes any necessary post-parsing transformations.
func (dCmd *Daemon) BeforeRun() error {
	return nil
}

// AfterRun performs any necessary post-run cleanup.
func (dCmd *Daemon) AfterRun() error {
	return nil
}

// Run starts the daemon.
func (dCmd *Daemon) Run() int {
	log.WithField("version", version.Version).Info("Starting Quilt daemon")

	// If the TLS credentials do not exist, autogenerate credentials and write
	// them to disk.
	if _, err := util.Stat(cliPath.DefaultTLSDir); os.IsNotExist(err) {
		log.Infof("TLS credentials not found in %s, so generating credentials "+
			"and writing to disk", cliPath.DefaultTLSDir)
		if err := setupTLS(cliPath.DefaultTLSDir); err != nil {
			log.WithError(err).WithField("path", cliPath.DefaultTLSDir).Error(
				"TLS credential generation failed")
			return 1
		}
	}

	var sshKey ssh.Signer
	if dCmd.adminSSHPrivateKey != "" {
		var err error
		sshKey, err = parseSSHPrivateKey(dCmd.adminSSHPrivateKey)
		if err != nil {
			log.WithError(err).Errorf(
				"Failed to parse private key %s", dCmd.adminSSHPrivateKey)
			return 1
		}
	} else {
		log.Info("No admin key supplied, which is required " +
			"to copy TLS credentials to the cluster. " +
			"Auto-generating an in-memory key.")
		var err error
		sshKey, err = newSSHPrivateKey()
		if err != nil {
			log.WithError(err).Error("Failed to generate SSH key")
			return 1
		}
	}

	creds, err := tlsIO.ReadCredentials(cliPath.DefaultTLSDir)
	if err != nil {
		log.WithError(err).Error("Failed to parse TLS credentials")
		return 1
	}

	conn := db.New()
	go engine.Run(conn, getPublicKey(sshKey))
	go server.Run(conn, dCmd.host, true, creds)

	ca, err := tlsIO.ReadCA(cliPath.DefaultTLSDir)
	if err != nil {
		log.WithError(err).WithField("path", cliPath.DefaultTLSDir).Error(
			"Failed to parse certificate authority")
		return 1
	}

	go cloud.SyncCredentials(conn, sshKey, ca)
	cloud.Run(conn, creds)
	return 0
}

func newSSHPrivateKey() (ssh.Signer, error) {
	key, err := goRSA.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return ssh.NewSignerFromKey(key)
}

func parseSSHPrivateKey(path string) (ssh.Signer, error) {
	keyStr, err := util.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %s", err)
	}

	return ssh.ParsePrivateKey([]byte(keyStr))
}

func getPublicKey(sshPrivKey ssh.Signer) string {
	if sshPrivKey == nil {
		return ""
	}
	pubKey := base64.StdEncoding.EncodeToString(sshPrivKey.PublicKey().Marshal())
	pubKeyType := sshPrivKey.PublicKey().Type()
	return pubKeyType + " " + pubKey
}

func setupTLS(outDir string) error {
	if err := util.AppFs.MkdirAll(outDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %s", err)
	}

	ca, err := rsa.NewCertificateAuthority()
	if err != nil {
		return fmt.Errorf("failed to create CA: %s", err)
	}

	// Generate a signed certificate for use by the Daemon server, and client
	// connections.
	signed, err := rsa.NewSigned(ca)
	if err != nil {
		return fmt.Errorf("failed to create signed key pair: %s", err)
	}

	for _, f := range tlsIO.DaemonFiles(outDir, ca, signed) {
		if err := util.WriteFile(f.Path, []byte(f.Content), f.Mode); err != nil {
			return fmt.Errorf("failed to write file (%s): %s", f.Path, err)
		}
	}

	return nil
}
