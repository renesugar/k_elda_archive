package command

import (
	"flag"
	"os"
	"os/exec"

	"github.com/kelda/kelda/util"

	log "github.com/sirupsen/logrus"
)

// ConfigProvider represents a ConfigProvider command.
type ConfigProvider struct{}

var configProviderCommands = `kelda configure-provider`

var configProviderExplanation = `Set up cloud provider credentials.

This command helps ensure that cloud provider credentials are formatted correctly
and placed in the right location for Kelda to access. Kelda needs access to cloud
provider credentials in order to boot virtual machines in your account.`

// InstallFlags sets up parsing for command line flags.
func (cpCmd *ConfigProvider) InstallFlags(flags *flag.FlagSet) {
	flags.Usage = func() {
		util.PrintUsageString(
			configProviderCommands, configProviderExplanation, nil)
	}
}

// Parse parses the command line arguments for the stop command.
func (cpCmd *ConfigProvider) Parse(args []string) error {
	return nil
}

// BeforeRun makes any necessary post-parsing transformations.
func (cpCmd *ConfigProvider) BeforeRun() error {
	return nil
}

// AfterRun performs any necessary post-run cleanup.
func (cpCmd *ConfigProvider) AfterRun() error {
	return nil
}

// Run executes the Nodejs initializer that prompts the user.
func (cpCmd *ConfigProvider) Run() int {
	// Assumes `js/configure-provider/intializer.js` was installed in the path
	// somewhere as `kelda-configure-provider.js`. This is done automatically
	// by npm for users who installed Kelda using npm.
	executable := "kelda-configure-provider.js"
	executablePath, err := exec.LookPath(executable)
	if err != nil {
		log.Errorf("%s: Make sure that "+
			"js/configure-provider/initializer.js is installed in your "+
			"$PATH as %s. This is done automatically with "+
			"`npm install -g @kelda/install`, but if you're running Kelda "+
			"from source, you must set up the symlink manually. You can "+
			"do this by executing `ln -s <ABS_PATH_TO_KELDA_SOURCE>/"+
			"js/configure-provider/initializer.js /usr/local/bin/%s`",
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
