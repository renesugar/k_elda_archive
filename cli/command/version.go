package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/kelda/kelda/api/client"
	cliPath "github.com/kelda/kelda/cli/path"
	"github.com/kelda/kelda/connection/tls"
	tlsIO "github.com/kelda/kelda/connection/tls/io"
	"github.com/kelda/kelda/util"
	"github.com/kelda/kelda/version"

	log "github.com/sirupsen/logrus"
)

// Version prints the Kelda version information.
type Version struct {
	connectionFlags
}

// NewVersionCommand creates a new Version command instance.
func NewVersionCommand() *Version {
	return &Version{}
}

var versionCommands = "kelda version [OPTIONS]"
var versionExplanation = "Show the Kelda version information."

// InstallFlags sets up parsing for command line flags.
func (vCmd *Version) InstallFlags(flags *flag.FlagSet) {
	vCmd.connectionFlags.InstallFlags(flags)
	flags.Usage = func() {
		util.PrintUsageString(versionCommands, versionExplanation, flags)
	}
}

// Parse parses the command line arguments for the version command.
func (vCmd *Version) Parse(args []string) error {
	return nil
}

// Run prints the version information.
func (vCmd *Version) Run() int {
	fmt.Println("Client Version:", version.Version)

	fetchingVersionStr := "Fetching daemon version..."
	fmt.Print(fetchingVersionStr)
	daemonVersion, err := vCmd.fetchDaemonVersion(tlsIO.ReadCredentials, client.New)
	fmt.Print("\r" + strings.Repeat(" ", len(fetchingVersionStr)) + "\r")

	if err != nil {
		log.WithError(err).Error("Failed to fetch daemon version.")
		return 1
	}

	fmt.Println("Daemon Version:", daemonVersion)
	return 0
}

type credsGetter func(dir string) (tls.TLS, error)

func (vCmd *Version) fetchDaemonVersion(credsGetter credsGetter,
	clientGetter client.Getter) (string, error) {
	creds, err := credsGetter(cliPath.DefaultTLSDir)
	if err != nil {
		return "", err
	}

	c, err := clientGetter(vCmd.host, creds)
	if err != nil {
		return "", err
	}
	defer c.Close()
	return c.Version()
}

// BeforeRun makes any necessary post-parsing transformations.
func (vCmd *Version) BeforeRun() error {
	return nil
}

// AfterRun performs any necessary post-run cleanup.
func (vCmd *Version) AfterRun() error {
	return nil
}
