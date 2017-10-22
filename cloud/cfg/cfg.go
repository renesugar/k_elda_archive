package cfg

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	tlsIO "github.com/kelda/kelda/connection/tls/io"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/version"

	log "github.com/sirupsen/logrus"
)

const (
	keldaImage = "keldaio/kelda"
)

// Allow mocking out for the unit tests.
var ver = version.Version

// Ubuntu generates a cloud config file for the Ubuntu operating system with the
// corresponding `version`.
func Ubuntu(m db.Machine, inboundPublic string) string {
	t := template.Must(template.New("cloudConfig").Parse(cfgTemplate))

	img := fmt.Sprintf("%s:%s", keldaImage, ver)

	var cloudConfigBytes bytes.Buffer
	err := t.Execute(&cloudConfigBytes, struct {
		KeldaImage string
		SSHKeys    string
		LogLevel   string
		MinionOpts string
		TLSDir     string
	}{
		KeldaImage: img,
		SSHKeys:    strings.Join(m.SSHKeys, "\n"),
		LogLevel:   log.GetLevel().String(),
		MinionOpts: minionOptions(m.Role, inboundPublic),
		TLSDir:     tlsIO.MinionTLSDir,
	})
	if err != nil {
		panic(err)
	}

	return cloudConfigBytes.String()
}

func minionOptions(role db.Role, inboundPublic string) string {
	options := fmt.Sprintf("--role %q", role)

	if inboundPublic != "" {
		options += fmt.Sprintf(" --inbound-pub-intf %q", inboundPublic)
	}
	return options
}
