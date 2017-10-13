package cfg

import (
	"testing"

	"github.com/kelda/kelda/db"

	log "github.com/sirupsen/logrus"
)

func TestCloudConfig(t *testing.T) {
	cfgTemplate = "({{.KeldaImage}}) ({{.SSHKeys}}) " +
		"({{.MinionOpts}}) ({{.LogLevel}}) ({{.DockerOpts}})"

	log.SetLevel(log.InfoLevel)
	ver = "master"
	res := Ubuntu(db.Machine{
		SSHKeys: []string{"a", "b"},
		Role:    db.Master,
	}, "")
	exp := "(keldaio/kelda:master) (a\nb) (--role \"Master\") (info)" +
		" (-v /home/kelda/.kelda/tls:/home/kelda/.kelda/tls:ro)"
	if res != exp {
		t.Errorf("res: %s\nexp: %s", res, exp)
	}

	log.SetLevel(log.DebugLevel)
	ver = "1.2.3"
	res = Ubuntu(db.Machine{
		SSHKeys: []string{"a", "b"},
		Role:    db.Worker,
	}, "ib")
	exp = "(keldaio/kelda:1.2.3) (a\nb) (--role \"Worker\"" +
		" --inbound-pub-intf \"ib\") (debug)" +
		" (-v /home/kelda/.kelda/tls:/home/kelda/.kelda/tls:ro)"
	if res != exp {
		t.Errorf("res: %s\nexp: %s", res, exp)
	}
}
