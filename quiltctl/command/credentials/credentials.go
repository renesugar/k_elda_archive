package credentials

import (
	"github.com/quilt/quilt/connection"
	"github.com/quilt/quilt/connection/credentials"
	tlsIO "github.com/quilt/quilt/connection/credentials/tls/io"
)

// XXX: This is in a subpackage to prevent a cyclic dependency where the minion
// command imports the minion package, which has to import this file. Once
// the quiltctl package structure is rearranged to have a package for each subcommand,
// this function can go in the same package as the client getter.

// Read attempts to load the credentials defined within the given directory.
// If `dir` is empty, then we return `credentials.Insecure`; otherwise, we attempt
// to load TLS credentials using default names.
func Read(dir string) (connection.Credentials, error) {
	if dir == "" {
		return credentials.Insecure{}, nil
	}

	return tlsFromFile(dir)
}

var tlsFromFile = tlsIO.ReadCredentials
