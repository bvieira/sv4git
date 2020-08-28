package sv

import (
	"testing"
)

func TestValidateMessageProcessorImpl_Validate(t *testing.T) {
	p := NewValidateMessageProcessor([]string{"develop", "master"}, []string{"feat", "fix"})

	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{"single line valid message", "feat: add something", false},
		{"single line valid message with scope", "feat(scope): add something", false},
		{"single line invalid type message", "something: add something", true},
		{"single line invalid type message", "feat?: add something", true},

		{"multi line valid message", `feat: add something
		
		team: x`, false},

		{"multi line invalid message", `feat add something
		
		team: x`, true},

		{"support ! for breaking change", "feat!: add something", false},
		{"support ! with scope for breaking change", "feat(scope)!: add something", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.Validate(tt.message); (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessageProcessorImpl.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_firstLine(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"empty string", "", ""},

		{"single line string", "single line", "single line"},

		{"multi line string", `first line
		last line`, "first line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstLine(tt.value); got != tt.want {
				t.Errorf("firstLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
