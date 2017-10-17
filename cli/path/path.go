package path

import (
	"os"
	"path/filepath"
)

var (
	// quiltHome is where Kelda configuration (such as TLS credentials) are
	// stored.
	// The HOME environment variable is the user's home directory on all POSIX
	// compatible systems.
	quiltHome = filepath.Join(os.Getenv("HOME"), ".kelda")

	// DefaultTLSDir is the default location for users to store TLS credentials.
	DefaultTLSDir = filepath.Join(quiltHome, "tls")

	// DefaultSSHKeyPath is the default filepath where the private SSH key used
	// to access Kelda will be stored.
	DefaultSSHKeyPath = filepath.Join(quiltHome, "ssh_key")
)
