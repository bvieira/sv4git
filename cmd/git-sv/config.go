package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config env vars for cli configuration
type Config struct {
	MajorVersionTypes         []string          `envconfig:"MAJOR_VERSION_TYPES" default:""`
	MinorVersionTypes         []string          `envconfig:"MINOR_VERSION_TYPES" default:"feat"`
	PatchVersionTypes         []string          `envconfig:"PATCH_VERSION_TYPES" default:"build,ci,docs,fix,perf,refactor,style,test"`
	IncludeUnknownTypeAsPatch bool              `envconfig:"INCLUDE_UNKNOWN_TYPE_AS_PATCH" default:"true"`
	CommitMessageMetadata     map[string]string `envconfig:"COMMIT_MESSAGE_METADATA" default:"breakingchange:BREAKING CHANGE,issueid:jira"`
	TagPattern                string            `envconfig:"TAG_PATTERN" default:"v%d.%d.%d"`
	ReleaseNotesTags          map[string]string `envconfig:"RELEASE_NOTES_TAGS" default:"fix:Bug Fixes,feat:Features"`
}

func loadConfig() Config {
	var c Config
	err := envconfig.Process("SV", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	return c
}
