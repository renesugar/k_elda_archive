// Package io is used to ensure that the various pieces of code that interact
// with credentials read and write files to consistent locations.
package io

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kelda/kelda/connection/tls"
	"github.com/kelda/kelda/connection/tls/rsa"
	"github.com/kelda/kelda/util"
)

const (
	// MinionTLSDir is the directory in which the daemon will place TLS
	// credentials on cloud machines.
	MinionTLSDir = "/home/kelda/.kelda/tls"

	caCertFilename     = "certificate_authority.crt"
	caKeyFilename      = "certificate_authority.key"
	signedCertFilename = "kelda.crt"
	signedKeyFilename  = "kelda.key"
)

// File represents a file to be written to the filesystem.
type File struct {
	Path    string
	Content string
	Mode    os.FileMode
}

// ReadCredentials reads the TLS credentials contained within the directory.
func ReadCredentials(dir string) (tls.TLS, error) {
	caCert, err := util.ReadFile(CACertPath(dir))
	if err != nil {
		return tls.TLS{}, fmt.Errorf("read CA: %s", err)
	}

	signedCert, err := util.ReadFile(SignedCertPath(dir))
	if err != nil {
		return tls.TLS{}, fmt.Errorf("read signed cert: %s", err)
	}

	signedKey, err := util.ReadFile(SignedKeyPath(dir))
	if err != nil {
		return tls.TLS{}, fmt.Errorf("read signed key: %s", err)
	}

	return tls.New(caCert, signedCert, signedKey)
}

// ReadCA reads the certificate authority contained with the directory.
func ReadCA(dir string) (rsa.KeyPair, error) {
	caCert, err := util.ReadFile(CACertPath(dir))
	if err != nil {
		return rsa.KeyPair{}, fmt.Errorf("read cert: %s", err)
	}

	caKey, err := util.ReadFile(CAKeyPath(dir))
	if err != nil {
		return rsa.KeyPair{}, fmt.Errorf("read key: %s", err)
	}

	return rsa.New(caCert, caKey)
}

// MinionFiles defines how files should be written to disk for installation on
// minions.
func MinionFiles(dir string, ca, signed rsa.KeyPair) []File {
	return []File{
		{Path: CACertPath(dir), Content: ca.CertString(), Mode: 0644},
		{Path: SignedCertPath(dir), Content: signed.CertString(), Mode: 0644},
		{Path: SignedKeyPath(dir), Content: signed.PrivateKeyString(),
			Mode: 0600},
	}
}

// DaemonFiles defines how files should be written to disk for use by the daemon.
func DaemonFiles(dir string, ca, signed rsa.KeyPair) []File {
	return append(MinionFiles(dir, ca, signed),
		File{Path: CAKeyPath(dir), Content: ca.PrivateKeyString(), Mode: 0600})
}

// CACertPath defines where to write the certificate for the certificate authority.
func CACertPath(dir string) string {
	return filepath.Join(dir, caCertFilename)
}

// CAKeyPath defines where to write the private key for the certificate authority.
func CAKeyPath(dir string) string {
	return filepath.Join(dir, caKeyFilename)
}

// SignedCertPath defines where to write the certificate for the signed certificate.
func SignedCertPath(dir string) string {
	return filepath.Join(dir, signedCertFilename)
}

// SignedKeyPath defines where to write the private key for the signed certificate.
func SignedKeyPath(dir string) string {
	return filepath.Join(dir, signedKeyFilename)
}
