package tests

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

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

// TestExampleBlueprints tests that all blueprints in the examples directory compile.
func TestExampleBlueprints() error {
	// Use absolute rather than relative paths, so that the Chdir command below works
	// regardless of the starting directory.
	absolutePath, err := filepath.Abs("../../examples/*/*.js")
	if err != nil {
		return err
	}

	blueprints, err := filepath.Glob(absolutePath)
	if err != nil {
		return err
	}

	for _, blueprint := range blueprints {
		log.Infof("Testing %s", blueprint)

		// Change into the directory of the blueprint in order to install
		// dependencies.
		os.Chdir(filepath.Dir(blueprint))
		if err = run("npm", "install", "."); err != nil {
			return err
		}

		if err = tryRunBlueprint(blueprint); err != nil {
			return err
		}
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

	blueprints, err := filepath.Glob("./tests/*/*.js")
	if err != nil {
		return err
	}

	for _, blueprint := range blueprints {
		log.Infof("Testing %s", blueprint)
		if err := tryRunBlueprint("./" + blueprint); err != nil {
			return err
		}
	}
	return nil
}
