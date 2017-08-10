package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/join"
)

type connectionTester struct {
	labelMap      map[string]db.Label
	connectionMap map[string][]string
	allIPs        []string
}

func newConnectionTester(clnt client.Client) (connectionTester, error) {
	labels, err := clnt.QueryLabels()
	if err != nil {
		return connectionTester{}, err
	}

	allIPsSet := make(map[string]struct{})
	labelMap := make(map[string]db.Label)
	for _, label := range labels {
		labelMap[label.Label] = label

		for _, ip := range append(label.ContainerIPs, label.IP) {
			allIPsSet[ip] = struct{}{}
		}
	}

	var allIPs []string
	for ip := range allIPsSet {
		allIPs = append(allIPs, ip)
	}

	connections, err := clnt.QueryConnections()
	if err != nil {
		return connectionTester{}, err
	}

	connectionMap := make(map[string][]string)
	for _, conn := range connections {
		connectionMap[conn.From] = append(connectionMap[conn.From], conn.To)
		// Connections are bi-directional.
		connectionMap[conn.To] = append(connectionMap[conn.To], conn.From)
	}

	return connectionTester{
		labelMap:      labelMap,
		connectionMap: connectionMap,
		allIPs:        allIPs,
	}, nil
}

type pingResult struct {
	target    string
	reachable bool
	err       error
	cmdTime   commandTime
}

// We have to limit our parallelization because each `quilt exec` creates a new SSH login
// session. Doing this quickly in parallel breaks system-logind
// on the remote machine: https://github.com/systemd/systemd/issues/2925.
// Furthermore, the concurrency limit cannot exceed the sshd MaxStartups setting,
// or else the SSH connections may be randomly rejected.
const execConcurrencyLimit = 5

func (tester connectionTester) pingAll(container db.Container) []pingResult {
	pingResultsChan := make(chan pingResult, len(tester.allIPs))

	// Create worker threads.
	pingRequests := make(chan string, execConcurrencyLimit)
	var wg sync.WaitGroup
	for i := 0; i < execConcurrencyLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range pingRequests {
				startTime := time.Now()
				_, err := ping(container.StitchID, ip)
				pingResultsChan <- pingResult{
					target:    ip,
					reachable: err == nil,
					err:       err,
					cmdTime:   commandTime{startTime, time.Now()},
				}
			}
		}()
	}

	// Feed worker threads.
	for _, ip := range tester.allIPs {
		pingRequests <- ip
	}
	close(pingRequests)
	wg.Wait()
	close(pingResultsChan)

	// Collect results.
	var pingResults []pingResult
	for res := range pingResultsChan {
		pingResults = append(pingResults, res)
	}

	return pingResults
}

func (tester connectionTester) test(container db.Container) (failures []error) {
	// We should be able to ping ourselves.
	expReachable := map[string]struct{}{
		container.IP: {},
	}
	for _, label := range container.Labels {
		for _, toLabelName := range tester.connectionMap[label] {
			toLabel := tester.labelMap[toLabelName]
			for _, ip := range append(toLabel.ContainerIPs, toLabel.IP) {
				expReachable[ip] = struct{}{}
			}
		}
	}

	var expPings []pingResult
	for _, ip := range tester.allIPs {
		_, reachable := expReachable[ip]
		expPings = append(expPings, pingResult{
			target:    ip,
			reachable: reachable,
		})
	}
	pingResults := tester.pingAll(container)
	_, _, failuresIntf := join.HashJoin(pingSlice(expPings), pingSlice(pingResults),
		ignoreErrorField, ignoreErrorField)

	for _, badIntf := range failuresIntf {
		bad := badIntf.(pingResult)
		var failure error
		if bad.reachable {
			failure = fmt.Errorf("(%s) could ping unauthorized container %s",
				bad.cmdTime, bad.target)
		} else {
			failure = fmt.Errorf("(%s) couldn't ping authorized container "+
				"%s: %s", bad.cmdTime, bad.target, bad.err)
		}
		failures = append(failures, failure)
	}

	return failures
}

func ignoreErrorField(pingResultIntf interface{}) interface{} {
	return pingResult{
		target:    pingResultIntf.(pingResult).target,
		reachable: pingResultIntf.(pingResult).reachable,
	}
}

// ping `target` from within container `id` with 3 packets, with a timeout of
// 1 second for each packet.
func ping(id string, target string) (string, error) {
	return quiltSSH(id, "ping", "-c", "3", "-W", "1", target)
}

type pingSlice []pingResult

func (ps pingSlice) Get(ii int) interface{} {
	return ps[ii]
}

func (ps pingSlice) Len() int {
	return len(ps)
}
