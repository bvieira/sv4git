package sv

import (
	"reflect"
	"testing"
)

var cfg = CommitMessageConfig{
	Types: []string{"feat", "fix"},
	Scope: CommitMessageScopeConfig{},
	Footer: map[string]CommitMessageFooterConfig{
		"issue":           {Key: "jira", KeySynonyms: []string{"Jira"}, Regex: "[A-Z]+-[0-9]+"},
		"breaking-change": {Key: "BREAKING CHANGE", KeySynonyms: []string{"BREAKING CHANGES"}},
		"refs":            {Key: "Refs", UseHash: true},
	},
}

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

func TestMessageProcessorImpl_Validate(t *testing.T) {
	p := NewMessageProcessor(cfg, []string{"develop", "master"}, branchIssueRegex)

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
				t.Errorf("MessageProcessorImpl.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageProcessorImpl_Enhance(t *testing.T) {
	p := NewMessageProcessor(cfg, []string{"develop", "master"}, branchIssueRegex)

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
				t.Errorf("MessageProcessorImpl.Enhance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MessageProcessorImpl.Enhance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageProcessorImpl_IssueID(t *testing.T) {
	p := NewMessageProcessor(cfg, []string{"develop", "master"}, branchIssueRegex)

	tests := []struct {
		name    string
		branch  string
		want    string
		wantErr bool
	}{
		{"simple branch", "JIRA-123", "JIRA-123", false},
		{"branch with prefix", "feature/JIRA-123", "JIRA-123", false},
		{"branch with prefix and posfix", "feature/JIRA-123-some-description", "JIRA-123", false},
		{"branch not found", "feature/wrong123-some-description", "", false},
		{"empty branch", "", "", false},
		{"unexpected branch name", "feature /JIRA-123", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.IssueID(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageProcessorImpl.IssueID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MessageProcessorImpl.IssueID() = %v, want %v", got, tt.want)
			}
		})
	}
}

const (
	multilineBody = `a
b
c`
	fullFooter = `BREAKING CHANGE: breaks
jira: JIRA-123`
)

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
			if got := hasFooter(tt.message, "BREAKING CHANGE"); got != tt.want {
				t.Errorf("hasFooter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// conventional commit tests

var completeBody = `some descriptions

jira: JIRA-123
BREAKING CHANGE: this change breaks everything`

var issueOnlyBody = `some descriptions

jira: JIRA-456`

var issueSynonymsBody = `some descriptions

Jira: JIRA-789`

var hashMetadataBody = `some descriptions

Jira: JIRA-999
Refs #123`

func TestMessageProcessorImpl_Parse(t *testing.T) {
	p := NewMessageProcessor(cfg, []string{"develop", "master"}, branchIssueRegex)

	tests := []struct {
		name    string
		subject string
		body    string
		want    CommitMessage
	}{
		{"simple message", "feat: something awesome", "", CommitMessage{Type: "feat", Scope: "", Description: "something awesome", Body: "", IsBreakingChange: false, Metadata: map[string]string{}}},
		{"message with scope", "feat(scope): something awesome", "", CommitMessage{Type: "feat", Scope: "scope", Description: "something awesome", Body: "", IsBreakingChange: false, Metadata: map[string]string{}}},
		{"unmapped type", "unkn: something unknown", "", CommitMessage{Type: "unkn", Scope: "", Description: "something unknown", Body: "", IsBreakingChange: false, Metadata: map[string]string{}}},
		{"jira and breaking change metadata", "feat: something new", completeBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: completeBody, IsBreakingChange: true, Metadata: map[string]string{issueKey: "JIRA-123", breakingKey: "this change breaks everything"}}},
		{"jira only metadata", "feat: something new", issueOnlyBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: issueOnlyBody, IsBreakingChange: false, Metadata: map[string]string{issueKey: "JIRA-456"}}},
		{"jira synonyms metadata", "feat: something new", issueSynonymsBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: issueSynonymsBody, IsBreakingChange: false, Metadata: map[string]string{issueKey: "JIRA-789"}}},
		{"breaking change with exclamation mark", "feat!: something new", "", CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: "", IsBreakingChange: true, Metadata: map[string]string{}}},
		{"hash metadata", "feat: something new", hashMetadataBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: hashMetadataBody, IsBreakingChange: false, Metadata: map[string]string{issueKey: "JIRA-999", "refs": "#123"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.Parse(tt.subject, tt.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageProcessorImpl.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageProcessorImpl_Format(t *testing.T) {
	p := NewMessageProcessor(cfg, []string{"develop", "master"}, branchIssueRegex)

	tests := []struct {
		name       string
		msg        CommitMessage
		wantHeader string
		wantBody   string
		wantFooter string
	}{
		{"simple message", NewCommitMessage("feat", "", "something", "", "", ""), "feat: something", "", ""},
		{"with issue", NewCommitMessage("feat", "", "something", "", "JIRA-123", ""), "feat: something", "", "jira: JIRA-123"},
		{"with breaking change", NewCommitMessage("feat", "", "something", "", "", "breaks"), "feat: something", "", "BREAKING CHANGE: breaks"},
		{"with scope", NewCommitMessage("feat", "scope", "something", "", "", ""), "feat(scope): something", "", ""},
		{"with body", NewCommitMessage("feat", "", "something", "body", "", ""), "feat: something", "body", ""},
		{"with multiline body", NewCommitMessage("feat", "", "something", multilineBody, "", ""), "feat: something", multilineBody, ""},
		{"full message", NewCommitMessage("feat", "scope", "something", multilineBody, "JIRA-123", "breaks"), "feat(scope): something", multilineBody, fullFooter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := p.Format(tt.msg)
			if got != tt.wantHeader {
				t.Errorf("MessageProcessorImpl.Format() header got = %v, want %v", got, tt.wantHeader)
			}
			if got1 != tt.wantBody {
				t.Errorf("MessageProcessorImpl.Format() body got = %v, want %v", got1, tt.wantBody)
			}
			if got2 != tt.wantFooter {
				t.Errorf("MessageProcessorImpl.Format() footer got = %v, want %v", got2, tt.wantFooter)
			}
		})
	}
}
