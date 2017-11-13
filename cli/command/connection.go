package command

import (
	"flag"
	"os"

	"github.com/kelda/kelda/api"
	"github.com/kelda/kelda/api/client"
	cliPath "github.com/kelda/kelda/cli/path"
	"github.com/kelda/kelda/connection"
	tlsIO "github.com/kelda/kelda/connection/tls/io"
)

type connectionFlags struct {
	host string
}

func (cf *connectionFlags) InstallFlags(flags *flag.FlagSet) {
	defaultSocket := os.Getenv("KELDA_HOST")
	if defaultSocket == "" {
		defaultSocket = api.DefaultSocket
	}
	flags.StringVar(&cf.host, "H", defaultSocket, "the host to connect to. This "+
		"flag can also be specified by setting the KELDA_HOST environment "+
		"variable. If the flag is set using both the environment variable and a "+
		"command line argument, the command line value takes precedence.")
}

type connectionHelper struct {
	creds  connection.Credentials
	client client.Client

	connectionFlags
}

func (ch *connectionHelper) BeforeRun() (err error) {
	// Load the credentials that will be used by Kelda clients and servers.
	ch.creds, err = tlsIO.ReadCredentials(cliPath.DefaultTLSDir)
	if err != nil {
		return err
	}
	return ch.setupClient(client.New)
}

func (ch *connectionHelper) AfterRun() error {
	return ch.client.Close()
}

func (ch *connectionHelper) setupClient(getter client.Getter) (err error) {
	ch.client, err = getter(ch.host, ch.creds)
	return err
}
