package command

import (
	"flag"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	cliPath "github.com/quilt/quilt/cli/path"
	"github.com/quilt/quilt/connection"
	tlsIO "github.com/quilt/quilt/connection/tls/io"
)

type connectionFlags struct {
	host string
}

func (cf *connectionFlags) InstallFlags(flags *flag.FlagSet) {
	flags.StringVar(&cf.host, "H", api.DefaultSocket, "the host to connect to")
}

type connectionHelper struct {
	creds  connection.Credentials
	client client.Client

	connectionFlags
}

func (ch *connectionHelper) BeforeRun() (err error) {
	// Load the credentials that will be used by Quilt clients and servers.
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
