package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/guiyomh/gitlab-wizard/pkg/flagset"
	"github.com/xanzy/go-gitlab"
)

type ArtifactDownloadCommand struct {
	*BaseCommand

	flagProject       string
	flagJob           string
	flagCommit        string
	flagExtract       bool
	flagForceDownload bool
}

func (c *ArtifactDownloadCommand) Synopsis() string {
	return "Downloading a gitlab artifact"
}

func (c *ArtifactDownloadCommand) Help() string {
	helpText := `
Usage: gitlab-wizard artifact download [option] [args]

	This command downloads an artifact from a successful job for a given commit ID

` + c.Flags().Help()

	return strings.TrimSpace(helpText)
}

func (c *ArtifactDownloadCommand) Flags() *flagset.FlagSets {
	set := c.flagSet()
	f := set.NewFlagSet("Downloading Options")

	f.StringVar(&flagset.StringVar{
		Name:   "project",
		Target: &c.flagProject,
		Usage:  "Project ID or project path (eg: apps/my-project).",
		EnvVar: "CI_PROJECT_ID",
	})

	f.StringVar(&flagset.StringVar{
		Name:   "job",
		Target: &c.flagJob,
		Usage:  "Name of the job that produce a artifact.",
	})

	f.StringVar(&flagset.StringVar{
		Name:   "commit",
		Target: &c.flagCommit,
		Usage:  "SHA1 of the commit that produce the artifact.",
		EnvVar: "CI_BUILD_REF",
	})

	f.BoolVar(&flagset.BoolVar{
		Name:    "force",
		Target:  &c.flagForceDownload,
		Usage:   "if true, override the existing file",
		Default: false,
	})

	f.BoolVar(&flagset.BoolVar{
		Name:    "extract",
		Target:  &c.flagExtract,
		Usage:   "if true, extract the downloaded artefact in current path",
		Default: false,
	})

	return set
}

func (c *ArtifactDownloadCommand) Run(args []string) int {
	f := c.Flags()

	if err := f.Parse(args); err != nil {
		c.UI.Error(fmt.Sprintf("ERROR: %s", err))
		return 1
	}

	git, err := c.Client()
	if err != nil {
		c.UI.Error(fmt.Sprintf("ERROR: %s", err))
		return 1
	}
	job, err := c.findJob(git)
	if err != nil {
		c.UI.Error(fmt.Sprintf("ERROR: %s", err))
		return 1
	}
	c.UI.Info(fmt.Sprintf("Job found (ID=%d startedAt=%s)\n", job.ID, job.StartedAt.Format(time.RFC822Z)))
	z, err := c.downloadArtifact(git, job)
	if err != nil {
		c.UI.Error(fmt.Sprintf("ERROR: %s", err))
		return 1
	}
	if !c.flagExtract {
		c.UI.Info(fmt.Sprintf("Artifact downloaded as %s", z))
	} else {
		err = c.unzipArtifact(z)
		if err != nil {
			c.UI.Error(fmt.Sprintf("ERROR: %s", err))
		}
	}

	return 0
}

func (c *ArtifactDownloadCommand) findJob(git *gitlab.Client) (*gitlab.Job, error) {

	opt := &gitlab.ListJobsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		Scope: []gitlab.BuildStateValue{
			gitlab.Success,
		},
	}
	for {
		jobs, resp, err := git.Jobs.ListProjectJobs(c.flagProject, opt)
		if err != nil {
			return nil, err
		}

		for _, job := range jobs {
			if job.Commit.ID != c.flagCommit {
				continue
			}
			if job.Name != c.flagJob {
				continue
			}
			return &job, nil
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil, fmt.Errorf("Couldn't found the job %s in success for commit %s", c.flagJob, c.flagCommit)
}

func (c *ArtifactDownloadCommand) downloadArtifact(git *gitlab.Client, job *gitlab.Job) (string, error) {
	zip := fmt.Sprintf("%s.zip", c.flagCommit)

	if _, err := os.Stat(zip); !os.IsNotExist(err) && !c.flagForceDownload {
		return zip, fmt.Errorf("artifact file is already downloaded as %s", zip)
	}
	reader, _, err := git.Jobs.GetJobArtifacts(c.flagProject, job.ID)
	if err != nil {
		return zip, err
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return zip, err
	}
	err = ioutil.WriteFile(zip, b, 0644)
	if err != nil {
		return zip, err
	}
	return zip, nil
}

func (c *ArtifactDownloadCommand) unzipArtifact(src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	for _, f := range r.File {
		err := c.extractAndWrite(f)
		if err != nil {
			return err
		}
	}
	os.Remove(src)
	return nil
}

func (c *ArtifactDownloadCommand) extractAndWrite(f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); err != nil {
			panic(err)
		}
	}()
	c.UI.Info(fmt.Sprintf("\textract: %s", f.Name))
	if f.FileInfo().IsDir() {
		if err = os.MkdirAll(f.Name, f.Mode()); err != nil {
			return err
		}
	} else {
		if err = os.MkdirAll(filepath.Dir(f.Name), f.Mode()); err != nil {
			return err
		}
		f, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}
