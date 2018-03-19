package command

import (
	"flag"
	"os"
	"os/exec"

	"github.com/kelda/kelda/util"

	log "github.com/sirupsen/logrus"
)

// BaseInfra represents a BaseInfra command.
type BaseInfra struct{}

var baseInfraCommands = `kelda base-infrastructure`

var baseInfraExplanation = `Create a new base infrastructure. The infrastructure can be
used in blueprints by calling 'baseInfrastructure()'.

A base infrastructure is an easy way to reuse an infrastructure across blueprints.
The command prompts the user for information about their desired infrastructure
and then creates an infrastructure based on the answers.

To edit the infrastructure after creation, either rerun 'kelda base-infrastructure'
or directly edit the infrastructure blueprint stored in '~/.kelda/infra/default.js'.`

// InstallFlags sets up parsing for command line flags.
func (biCmd *BaseInfra) InstallFlags(flags *flag.FlagSet) {
	flags.Usage = func() {
		util.PrintUsageString(baseInfraCommands, baseInfraExplanation, nil)
	}
}

// Parse parses the command line arguments for the stop command.
func (biCmd *BaseInfra) Parse(args []string) error {
	return nil
}

// BeforeRun makes any necessary post-parsing transformations.
func (biCmd *BaseInfra) BeforeRun() error {
	return nil
}

// AfterRun performs any necessary post-run cleanup.
func (biCmd *BaseInfra) AfterRun() error {
	return nil
}

// Run executes the Nodejs initializer that prompts the user.
func (biCmd *BaseInfra) Run() int {
	// Assumes `js/base-infrastructure/intializer.js` was installed in the path
	// somewhere as `kelda-base-infra-init.js`. This is done automatically by
	// npm for users who install Kelda with npm.
	executable := "kelda-base-infra-init.js"
	executablePath, err := exec.LookPath(executable)
	if err != nil {
		log.Errorf("%s: Make sure that "+
			"js/base-infrastructure/intializer.js is installed in your "+
			"$PATH as %s. This is done automatically with "+
			"`npm install -g @kelda/install`, but if you're running Kelda "+
			"from source, you must set up the symlink manually. You can "+
			"do this by executing `ln -s <ABS_PATH_TO_KELDA_SOURCE>/"+
			"js/base-infrastructure/initializer.js /usr/local/bin/%s`",
			err, executable, executable)
		return 1
	}

	nodeBinary, err := util.GetNodeBinary()
	if err != nil {
		log.Error(err)
		return 1
	}

	cmd := exec.Command(nodeBinary, executablePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return 1
	}

	return 0
}
