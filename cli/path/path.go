package path

import (
	"os"
	"path/filepath"
)

var (
	// quiltHome is where Quilt configuration (such as TLS credentials) are
	// stored.
	// The HOME environment variable is the user's home directory on all POSIX
	// compatible systems.
	quiltHome = filepath.Join(os.Getenv("HOME"), ".quilt")

	// DefaultTLSDir is the default location for users to store TLS credentials.
	DefaultTLSDir = filepath.Join(quiltHome, "tls")
)
