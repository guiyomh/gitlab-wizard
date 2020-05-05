package cmd

import (
	"strings"

	"github.com/mitchellh/cli"
)

type ArtifactCommand struct {
}

func (c *ArtifactCommand) Synopsis() string {
	return "Interact with gitlab artifact"
}

func (c *ArtifactCommand) Help() string {
	helpText := `
Usage: gitlab-wizard artifact <subcommand> [option] [args]

	This command groups subcommands for interacting with gitlab artifact

	Please see the individual subcommand help for detailed usage information.
`
	return strings.TrimSpace(helpText)
}

func (c *ArtifactCommand) Run(args []string) int {
	return cli.RunResultHelp
}
