package main

import (
	"log"
	"os"
	"sv4git/sv"

	"github.com/urfave/cli/v2"
)

// Version for git-sv
var Version = ""

func main() {
	cfg := loadConfig()

	git := sv.NewGit(cfg.BreakingChangePrefixes, cfg.IssueIDPrefixes, cfg.TagPattern)
	semverProcessor := sv.NewSemVerCommitsProcessor(cfg.IncludeUnknownTypeAsPatch, cfg.MajorVersionTypes, cfg.MinorVersionTypes, cfg.PatchVersionTypes)
	releasenotesProcessor := sv.NewReleaseNoteProcessor(cfg.ReleaseNotesTags)

	app := cli.NewApp()
	app.Name = "sv"
	app.Version = Version
	app.Usage = "semantic version for git"
	app.Commands = []*cli.Command{
		{
			Name:    "current-version",
			Aliases: []string{"cv"},
			Usage:   "get last released version from git",
			Action:  currentVersionHandler(git),
		},
		{
			Name:    "next-version",
			Aliases: []string{"nv"},
			Usage:   "generate the next version based on git commit messages",
			Action:  nextVersionHandler(git, semverProcessor),
		},
		{
			Name:    "commit-log",
			Aliases: []string{"cl"},
			Usage:   "list all commit logs since last version as jsons",
			Action:  commitLogHandler(git, semverProcessor),
			Flags:   []cli.Flag{&cli.StringFlag{Name: "t", Usage: "get commit log from tag"}},
		},
		{
			Name:    "release-notes",
			Aliases: []string{"rn"},
			Usage:   "generate release notes",
			Action:  releaseNotesHandler(git, semverProcessor, releasenotesProcessor),
			Flags:   []cli.Flag{&cli.StringFlag{Name: "t", Usage: "get release note from tag"}},
		},
		{
			Name:    "tag",
			Aliases: []string{"tg"},
			Usage:   "generate tag with version based on git commit messages",
			Action:  tagHandler(git, semverProcessor, releasenotesProcessor),
		},
	}

	apperr := app.Run(os.Args)
	if apperr != nil {
		log.Fatal(apperr)
	}
}
