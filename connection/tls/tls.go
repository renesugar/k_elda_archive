package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TLS satisfies the connection.Credentials interface for configuring grpc
// connections encrypted by TLS.
// It requires the public certificate of a certificate authority, and a
// certificate key pair signed by the given certificate authority. Communication
// is only allowed if the peer has a valid public key signed by the given
// certificate authority.
// The rsa subpackage contains code to generate certificates compatible with
// this authentication scheme.
type TLS struct {
	keyPair tls.Certificate
	caPool  *x509.CertPool
}

// ServerOpts gets the grpc options for creating a server.
func (tlsAuth TLS) ServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{grpc.Creds(
		credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{tlsAuth.keyPair},
			ClientCAs:    tlsAuth.caPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}),
	)}
}

// ClientOpts gets the grpc options for connecting as a client.
func (tlsAuth TLS) ClientOpts() []grpc.DialOption {
	return []grpc.DialOption{grpc.WithTransportCredentials(
		credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{tlsAuth.keyPair},

			// We use a custom VerifyPeerCertificate that only checks whether
			// the certificate is signed by the expected CA, and ignores
			// the server's hostname. This greatly simplifies the certificate
			// generation logic because it doesn't need to account for IP
			// address changes. This is safe to do because the client only
			// trusts a single CA, and we have complete control over what
			// certificates the CA signs.
			InsecureSkipVerify:    true,
			VerifyPeerCertificate: tlsAuth.verifySignedByCA,
		}),
	)}
}

// verifySignedByCA verifies that at least one certificate is signed by the
// expected CA. It is different from the default implementation because it does
// verify the peer's hostname.
func (tlsAuth TLS) verifySignedByCA(rawCertsSlice [][]byte,
	_ [][]*x509.Certificate) error {
	var verifyErrors []string
	for _, rawCerts := range rawCertsSlice {
		parsedCerts, err := x509.ParseCertificates(rawCerts)
		if err != nil {
			verifyErrors = append(verifyErrors, err.Error())
			continue
		}

		for _, cert := range parsedCerts {
			_, err = cert.Verify(x509.VerifyOptions{
				Roots: tlsAuth.caPool,
			})
			if err == nil {
				return nil
			}

			verifyErrors = append(verifyErrors, err.Error())
		}
	}

	return fmt.Errorf("failed to verify peer certificates: [%s]",
		strings.Join(verifyErrors, ", "))
}

// New creates a TLS instance from the given CA and signed certificate and key.
func New(ca, cert, key string) (TLS, error) {
	keyPair, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return TLS{}, err
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM([]byte(ca)) {
		return TLS{}, errors.New("failed to create CA cert pool")
	}

	return TLS{keyPair, caPool}, nil
}
