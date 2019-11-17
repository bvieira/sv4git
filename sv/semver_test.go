package sv

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
)

func TestSemVerCommitsProcessorImpl_NextVersion(t *testing.T) {
	tests := []struct {
		name           string
		unknownAsPatch bool
		version        semver.Version
		commits        []GitCommitLog
		want           semver.Version
	}{
		{"no update", false, version("0.0.0"), []GitCommitLog{}, version("0.0.0")},
		{"no update on unknown type", false, version("0.0.0"), []GitCommitLog{commitlog("a", map[string]string{})}, version("0.0.0")},
		{"update patch on unknown type", true, version("0.0.0"), []GitCommitLog{commitlog("a", map[string]string{})}, version("0.0.1")},
		{"patch update", false, version("0.0.0"), []GitCommitLog{commitlog("patch", map[string]string{})}, version("0.0.1")},
		{"minor update", false, version("0.0.0"), []GitCommitLog{commitlog("patch", map[string]string{}), commitlog("minor", map[string]string{})}, version("0.1.0")},
		{"major update", false, version("0.0.0"), []GitCommitLog{commitlog("patch", map[string]string{}), commitlog("major", map[string]string{})}, version("1.0.0")},
		{"breaking change update", false, version("0.0.0"), []GitCommitLog{commitlog("patch", map[string]string{}), commitlog("patch", map[string]string{"breakingchange": "break"})}, version("1.0.0")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewSemVerCommitsProcessor(tt.unknownAsPatch, []string{"major"}, []string{"minor"}, []string{"patch"})
			if got := p.NextVersion(tt.version, tt.commits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SemVerCommitsProcessorImpl.NextVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    semver.Version
		wantErr bool
	}{
		{"empty version", "", version("0.0.0"), false},
		{"invalid version", "abc", semver.Version{}, true},
		{"valid version", "1.2.3", version("1.2.3"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
