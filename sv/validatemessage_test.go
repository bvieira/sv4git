package sv

import (
	"testing"
)

const (
	branchIssueRegex = "^([a-z]+\\/)?([A-Z]+-[0-9]+)(-.*)?"
	issueRegex       = "[A-Z]+-[0-9]+"
)

// messages samples start
var fullMessage = `fix: correct minor typos in code

see the issue for details

on typos fixed.

Reviewed-by: Z
Refs #133`
var fullMessageWithJira = `fix: correct minor typos in code

see the issue for details

on typos fixed.

Reviewed-by: Z
Refs #133
jira: JIRA-456`
var fullMessageRefs = `fix: correct minor typos in code

see the issue for details

on typos fixed.

Refs #133`
var subjectAndBodyMessage = `fix: correct minor typos in code

see the issue for details

on typos fixed.`
var subjectAndFooterMessage = `refactor!: drop support for Node 6

BREAKING CHANGE: refactor to use JavaScript features not available in Node 6.`

// multiline samples end

func TestValidateMessageProcessorImpl_Validate(t *testing.T) {
	p := NewValidateMessageProcessor([]string{"develop", "master"}, []string{"feat", "fix"}, "jira", branchIssueRegex, issueRegex)

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

func TestValidateMessageProcessorImpl_Enhance(t *testing.T) {
	p := NewValidateMessageProcessor([]string{"develop", "master"}, []string{"feat", "fix"}, "jira", branchIssueRegex, issueRegex)

	tests := []struct {
		name    string
		branch  string
		message string
		want    string
		wantErr bool
	}{
		{"issue on branch name", "JIRA-123", "fix: fix something", "\njira: JIRA-123", false},
		{"issue on branch name with description", "JIRA-123-some-description", "fix: fix something", "\njira: JIRA-123", false},
		{"issue on branch name with prefix", "feature/JIRA-123", "fix: fix something", "\njira: JIRA-123", false},
		{"with footer", "JIRA-123", fullMessage, "jira: JIRA-123", false},
		{"with issue on footer", "JIRA-123", fullMessageWithJira, "", false},
		{"issue on branch name with prefix and description", "feature/JIRA-123-some-description", "fix: fix something", "\njira: JIRA-123", false},
		{"no issue on branch name", "branch", "fix: fix something", "", true},
		{"unexpected branch name", "feature /JIRA-123", "fix: fix something", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Enhance(tt.branch, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessageProcessorImpl.Enhance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateMessageProcessorImpl.Enhance() = %v, want %v", got, tt.want)
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

func Test_hasIssueID(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		issueKeyName string
		want         bool
	}{
		{"single line without issue", "feat: something", "jira", false},
		{"multi line without issue", `feat: something
		
yay`, "jira", false},
		{"multi line without jira issue", `feat: something
		
jira1: JIRA-123`, "jira", false},
		{"multi line with issue", `feat: something
		
jira: JIRA-123`, "jira", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasIssueID(tt.message, tt.issueKeyName); got != tt.want {
				t.Errorf("hasIssueID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasFooter(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{"simple message", "feat: add something", false},
		{"full messsage", fullMessage, true},
		{"full messsage with refs", fullMessageRefs, true},
		{"subject and footer message", subjectAndFooterMessage, true},
		{"subject and body message", subjectAndBodyMessage, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasFooter(tt.message); got != tt.want {
				t.Errorf("hasFooter() = %v, want %v", got, tt.want)
			}
		})
	}
}
