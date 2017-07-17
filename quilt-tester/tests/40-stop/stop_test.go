package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/minion/supervisor/images"

	log "github.com/Sirupsen/logrus"
)

func TestStopContainer(t *testing.T) {
	if err := exec.Command("quilt", "stop", "-containers").Run(); err != nil {
		t.Fatalf("couldn't run stop command: %s", err.Error())
	}

	c, err := client.New(api.DefaultSocket)
	if err != nil {
		t.Fatalf("couldn't get quiltctl client: %s", err.Error())
	}
	defer c.Close()

	var passed bool
	tries := 1
	for ; tries <= 5; tries++ {
		log.Info("Sleeping thirty seconds for `quilt stop -containers` " +
			"to take effect")
		time.Sleep(30 * time.Second)

		machines, err := c.QueryMachines()
		if err != nil {
			t.Fatalf("couldn't query machines: %s", err.Error())
		}

		passed = true
		for _, m := range machines {
			containersRaw, err := exec.Command("quilt", "ssh", m.StitchID,
				"docker", "ps", "--format", "{{.Names}}").Output()
			if err != nil {
				passed = false
				t.Errorf("couldn't run `docker ps`: %s", err.Error())
				continue
			}

			containersStr := strings.TrimSpace(string(containersRaw))
			fmt.Printf("Containers on machine %s:\n", m.StitchID)
			fmt.Println(string(containersRaw))

			names := strings.Split(containersStr, "\n")
			if len(filterQuiltContainers(names)) > 0 {
				passed = false
				t.Errorf("machine %s has unexpected containers",
					m.StitchID)
			}
		}

		if passed {
			break
		}
	}

	switch {
	case tries > 1 && passed:
		t.Logf("Although the containers weren't stopped on the first check, "+
			"they eventually stopped on try %d.", tries)
	case !passed:
		t.Error("Some containers were never stopped")
	}
}

var quiltContainers = map[string]struct{}{
	images.Etcd:          {},
	images.Ovncontroller: {},
	images.Ovnnorthd:     {},
	images.Ovsdb:         {},
	images.Ovsvswitchd:   {},
	images.Registry:      {},
	"minion":             {},
}

func filterQuiltContainers(containers []string) (filtered []string) {
	for _, c := range containers {
		if _, ok := quiltContainers[c]; !ok && c != "" {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
