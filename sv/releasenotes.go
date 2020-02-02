package sv

import (
	"time"

	"github.com/Masterminds/semver"
)

// ReleaseNoteProcessor release note processor interface.
type ReleaseNoteProcessor interface {
	Create(version semver.Version, date time.Time, commits []GitCommitLog) ReleaseNote
}

// ReleaseNoteProcessorImpl release note based on commit log.
type ReleaseNoteProcessorImpl struct {
	tags map[string]string
}

// NewReleaseNoteProcessor ReleaseNoteProcessor constructor.
func NewReleaseNoteProcessor(tags map[string]string) *ReleaseNoteProcessorImpl {
	return &ReleaseNoteProcessorImpl{tags: tags}
}

// Create create a release note based on commits.
func (p ReleaseNoteProcessorImpl) Create(version semver.Version, date time.Time, commits []GitCommitLog) ReleaseNote {
	sections := make(map[string]ReleaseNoteSection)
	var breakingChanges []string
	for _, commit := range commits {
		if name, exists := p.tags[commit.Type]; exists {
			section, sexists := sections[commit.Type]
			if !sexists {
				section = ReleaseNoteSection{Name: name}
			}
			section.Items = append(section.Items, commit)
			sections[commit.Type] = section
		}
		if value, exists := commit.Metadata[BreakingChangesKey]; exists {
			breakingChanges = append(breakingChanges, value)
		}
	}

	return ReleaseNote{Version: version, Date: date.Truncate(time.Minute), Sections: sections, BreakingChanges: breakingChanges}
}

// ReleaseNote release note.
type ReleaseNote struct {
	Version         semver.Version
	Date            time.Time
	Sections        map[string]ReleaseNoteSection
	BreakingChanges []string
}

// ReleaseNoteSection release note section.
type ReleaseNoteSection struct {
	Name  string
	Items []GitCommitLog
}
