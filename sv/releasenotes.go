package sv

import (
	"time"

	"github.com/Masterminds/semver/v3"
)

// ReleaseNoteProcessor release note processor interface.
type ReleaseNoteProcessor interface {
	Create(version *semver.Version, tag string, date time.Time, commits []GitCommitLog) ReleaseNote
}

// ReleaseNoteProcessorImpl release note based on commit log.
type ReleaseNoteProcessorImpl struct {
	cfg ReleaseNotesConfig
}

// NewReleaseNoteProcessor ReleaseNoteProcessor constructor.
func NewReleaseNoteProcessor(cfg ReleaseNotesConfig) *ReleaseNoteProcessorImpl {
	return &ReleaseNoteProcessorImpl{cfg: cfg}
}

// Create create a release note based on commits.
func (p ReleaseNoteProcessorImpl) Create(version *semver.Version, tag string, date time.Time, commits []GitCommitLog) ReleaseNote {
	sections := make(map[string]ReleaseNoteSection)
	authors := make(map[string]struct{})
	var breakingChanges []string
	for _, commit := range commits {
		authors[commit.AuthorName] = struct{}{}
		if name, exists := p.cfg.Headers[commit.Message.Type]; exists {
			section, sexists := sections[commit.Message.Type]
			if !sexists {
				section = ReleaseNoteSection{Name: name, Types: []string{commit.Message.Type}} //TODO: change to support more than one type per section
			}
			section.Items = append(section.Items, commit)
			sections[commit.Message.Type] = section
		}
		if commit.Message.BreakingMessage() != "" {
			// TODO: if no message found, should use description instead?
			breakingChanges = append(breakingChanges, commit.Message.BreakingMessage())
		}
	}

	var breakingChangeSection BreakingChangeSection
	if name, exists := p.cfg.Headers[breakingChangeMetadataKey]; exists && len(breakingChanges) > 0 {
		breakingChangeSection = BreakingChangeSection{Name: name, Messages: breakingChanges}
	}
	return ReleaseNote{Version: version, Tag: tag, Date: date.Truncate(time.Minute), Sections: sections, BreakingChanges: breakingChangeSection, AuthorsNames: authors}
}

// ReleaseNote release note.
type ReleaseNote struct {
	Version         *semver.Version
	Tag             string
	Date            time.Time
	Sections        map[string]ReleaseNoteSection
	BreakingChanges BreakingChangeSection
	AuthorsNames    map[string]struct{}
}

// BreakingChangeSection breaking change section.
type BreakingChangeSection struct {
	Name     string
	Messages []string
}

// ReleaseNoteSection release note section.
type ReleaseNoteSection struct {
	Name  string
	Types []string
	Items []GitCommitLog
}
