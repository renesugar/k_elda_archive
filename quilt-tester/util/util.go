package util

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/quilt/quilt/db"
)

// CheckPublicConnections test that HTTP GETs against all ports that are
// connected to the public internet succeed.
func CheckPublicConnections(t *testing.T, machines []db.Machine,
	containers []db.Container, connections []db.Connection) {

	// Map of label to its publicly exposed ports.
	pubConns := map[string][]int{}
	for _, conn := range connections {
		if conn.From == "public" {
			for port := conn.MinPort; port <= conn.MaxPort; port++ {
				pubConns[conn.To] = append(pubConns[conn.To], port)
			}
		}
	}

	mapper := newIPMapper(machines)
	for _, cont := range containers {
		contIP, err := mapper.containerIP(cont)
		if err != nil {
			t.Error(err)
			continue
		}
		for _, lbl := range cont.Labels {
			ports, ok := pubConns[lbl]
			if !ok {
				continue
			}
			for _, port := range ports {
				tryGet(t, contIP+":"+strconv.Itoa(port))
			}
		}
	}
}

func tryGet(t *testing.T, ip string) {
	log.Info("\n\n\nTesting ", ip)

	for i := 0; i < 10; i++ {
		resp, err := http.Get("http://" + ip)
		if err != nil {
			t.Errorf("%s - HTTP GET error: %s", ip, err)
			continue
		}

		if resp.StatusCode != 200 {
			t.Errorf("%s - bad response code: %d", ip, resp.StatusCode)
		}
		log.Info(resp)
	}
}

type ipMapper map[string]string

func newIPMapper(machines []db.Machine) ipMapper {
	mapper := make(ipMapper)
	for _, m := range machines {
		mapper[m.PrivateIP] = m.PublicIP
	}
	return mapper
}

func (mapper ipMapper) containerIP(c db.Container) (string, error) {
	ip, ok := mapper[c.Minion]
	if !ok {
		return "", fmt.Errorf("no public IP for %v", c)
	}
	return ip, nil
}
