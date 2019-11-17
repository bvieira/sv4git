package sv

import (
	"time"

	"github.com/Masterminds/semver"
)

func version(v string) semver.Version {
	r, _ := semver.NewVersion(v)
	return *r
}

func commitlog(t string, metadata map[string]string) GitCommitLog {
	return GitCommitLog{
		Type:     t,
		Subject:  "subject text",
		Metadata: metadata,
	}
}

func releaseNote(sections map[string]ReleaseNoteSection, breakingChanges []string) ReleaseNote {
	return ReleaseNote{
		Date:            time.Now().Truncate(time.Minute),
		Sections:        sections,
		BreakingChanges: breakingChanges,
	}
}

func rnSection(name string, items []GitCommitLog) ReleaseNoteSection {
	return ReleaseNoteSection{
		Name:  name,
		Items: items,
	}
}
