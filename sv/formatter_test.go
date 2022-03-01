package sv

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
)

var templatesFS = os.DirFS("../cmd/git-sv/resources/templates")

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
			got, err := NewOutputFormatter(templatesFS).FormatReleaseNote(tt.input)
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
	sections := []ReleaseNoteSection{
		newReleaseNoteCommitsSection("Features", []string{"feat"}, []GitCommitLog{commitlog("feat", map[string]string{}, "a")}),
		newReleaseNoteCommitsSection("Bug Fixes", []string{"fix"}, []GitCommitLog{commitlog("fix", map[string]string{}, "a")}),
		newReleaseNoteCommitsSection("Build", []string{"build"}, []GitCommitLog{commitlog("build", map[string]string{}, "a")}),
		ReleaseNoteBreakingChangeSection{"Breaking Changes", []string{"break change message"}},
	}
	return releaseNote(v, tag, date, sections, map[string]struct{}{"a": {}})
}

func Test_checkTemplatesExecution(t *testing.T) {
	tpls := NewOutputFormatter(templatesFS).templates
	tests := []struct {
		template  string
		variables interface{}
	}{
		{"changelog-md.tpl", changelogVariables("v1.0.0", "v1.0.1")},
		{"releasenotes-md.tpl", releaseNotesVariables("v1.0.0")},
	}
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			var b bytes.Buffer
			err := tpls.ExecuteTemplate(&b, tt.template, tt.variables)
			if err != nil {
				t.Errorf("invalid template err = %v", err)
				return
			}
			if len(b.Bytes()) <= 0 {
				t.Errorf("empty template")
			}
		})
	}
}

func releaseNotesVariables(release string) releaseNoteTemplateVariables {
	return releaseNoteTemplateVariables{
		Release: release,
		Date:    time.Date(2006, 1, 02, 0, 0, 0, 0, time.UTC),
		Sections: []ReleaseNoteSection{
			newReleaseNoteCommitsSection("Features", []string{"feat"}, []GitCommitLog{commitlog("feat", map[string]string{}, "a")}),
			newReleaseNoteCommitsSection("Bug Fixes", []string{"fix"}, []GitCommitLog{commitlog("fix", map[string]string{}, "a")}),
			newReleaseNoteCommitsSection("Build", []string{"build"}, []GitCommitLog{commitlog("build", map[string]string{}, "a")}),
			ReleaseNoteBreakingChangeSection{"Breaking Changes", []string{"break change message"}},
		},
	}
}

func changelogVariables(releases ...string) []releaseNoteTemplateVariables {
	var variables []releaseNoteTemplateVariables
	for _, r := range releases {
		variables = append(variables, releaseNotesVariables(r))
	}
	return variables

}
