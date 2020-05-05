package cmd

import (
	"sync"

	"github.com/guiyomh/gitlab-wizard/pkg/flagset"
	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	"github.com/xanzy/go-gitlab"
)

type BaseCommand struct {
	UI cli.Ui

	flags     *flagset.FlagSets
	flagsOnce sync.Once

	flagURL   string
	flagToken string
}

func (c *BaseCommand) flagSet() *flagset.FlagSets {
	c.flagsOnce.Do(func() {
		set := flagset.NewFlagSets(c.UI)

		f := set.NewFlagSet("Gilab Options")

		f.StringVar(&flagset.StringVar{
			Name:       "url",
			Target:     &c.flagURL,
			EnvVar:     "CI_API_V4_URL",
			Completion: complete.PredictAnything,
			Usage:      "Base URL of the gitlab API",
			Default:    "https://gitlab.com/api/v4",
		})

		f.StringVar(&flagset.StringVar{
			Name:   "token",
			Target: &c.flagToken,
			EnvVar: "CI_BUILD_TOKEN",
			Usage:  "Gitlab TOKEN",
		})

		c.flags = set
	})
	return c.flags
}

func (c *BaseCommand) Client() (*gitlab.Client, error) {
	client, err := gitlab.NewClient(
		c.flagToken,
		gitlab.WithBaseURL(c.flagURL),
	)
	return client, err
}
