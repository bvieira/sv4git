package main

import (
	"fmt"
	"log"
	"os"
	"sv4git/sv"

	"github.com/urfave/cli/v2"
)

// Version for git-sv
var Version = ""

func main() {
	log.SetFlags(0)

	cfg := loadConfig()

	fmt.Printf("%+v\n", cfg)

	git := sv.NewGit(cfg.BreakingChangePrefixes, cfg.IssueIDPrefixes, cfg.TagPattern)
	semverProcessor := sv.NewSemVerCommitsProcessor(cfg.IncludeUnknownTypeAsPatch, cfg.MajorVersionTypes, cfg.MinorVersionTypes, cfg.PatchVersionTypes)
	releasenotesProcessor := sv.NewReleaseNoteProcessor(cfg.ReleaseNotesTags)
	outputFormatter := sv.NewOutputFormatter()
	validateMessageProcessor := sv.NewValidateMessageProcessor(cfg.ValidateMessageSkipBranches, cfg.CommitMessageTypes, cfg.IssueKeyName, cfg.BranchIssueRegex, cfg.IssueRegex)

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
			Flags:   []cli.Flag{&cli.StringFlag{Name: "t", Aliases: []string{"tag"}, Usage: "get commit log from tag"}},
		},
		{
			Name:    "release-notes",
			Aliases: []string{"rn"},
			Usage:   "generate release notes",
			Action:  releaseNotesHandler(git, semverProcessor, releasenotesProcessor, outputFormatter),
			Flags:   []cli.Flag{&cli.StringFlag{Name: "t", Aliases: []string{"tag"}, Usage: "get release note from tag"}},
		},
		{
			Name:    "changelog",
			Aliases: []string{"cgl"},
			Usage:   "generate changelog",
			Action:  changelogHandler(git, semverProcessor, releasenotesProcessor, outputFormatter),
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "size", Value: 10, Aliases: []string{"n"}, Usage: "get changelog from last 'n' tags"},
				&cli.BoolFlag{Name: "all", Usage: "ignore size parameter, get changelog for every tag"},
			},
		},
		{
			Name:    "tag",
			Aliases: []string{"tg"},
			Usage:   "generate tag with version based on git commit messages",
			Action:  tagHandler(git, semverProcessor),
		},
		{
			Name:    "commit",
			Aliases: []string{"cmt"},
			Usage:   "execute git commit with convetional commit message helper",
			Action:  commitHandler(cfg, git, validateMessageProcessor),
		},
		{
			Name:    "validate-commit-message",
			Aliases: []string{"vcm"},
			Usage:   "use as prepare-commit-message hook to validate message",
			Action:  validateCommitMessageHandler(git, validateMessageProcessor),
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "path", Required: true, Usage: "git working directory"},
				&cli.StringFlag{Name: "file", Required: true, Usage: "name of the file that contains the commit log message"},
				&cli.StringFlag{Name: "source", Required: true, Usage: "source of the commit message"},
			},
		},
	}

	apperr := app.Run(os.Args)
	if apperr != nil {
		log.Fatal(apperr)
	}
}
