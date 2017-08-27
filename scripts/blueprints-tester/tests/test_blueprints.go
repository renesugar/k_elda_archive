package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"github.com/quilt/quilt/stitch"
)

func tryRunBlueprint(blueprint string) error {
	_, err := stitch.FromFile(blueprint)
	return err
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	stderr := bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	if cmd.Run() != nil {
		return errors.New(stderr.String())
	}
	return nil
}

// TestCIBlueprints checks that the listed quilt-tester blueprints compile.
func TestCIBlueprints() error {
	// Make the working directory the root of the Quilt repo so that the following
	// relative paths will work.
	os.Chdir("../../quilt-tester")

	if err := run("npm", "install", "."); err != nil {
		return err
	}

	blueprints := []string{
		"./tests/100-logs/logs.js",
		"./tests/61-duplicate-cluster/duplicate-cluster.js",
		"./tests/60-duplicate-cluster-setup/duplicate-cluster-setup.js",
		"./tests/40-stop/stop.js",
		"./tests/30-mean/mean.js",
		"./tests/20-spark/spark.js",
		"./tests/15-bandwidth/bandwidth.js",
		"./tests/10-network/network.js",
		"./tests/outbound-public/outbound-public.js",
		"./tests/inbound-public/inbound-public.js",
		"./tests/elasticsearch/elasticsearch.js",
		"./tests/build-dockerfile/build-dockerfile.js",
		"./tests/etcd/etcd.js",
		"./tests/zookeeper/zookeeper.js",
		"./tests/connection-credentials/connection-credentials.js",
	}

	for _, blueprint := range blueprints {
		log.Infof("Testing %s", blueprint)
		if err := tryRunBlueprint(blueprint); err != nil {
			return err
		}
	}
	return nil
}
