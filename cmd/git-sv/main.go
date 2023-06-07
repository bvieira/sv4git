package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/bvieira/sv4git/v2/sv"
	"github.com/urfave/cli/v2"
)

// Version for git-sv.
var Version = "source"

const (
	configFilename     = "config.yml"
	repoConfigFilename = ".sv4git.yml"
	configDir          = ".sv4git"
)

var (
	//go:embed resources/templates/*.tpl
	defaultTemplatesFS embed.FS
)

func templateFS(filepath string) fs.FS {
	if _, err := os.Stat(filepath); err != nil {
		defaultTemplatesFS, _ := fs.Sub(defaultTemplatesFS, "resources/templates")
		return defaultTemplatesFS
	}
	return os.DirFS(filepath)
}

func main() {
	log.SetFlags(0)

	repoPath, rerr := getRepoPath()
	if rerr != nil {
		log.Fatal("failed to discovery repository top level, error: ", rerr)
	}

	cfg := loadCfg(repoPath)
	messageProcessor := sv.NewMessageProcessor(cfg.CommitMessage, cfg.Branches)
	git := sv.NewGit(messageProcessor, cfg.Tag)
	semverProcessor := sv.NewSemVerCommitsProcessor(cfg.Versioning, cfg.CommitMessage)
	releasenotesProcessor := sv.NewReleaseNoteProcessor(cfg.ReleaseNotes)
	outputFormatter := sv.NewOutputFormatter(templateFS(filepath.Join(repoPath, configDir, "templates")))

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
			Action:      commitLogHandler(git),
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
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "t", Aliases: []string{"tag"}, Usage: "get release note from tag"},
				&cli.StringFlag{Name: "p", Aliases: []string{"previous"}, Usage: "previous tag to use"},
			},
		},
		{
			Name:    "changelog",
			Aliases: []string{"cgl"},
			Usage:   "generate changelog",
			Action:  changelogHandler(git, semverProcessor, releasenotesProcessor, outputFormatter),
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "size", Value: 10, Aliases: []string{"n"}, Usage: "get changelog from last 'n' tags"},
				&cli.BoolFlag{Name: "all", Usage: "ignore size parameter, get changelog for every tag"},
				&cli.BoolFlag{Name: "add-next-version", Usage: "add next version on change log (commits since last tag, but only if there is a new version to release)"},
				&cli.BoolFlag{Name: "semantic-version-only", Usage: "only show tags 'SemVer-ish'"},
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
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "no-scope", Aliases: []string{"nsc"}, Usage: "do not prompt for commit scope"},
				&cli.BoolFlag{Name: "no-body", Aliases: []string{"nbd"}, Usage: "do not prompt for commit body"},
				&cli.BoolFlag{Name: "no-issue", Aliases: []string{"nis"}, Usage: "do not prompt for commit issue, will try to recover from branch if enabled"},
				&cli.BoolFlag{Name: "no-breaking", Aliases: []string{"nbc"}, Usage: "do not prompt for breaking changes"},
				&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Usage: "define commit type"},
				&cli.StringFlag{Name: "scope", Aliases: []string{"s"}, Usage: "define commit scope"},
				&cli.StringFlag{Name: "description", Aliases: []string{"d"}, Usage: "define commit description"},
				&cli.StringFlag{Name: "breaking-change", Aliases: []string{"b"}, Usage: "define commit breaking change message"},
			},
		},
		{
			Name:    "validate-commit-message",
			Aliases: []string{"vcm"},
			Usage:   "use as prepare-commit-message hook to validate and enhance commit message",
			Action:  validateCommitMessageHandler(git, messageProcessor),
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "path", Required: true, Usage: "git working directory"},
				&cli.StringFlag{Name: "file", Required: true, Usage: "name of the file that contains the commit log message"},
				&cli.StringFlag{Name: "source", Required: true, Usage: "source of the commit message"},
			},
		},
	}

	if apperr := app.Run(os.Args); apperr != nil {
		log.Fatal("ERROR: ", apperr)
	}
}

func loadCfg(repoPath string) Config {
	cfg := defaultConfig()

	envCfg := loadEnvConfig()
	if envCfg.Home != "" {
		homeCfgFilepath := filepath.Join(envCfg.Home, configFilename)
		if homeCfg, err := readConfig(homeCfgFilepath); err == nil {
			if merr := merge(&cfg, migrateConfig(homeCfg, homeCfgFilepath)); merr != nil {
				log.Fatal("failed to merge user config, error: ", merr)
			}
		}
	}

	repoCfgFilepath := filepath.Join(repoPath, repoConfigFilename)
	if repoCfg, err := readConfig(repoCfgFilepath); err == nil {
		if merr := merge(&cfg, migrateConfig(repoCfg, repoCfgFilepath)); merr != nil {
			log.Fatal("failed to merge repo config, error: ", merr)
		}
		if len(repoCfg.ReleaseNotes.Headers) > 0 { // mergo is merging maps, headers will be overwritten
			cfg.ReleaseNotes.Headers = repoCfg.ReleaseNotes.Headers
		}
	}

	return cfg
}
