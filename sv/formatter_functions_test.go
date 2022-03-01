package sv

import (
	"reflect"
	"testing"
	"time"
)

func Test_timeFormat(t *testing.T) {
	tests := []struct {
		name   string
		time   time.Time
		format string
		want   string
	}{
		{"valid time", time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), "2006-01-02", "2022-01-01"},
		{"empty time", time.Time{}, "2006-01-02", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := timeFormat(tt.time, tt.format); got != tt.want {
				t.Errorf("timeFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSection(t *testing.T) {
	tests := []struct {
		name        string
		sections    []ReleaseNoteSection
		sectionName string
		want        ReleaseNoteSection
	}{
		{"existing section", []ReleaseNoteSection{ReleaseNoteCommitsSection{Name: "section 0"}, ReleaseNoteCommitsSection{Name: "section 1"}, ReleaseNoteCommitsSection{Name: "section 2"}}, "section 1", ReleaseNoteCommitsSection{Name: "section 1"}},
		{"nonexisting section", []ReleaseNoteSection{ReleaseNoteCommitsSection{Name: "section 0"}, ReleaseNoteCommitsSection{Name: "section 1"}, ReleaseNoteCommitsSection{Name: "section 2"}}, "section 10", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSection(tt.sections, tt.sectionName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSection() = %v, want %v", got, tt.want)
			}
		})
	}
}
