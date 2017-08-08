package command

import (
	"flag"
	"os"
	"os/exec"
	"path"

	"github.com/quilt/quilt/util"
)

// Init represents an Init command.
type Init struct{}

var initCommands = `quilt init`

var initExplanation = `Create a new infrastructure to use with
baseInfrastructure().

After creating an infrastructure named 'infra', the infrastructure can be used
in any blueprint by calling baseInfrastructure(quilt, 'infra').`

// InstallFlags sets up parsing for command line flags.
func (iCmd *Init) InstallFlags(flags *flag.FlagSet) {
	flags.Usage = func() {
		util.PrintUsageString(initCommands, initExplanation, flags)
	}
}

// Parse parses the command line arguments for the stop command.
func (iCmd *Init) Parse(args []string) error {
	return nil
}

// BeforeRun makes any necessary post-parsing transformations.
func (iCmd *Init) BeforeRun() error {
	return nil
}

// AfterRun performs any necessary post-run cleanup.
func (iCmd *Init) AfterRun() error {
	return nil
}

// Run executes the Nodejs initializer that prompts the user.
func (iCmd *Init) Run() int {
	// This only works when Quilt is installed with `go get`, and `quilt init` is
	// executed from the Quilt root.
	initializerPath := path.Join("quiltctl", "command", "init", "initializer.js")
	cmd := exec.Command("node", initializerPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return 1
	}

	return 0
}
