package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/bvieira/sv4git/v2/sv"
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

// EnvConfig env vars for cli configuration.
type EnvConfig struct {
	Home string `envconfig:"SV4GIT_HOME" default:""`
}

func loadEnvConfig() EnvConfig {
	var c EnvConfig
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal("failed to load env config, error: ", err.Error())
	}
	return c
}

// Config cli yaml config.
type Config struct {
	Version       string                 `yaml:"version"`
	Versioning    sv.VersioningConfig    `yaml:"versioning"`
	Tag           sv.TagConfig           `yaml:"tag"`
	ReleaseNotes  sv.ReleaseNotesConfig  `yaml:"release-notes"`
	Branches      sv.BranchesConfig      `yaml:"branches"`
	CommitMessage sv.CommitMessageConfig `yaml:"commit-message"`
}

func getRepoPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", combinedOutputErr(err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func combinedOutputErr(err error, out []byte) error {
	msg := strings.Split(string(out), "\n")
	return fmt.Errorf("%v - %s", err, msg[0])
}

func readConfig(filepath string) (Config, error) {
	content, rerr := os.ReadFile(filepath)
	if rerr != nil {
		return Config{}, rerr
	}

	var cfg Config
	cerr := yaml.Unmarshal(content, &cfg)
	if cerr != nil {
		return Config{}, fmt.Errorf("could not parse config from path: %s, error: %v", filepath, cerr)
	}

	return cfg, nil
}

func defaultConfig() Config {
	skipDetached := false
	pattern := "%d.%d.%d"
	filter := ""
	return Config{
		Version: "1.1",
		Versioning: sv.VersioningConfig{
			UpdateMajor:   []string{},
			UpdateMinor:   []string{"feat"},
			UpdatePatch:   []string{"build", "ci", "chore", "docs", "fix", "perf", "refactor", "style", "test"},
			IgnoreUnknown: false,
		},
		Tag: sv.TagConfig{
			Pattern: &pattern,
			Filter:  &filter,
		},
		ReleaseNotes: sv.ReleaseNotesConfig{
			Sections: []sv.ReleaseNotesSectionConfig{
				{Name: "Features", SectionType: sv.ReleaseNotesSectionTypeCommits, CommitTypes: []string{"feat"}},
				{Name: "Bug Fixes", SectionType: sv.ReleaseNotesSectionTypeCommits, CommitTypes: []string{"fix"}},
				{Name: "Breaking Changes", SectionType: sv.ReleaseNotesSectionTypeBreakingChanges},
			},
		},
		Branches: sv.BranchesConfig{
			Prefix:       "([a-z]+\\/)?",
			Suffix:       "(-.*)?",
			DisableIssue: false,
			Skip:         []string{"master", "main", "developer"},
			SkipDetached: &skipDetached,
		},
		CommitMessage: sv.CommitMessageConfig{
			Types: []string{"build", "ci", "chore", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"},
			Scope: sv.CommitMessageScopeConfig{},
			Footer: map[string]sv.CommitMessageFooterConfig{
				"issue": {Key: "jira", KeySynonyms: []string{"Jira", "JIRA"}},
			},
			Issue:          sv.CommitMessageIssueConfig{Regex: "[A-Z]+-[0-9]+"},
			HeaderSelector: "",
		},
	}
}

func merge(dst *Config, src Config) error {
	err := mergo.Merge(dst, src, mergo.WithOverride, mergo.WithTransformers(&mergeTransformer{}))
	if err == nil {
		if len(src.ReleaseNotes.Headers) > 0 { // mergo is merging maps, ReleaseNotes.Headers should be overwritten
			dst.ReleaseNotes.Headers = src.ReleaseNotes.Headers
		}
	}
	return err
}

type mergeTransformer struct{}

func (t *mergeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Slice {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() && !src.IsNil() {
				dst.Set(src)
			}
			return nil
		}
	}

	if typ.Kind() == reflect.Ptr {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() && !src.IsNil() {
				dst.Set(src)
			}
			return nil
		}
	}
	return nil
}

func migrateConfig(cfg Config, filename string) Config {
	if cfg.ReleaseNotes.Headers == nil {
		return cfg
	}
	warnf("config 'release-notes.headers' on %s is deprecated, please use 'sections' instead!", filename)

	return Config{
		Version:    cfg.Version,
		Versioning: cfg.Versioning,
		Tag:        cfg.Tag,
		ReleaseNotes: sv.ReleaseNotesConfig{
			Sections: migrateReleaseNotesConfig(cfg.ReleaseNotes.Headers),
		},
		Branches:      cfg.Branches,
		CommitMessage: cfg.CommitMessage,
	}
}

func migrateReleaseNotesConfig(headers map[string]string) []sv.ReleaseNotesSectionConfig {
	order := []string{"feat", "fix", "refactor", "perf", "test", "build", "ci", "chore", "docs", "style"}
	var sections []sv.ReleaseNotesSectionConfig
	for _, key := range order {
		if name, exists := headers[key]; exists {
			sections = append(sections, sv.ReleaseNotesSectionConfig{Name: name, SectionType: sv.ReleaseNotesSectionTypeCommits, CommitTypes: []string{key}})
		}
	}
	if name, exists := headers["breaking-change"]; exists {
		sections = append(sections, sv.ReleaseNotesSectionConfig{Name: name, SectionType: sv.ReleaseNotesSectionTypeBreakingChanges})
	}
	return sections
}
