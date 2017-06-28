package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/connection/credentials"
	"github.com/quilt/quilt/db"
)

func TestEtcd(t *testing.T) {
	clnt, err := client.New(api.DefaultSocket, credentials.Insecure{})
	if err != nil {
		t.Fatalf("couldn't get quiltctl client: %s", err)
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		t.Fatalf("couldn't query containers: %s", err)
	}

	test(t, containers)
}

func test(t *testing.T, containers []db.Container) {
	for _, c := range containers {
		if !strings.Contains(c.Image, "etcd") {
			continue
		}

		fmt.Printf("Checking etcd health from %s\n", c.StitchID)
		out, err := exec.Command("quilt", "ssh", c.StitchID,
			"etcdctl", "cluster-health").CombinedOutput()
		fmt.Println(string(out))
		if err != nil || !strings.Contains(string(out), "cluster is healthy") {
			t.Errorf("cluster is unhealthy: %s", err)
		}
	}
}
