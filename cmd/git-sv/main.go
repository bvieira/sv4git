package main

import (
	"log"
	"os"
	"path/filepath"
	"sv4git/sv"

	"github.com/imdario/mergo"
	"github.com/urfave/cli/v2"
)

// Version for git-sv
var Version = ""

const (
	configFilename     = "config.yml"
	repoConfigFilename = ".sv4git.yml"
)

func main() {
	log.SetFlags(0)

	envCfg := loadEnvConfig()

	cfg := defaultConfig()

	if envCfg.Home != "" {
		if homeCfg, err := loadConfig(filepath.Join(envCfg.Home, configFilename)); err == nil {
			if merr := mergo.Merge(&cfg, homeCfg, mergo.WithOverride); merr != nil {
				log.Fatal(merr)
			}
		}
	}

	repoPath, rerr := getRepoPath()
	if rerr != nil {
		log.Fatal(rerr)
	}

	if repoCfg, err := loadConfig(filepath.Join(repoPath, repoConfigFilename)); err == nil {
		if merr := mergo.Merge(&cfg, repoCfg, mergo.WithOverride); merr != nil {
			log.Fatal(merr)
		}
		if len(repoCfg.ReleaseNotes.Headers) > 0 { // mergo is merging maps, headers will be overwritten
			cfg.ReleaseNotes.Headers = repoCfg.ReleaseNotes.Headers
		}
	}

	messageProcessor := sv.NewMessageProcessor(cfg.CommitMessage, cfg.Branches)
	git := sv.NewGit(messageProcessor, cfg.Tag)
	semverProcessor := sv.NewSemVerCommitsProcessor(cfg.Versioning, cfg.CommitMessage)
	releasenotesProcessor := sv.NewReleaseNoteProcessor(cfg.ReleaseNotes)
	outputFormatter := sv.NewOutputFormatter()

	app := cli.NewApp()
	app.Name = "sv"
	app.Version = Version
	app.Usage = "semantic version for git"
	app.Commands = []*cli.Command{
		{
			Name:    "config",
			Aliases: []string{"cfg"},
			Usage:   "cli configuration",
			Subcommands: []*cli.Command{
				{
					Name:   "default",
					Usage:  "show default config",
					Action: configDefaultHandler(),
				},
				{
					Name:   "show",
					Usage:  "show current config",
					Action: configShowHandler(cfg),
				},
			},
		},
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
			Name:        "commit-log",
			Aliases:     []string{"cl"},
			Usage:       "list all commit logs according to range as jsons",
			Description: "The range filter is used based on git log filters, check https://git-scm.com/docs/git-log for more info. When flag range is \"tag\" and start is empty, last tag created will be used instead. When flag range is \"date\", if \"end\" is YYYY-MM-DD the range will be inclusive.",
			Action:      commitLogHandler(git, semverProcessor),
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "t", Aliases: []string{"tag"}, Usage: "get commit log from a specific tag"},
				&cli.StringFlag{Name: "r", Aliases: []string{"range"}, Usage: "type of range of commits, use: tag, date or hash", Value: string(sv.TagRange)},
				&cli.StringFlag{Name: "s", Aliases: []string{"start"}, Usage: "start range of git log revision range, if date, the value is used on since flag instead"},
				&cli.StringFlag{Name: "e", Aliases: []string{"end"}, Usage: "end range of git log revision range, if date, the value is used on until flag instead"},
			},
		},
		{
			Name:        "commit-notes",
			Aliases:     []string{"cn"},
			Usage:       "generate a commit notes according to range",
			Description: "The range filter is used based on git log filters, check https://git-scm.com/docs/git-log for more info. When flag range is \"tag\" and start is empty, last tag created will be used instead. When flag range is \"date\", if \"end\" is YYYY-MM-DD the range will be inclusive.",
			Action:      commitNotesHandler(git, releasenotesProcessor, outputFormatter),
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "r", Aliases: []string{"range"}, Usage: "type of range of commits, use: tag, date or hash", Required: true},
				&cli.StringFlag{Name: "s", Aliases: []string{"start"}, Usage: "start range of git log revision range, if date, the value is used on since flag instead"},
				&cli.StringFlag{Name: "e", Aliases: []string{"end"}, Usage: "end range of git log revision range, if date, the value is used on until flag instead"},
			},
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
			Action:  commitHandler(cfg, git, messageProcessor),
		},
		{
			Name:    "validate-commit-message",
			Aliases: []string{"vcm"},
			Usage:   "use as prepare-commit-message hook to validate message",
			Action:  validateCommitMessageHandler(git, messageProcessor),
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
