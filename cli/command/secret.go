package command

import (
	"errors"
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/kelda/kelda/util"
)

// Secret defines the options for the Secret command.
type Secret struct {
	name, value string
	connectionHelper
}

var secretCommands = "kelda secret NAME VALUE"
var secretExplanation = `Securely set a secret association. This command must
be run before any containers referencing the associated secret can be started.`

// InstallFlags sets up parsing for command line flags.
func (secretCmd *Secret) InstallFlags(flags *flag.FlagSet) {
	secretCmd.connectionHelper.InstallFlags(flags)
	flags.Usage = func() {
		util.PrintUsageString(secretCommands, secretExplanation, flags)
	}
}

// Parse parses the command line arguments for the secret command.
func (secretCmd *Secret) Parse(args []string) error {
	if len(args) != 2 {
		return errors.New("a name and value must be supplied")
	}

	secretCmd.name = args[0]
	secretCmd.value = args[1]
	return nil
}

// Run implements the secret command.
func (secretCmd Secret) Run() int {
	err := secretCmd.client.SetSecret(secretCmd.name, secretCmd.value)
	if err != nil {
		log.WithError(err).Error("Failed to set secret")
		return 1
	}
	return 0
}
