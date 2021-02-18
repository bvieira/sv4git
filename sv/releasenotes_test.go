package sv

import (
	"reflect"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
)

func TestReleaseNoteProcessorImpl_Create(t *testing.T) {
	date := time.Now()

	tests := []struct {
		name    string
		version *semver.Version
		date    time.Time
		commits []GitCommitLog
		want    ReleaseNote
	}{
		{
			name:    "mapped tag",
			version: semver.MustParse("1.0.0"),
			date:    date,
			commits: []GitCommitLog{commitlog("t1", map[string]string{})},
			want:    releaseNote(semver.MustParse("1.0.0"), date, map[string]ReleaseNoteSection{"t1": newReleaseNoteSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, nil),
		},
		{
			name:    "unmapped tag",
			version: semver.MustParse("1.0.0"),
			date:    date,
			commits: []GitCommitLog{commitlog("t1", map[string]string{}), commitlog("unmapped", map[string]string{})},
			want:    releaseNote(semver.MustParse("1.0.0"), date, map[string]ReleaseNoteSection{"t1": newReleaseNoteSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, nil),
		},
		{
			name:    "breaking changes tag",
			version: semver.MustParse("1.0.0"),
			date:    date,
			commits: []GitCommitLog{commitlog("t1", map[string]string{}), commitlog("unmapped", map[string]string{"breaking-change": "breaks"})},
			want:    releaseNote(semver.MustParse("1.0.0"), date, map[string]ReleaseNoteSection{"t1": newReleaseNoteSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, []string{"breaks"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewReleaseNoteProcessor(ReleaseNotesConfig{Headers: map[string]string{"t1": "Tag 1", "t2": "Tag 2", "breaking-change": "Breaking Changes"}})
			if got := p.Create(tt.version, tt.date, tt.commits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReleaseNoteProcessorImpl.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}
