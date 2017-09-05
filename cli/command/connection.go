package command

import (
	"flag"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/cli/command/credentials"
	"github.com/quilt/quilt/connection"
)

type connectionFlags struct {
	host   string
	tlsDir string
}

func (cf *connectionFlags) InstallFlags(flags *flag.FlagSet) {
	flags.StringVar(&cf.host, "H", api.DefaultSocket, "the host to connect to")
	flags.StringVar(&cf.tlsDir, "tls-dir", "",
		"the directory in which to lookup tls certs")
}

type connectionHelper struct {
	creds  connection.Credentials
	client client.Client

	connectionFlags
}

func (ch *connectionHelper) BeforeRun() (err error) {
	// Load the credentials that will be used by Quilt clients and servers.
	ch.creds, err = credentials.Read(ch.tlsDir)
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
