package sv

import (
	"reflect"
	"testing"
)

func TestReleaseNoteProcessorImpl_Get(t *testing.T) {
	tests := []struct {
		name    string
		commits []GitCommitLog
		want    ReleaseNote
	}{
		{
			name:    "mapped tag",
			commits: []GitCommitLog{commitlog("t1", map[string]string{})},
			want:    releaseNote(map[string]ReleaseNoteSection{"t1": rnSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, nil),
		},
		{
			name:    "unmapped tag",
			commits: []GitCommitLog{commitlog("t1", map[string]string{}), commitlog("unmapped", map[string]string{})},
			want:    releaseNote(map[string]ReleaseNoteSection{"t1": rnSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, nil),
		},
		{
			name:    "breaking changes tag",
			commits: []GitCommitLog{commitlog("t1", map[string]string{}), commitlog("unmapped", map[string]string{"breakingchange": "breaks"})},
			want:    releaseNote(map[string]ReleaseNoteSection{"t1": rnSection("Tag 1", []GitCommitLog{commitlog("t1", map[string]string{})})}, []string{"breaks"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewReleaseNoteProcessor(map[string]string{"t1": "Tag 1", "t2": "Tag 2"})
			if got := p.Get(tt.commits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReleaseNoteProcessorImpl.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
