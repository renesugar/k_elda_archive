package main

import (
	"testing"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/connection/credentials"
	"github.com/quilt/quilt/integration-tester/util"
)

func TestInboundConnections(t *testing.T) {
	clnt, err := client.New(api.DefaultSocket, credentials.Insecure{})
	if err != nil {
		t.Fatalf("couldn't get api client: %s", err.Error())
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		t.Fatalf("couldn't query containers: %s", err.Error())
	}

	connections, err := clnt.QueryConnections()
	if err != nil {
		t.Fatalf("couldn't query connections: %s", err.Error())
	}

	machines, err := clnt.QueryMachines()
	if err != nil {
		t.Fatalf("couldn't query machines: %s", err.Error())
	}

	util.CheckPublicConnections(t, machines, containers, connections)
}
