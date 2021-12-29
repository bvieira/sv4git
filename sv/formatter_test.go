package sv

import (
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
)

var dateChangelog = `## v1.0.0 (2020-05-01)
`

var nonVersioningChangelog = `## abc (2020-05-01)
`

var emptyDateChangelog = `## v1.0.0
`

var emptyVersionChangelog = `## 2020-05-01
`

var fullChangeLog = `## v1.0.0 (2020-05-01)

### Features

- subject text ()

### Bug Fixes

- subject text ()

### Build

- subject text ()

### Breaking Changes

- break change message
`

func TestOutputFormatterImpl_FormatReleaseNote(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2020-05-01")

	tests := []struct {
		name    string
		input   ReleaseNote
		want    string
		wantErr bool
	}{
		{"with date", emptyReleaseNote("1.0.0", date.Truncate(time.Minute)), dateChangelog, false},
		{"without date", emptyReleaseNote("1.0.0", time.Time{}.Truncate(time.Minute)), emptyDateChangelog, false},
		{"without version", emptyReleaseNote("", date.Truncate(time.Minute)), emptyVersionChangelog, false},
		{"non versioning tag", emptyReleaseNote("abc", date.Truncate(time.Minute)), nonVersioningChangelog, false},
		{"full changelog", fullReleaseNote("1.0.0", date.Truncate(time.Minute)), fullChangeLog, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOutputFormatter().FormatReleaseNote(tt.input)
			if got != tt.want {
				t.Errorf("OutputFormatterImpl.FormatReleaseNote() = %v, want %v", got, tt.want)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("OutputFormatterImpl.FormatReleaseNote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func emptyReleaseNote(tag string, date time.Time) ReleaseNote {
	v, _ := semver.NewVersion(tag)
	return ReleaseNote{
		Version: v,
		Tag:     tag,
		Date:    date,
	}
}

func fullReleaseNote(tag string, date time.Time) ReleaseNote {
	v, _ := semver.NewVersion(tag)
	sections := map[string]ReleaseNoteSection{
		"build": newReleaseNoteSection("Build", []GitCommitLog{commitlog("build", map[string]string{})}),
		"feat":  newReleaseNoteSection("Features", []GitCommitLog{commitlog("feat", map[string]string{})}),
		"fix":   newReleaseNoteSection("Bug Fixes", []GitCommitLog{commitlog("fix", map[string]string{})}),
	}
	return releaseNote(v, tag, date, sections, []string{"break change message"})
}
