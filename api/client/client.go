package client

//go:generate mockery -name Client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kelda/kelda/api"
	"github.com/kelda/kelda/api/pb"
	"github.com/kelda/kelda/connection"
	"github.com/kelda/kelda/db"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// The timeout for making requests to the daemon once we've connected.
	requestTimeout = time.Minute
)

// Client provides methods to interact with the Kelda daemon.
type Client interface {
	// Close the grpc connection.
	Close() error

	// QueryMachines retrieves the machines tracked by the Kelda daemon.
	QueryMachines() ([]db.Machine, error)

	// QueryContainers retrieves the containers tracked by the Kelda daemon.
	QueryContainers() ([]db.Container, error)

	// QueryEtcd retrieves the etcd information tracked by the Kelda daemon.
	QueryEtcd() ([]db.Etcd, error)

	// QueryConnections retrieves the connection information tracked by the
	// Kelda daemon.
	QueryConnections() ([]db.Connection, error)

	// QueryLoadBalancers retrieves the load balancer information tracked by
	// the Kelda daemon.
	QueryLoadBalancers() ([]db.LoadBalancer, error)

	// QueryBlueprints retrieves blueprint information tracked by the Kelda daemon.
	QueryBlueprints() ([]db.Blueprint, error)

	// QueryCounters retrieves the debugging counters tracked with the Kelda daemon.
	QueryCounters() ([]pb.Counter, error)

	// QueryCounters retrieves the debugging counters tracked by a Kelda minion.
	// Only defined on the daemon.
	QueryMinionCounters(string) ([]pb.Counter, error)

	// QueryImages retrieves the image information tracked by the Kelda daemon.
	QueryImages() ([]db.Image, error)

	// Deploy makes a request to the Kelda daemon to deploy the given deployment.
	// Only defined on the daemon.
	Deploy(deployment string) error

	// Version retrieves the Kelda version of the remote daemon.
	Version() (string, error)
}

// Getter obtains a client connected to the given address.
type Getter func(string, connection.Credentials) (Client, error)

type clientImpl struct {
	pbClient pb.APIClient
	cc       *grpc.ClientConn
}

// New creates a new Kelda client connected to `lAddr`.
func New(lAddr string, creds connection.Credentials) (Client, error) {
	proto, addr, err := api.ParseListenAddress(lAddr)
	if err != nil {
		return nil, err
	}

	cc, err := connection.Client(proto, addr, creds.ClientOpts())
	if err != nil {
		if err == context.DeadlineExceeded {
			err = daemonTimeoutError{
				host:         lAddr,
				connectError: err,
			}
		}
		return nil, err
	}

	pbClient := pb.NewAPIClient(cc)
	return clientImpl{
		pbClient: pbClient,
		cc:       cc,
	}, nil
}

// Writes the result into `v` a pointer to a slice of database structs.  For example
// *[]db.Machine.
func query(pbClient pb.APIClient, table db.TableType, v interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	reply, err := pbClient.Query(ctx, &pb.DBQuery{Table: string(table)})
	if err != nil {
		return err
	}

	replyBytes := []byte(reply.TableContents)
	return json.Unmarshal(replyBytes, v)
}

// Close the grpc connection.
func (c clientImpl) Close() error {
	return c.cc.Close()
}

// QueryMachines retrieves the machines tracked by the Kelda daemon.
func (c clientImpl) QueryMachines() ([]db.Machine, error) {
	var rows []db.Machine
	return rows, query(c.pbClient, db.MachineTable, &rows)
}

// QueryContainers retrieves the containers tracked by the Kelda daemon.
func (c clientImpl) QueryContainers() ([]db.Container, error) {
	var rows []db.Container
	return rows, query(c.pbClient, db.ContainerTable, &rows)
}

// QueryEtcd retrieves the etcd information tracked by the Kelda daemon.
func (c clientImpl) QueryEtcd() ([]db.Etcd, error) {
	var rows []db.Etcd
	return rows, query(c.pbClient, db.EtcdTable, &rows)
}

// QueryConnections retrieves the connection information tracked by the Kelda daemon.
func (c clientImpl) QueryConnections() ([]db.Connection, error) {
	var rows []db.Connection
	return rows, query(c.pbClient, db.ConnectionTable, &rows)
}

// QueryLoadBalancers retrieves the load balancer information tracked by the
// Kelda daemon.
func (c clientImpl) QueryLoadBalancers() ([]db.LoadBalancer, error) {
	var rows []db.LoadBalancer
	return rows, query(c.pbClient, db.LoadBalancerTable, &rows)
}

// QueryBlueprints retrieves the blueprint information tracked by the Kelda daemon.
func (c clientImpl) QueryBlueprints() ([]db.Blueprint, error) {
	var rows []db.Blueprint
	return rows, query(c.pbClient, db.BlueprintTable, &rows)
}

// QueryImages retrieves the image information tracked by the Kelda daemon.
func (c clientImpl) QueryImages() ([]db.Image, error) {
	var rows []db.Image
	return rows, query(c.pbClient, db.ImageTable, &rows)
}

// QueryCounters retrieves the debugging counters tracked with the Kelda daemon.
func (c clientImpl) QueryCounters() ([]pb.Counter, error) {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	reply, err := c.pbClient.QueryCounters(ctx, &pb.CountersRequest{})
	if err != nil {
		return nil, err
	}

	return parseCountersReply(reply), nil
}

// QueryCounters retrieves the debugging counters tracked by a Kelda minion.
func (c clientImpl) QueryMinionCounters(host string) ([]pb.Counter, error) {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	reply, err := c.pbClient.QueryMinionCounters(ctx,
		&pb.MinionCountersRequest{Host: host})
	if err != nil {
		return nil, err
	}

	return parseCountersReply(reply), nil
}

func parseCountersReply(reply *pb.CountersReply) (counters []pb.Counter) {
	for _, c := range reply.Counters {
		counters = append(counters, *c)
	}
	return counters
}

// Deploy makes a request to the Kelda daemon to deploy the given deployment.
func (c clientImpl) Deploy(deployment string) error {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	_, err := c.pbClient.Deploy(ctx, &pb.DeployRequest{Deployment: deployment})
	return err
}

// Version retrieves the Kelda version of the remote daemon.
func (c clientImpl) Version() (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	version, err := c.pbClient.Version(ctx, &pb.VersionRequest{})
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

// daemonTimeoutError represents when we are unable to connect to the Kelda
// daemon because of a timeout.
type daemonTimeoutError struct {
	host         string
	connectError error
}

func (err daemonTimeoutError) Error() string {
	return fmt.Sprintf("Unable to connect to the Kelda daemon at %s: %s. "+
		"Is the kelda daemon running? If not, you can start it with "+
		"`kelda daemon`.", err.host, err.connectError.Error())
}
