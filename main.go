package main

import (
	"fmt"
	"log"
	"os"

	"github.com/guiyomh/gitlab-wizard/cmd"
	"github.com/mitchellh/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	ui := &cli.BasicUi{
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	baseCmd := &cmd.BaseCommand{
		UI: ui,
	}

	c := cli.NewCLI("gitlab-wizard", fmt.Sprintf("%s - %s by %s at %s", version, commit, builtBy, date))
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"artifact ": func() (cli.Command, error) {
			return &cmd.ArtifactCommand{}, nil
		},
		"artifact download": func() (cli.Command, error) {
			return &cmd.ArtifactDownloadCommand{
				BaseCommand: baseCmd,
			}, nil
		},
	}
	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
