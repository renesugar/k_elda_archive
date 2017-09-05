// +build !windows

package cli

import "github.com/quilt/quilt/cli/command"

func init() {
	commands["minion"] = command.NewMinionCommand()
}
