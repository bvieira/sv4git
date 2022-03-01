package sv

import (
	"time"

	"github.com/Masterminds/semver/v3"
)

func version(v string) *semver.Version {
	r, _ := semver.NewVersion(v)
	return r
}

func commitlog(ctype string, metadata map[string]string, author string) GitCommitLog {
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
		AuthorName: author,
	}
}

func releaseNote(version *semver.Version, tag string, date time.Time, sections []ReleaseNoteSection, authorsNames map[string]struct{}) ReleaseNote {
	return ReleaseNote{
		Version:      version,
		Tag:          tag,
		Date:         date.Truncate(time.Minute),
		Sections:     sections,
		AuthorsNames: authorsNames,
	}
}

func newReleaseNoteCommitsSection(name string, types []string, items []GitCommitLog) ReleaseNoteCommitsSection {
	return ReleaseNoteCommitsSection{
		Name:  name,
		Types: types,
		Items: items,
	}
}
