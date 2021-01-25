package sv

import (
	"testing"
	"time"

	"github.com/Masterminds/semver"
)

var dateChangelog = `## v1.0.0 (2020-05-01)
`
var emptyDateChangelog = `## v1.0.0
`
var emptyVersionChangelog = `## 2020-05-01
`

func TestOutputFormatterImpl_FormatReleaseNote(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2020-05-01")

	tests := []struct {
		name  string
		input ReleaseNote
		want  string
	}{
		{"with date", emptyReleaseNote("1.0.0", date.Truncate(time.Minute)), dateChangelog},
		{"without date", emptyReleaseNote("1.0.0", time.Time{}.Truncate(time.Minute)), emptyDateChangelog},
		{"without version", emptyReleaseNote("", date.Truncate(time.Minute)), emptyVersionChangelog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOutputFormatter().FormatReleaseNote(tt.input); got != tt.want {
				t.Errorf("OutputFormatterImpl.FormatReleaseNote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func emptyReleaseNote(version string, date time.Time) ReleaseNote {
	var v *semver.Version
	if version != "" {
		v = semver.MustParse(version)
	}
	return ReleaseNote{
		Version: v,
		Date:    date,
	}
}
