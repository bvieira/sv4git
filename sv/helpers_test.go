package sv

import (
	"time"

	"github.com/Masterminds/semver/v3"
)

func version(v string) semver.Version {
	r, _ := semver.NewVersion(v)
	return *r
}

func commitlog(ctype string, metadata map[string]string) GitCommitLog {
	breaking := false
	if _, found := metadata[breakingChangeMetadataKey]; found {
		breaking = true
	}
	return GitCommitLog{
		Message: CommitMessage{
			Type:             ctype,
			Description:      "subject text",
			IsBreakingChange: breaking,
			Metadata:         metadata,
		},
	}
}

func releaseNote(version *semver.Version, date time.Time, sections map[string]ReleaseNoteSection, breakingChanges []string) ReleaseNote {
	var bchanges BreakingChangeSection
	if len(breakingChanges) > 0 {
		bchanges = BreakingChangeSection{Name: "Breaking Changes", Messages: breakingChanges}
	}
	return ReleaseNote{
		Version:         version,
		Date:            date.Truncate(time.Minute),
		Sections:        sections,
		BreakingChanges: bchanges,
	}
}

func newReleaseNoteSection(name string, items []GitCommitLog) ReleaseNoteSection {
	return ReleaseNoteSection{
		Name:  name,
		Items: items,
	}
}
