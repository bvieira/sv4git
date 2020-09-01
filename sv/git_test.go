package sv

import (
	"reflect"
	"testing"
	"time"
)

func Test_parseTagsOutput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []GitTag
		wantErr bool
	}{
		{"with date", "2020-05-01 18:00:00 -0300#1.0.0", []GitTag{{Name: "1.0.0", Date: date("2020-05-01 18:00:00 -0300")}}, false},
		{"without date", "#1.0.0", []GitTag{{Name: "1.0.0", Date: time.Time{}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTagsOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTagsOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTagsOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func date(input string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05 -0700", input)
	if err != nil {
		panic(err)
	}
	return t
}
