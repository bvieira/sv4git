package sv

import (
	"reflect"
	"testing"
	"time"
)

func TestReleaseNoteProcessorImpl_Get(t *testing.T) {
	date := time.Now()

	tests := []struct {
		name    string
		date    time.Time
		commits []GitCommitLog
		want    ReleaseNote
	}{
		{
			name:    "mapped tag",
			date:    date,
			commits: []GitCommitLog{commitlog("t1", map[string]string{})},
			want:    releaseNote(date, map[string]ReleaseNoteSection{"t1": rnSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, nil),
		},
		{
			name:    "unmapped tag",
			date:    date,
			commits: []GitCommitLog{commitlog("t1", map[string]string{}), commitlog("unmapped", map[string]string{})},
			want:    releaseNote(date, map[string]ReleaseNoteSection{"t1": rnSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, nil),
		},
		{
			name:    "breaking changes tag",
			date:    date,
			commits: []GitCommitLog{commitlog("t1", map[string]string{}), commitlog("unmapped", map[string]string{"breakingchange": "breaks"})},
			want:    releaseNote(date, map[string]ReleaseNoteSection{"t1": rnSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, []string{"breaks"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewReleaseNoteProcessor(map[string]string{"t1": "Tag 1", "t2": "Tag 2"})
			if got := p.Get(tt.date, tt.commits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReleaseNoteProcessorImpl.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
