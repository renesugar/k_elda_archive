package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

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
			RootCAs:      tlsAuth.caPool,
		}),
	)}
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
