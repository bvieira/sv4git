package sv

import (
	"reflect"
	"testing"
)

var cfg = CommitMessageConfig{
	Types: []string{"feat", "fix"},
	Scope: ScopeConfig{},
	Footer: map[string]FooterMetadataConfig{
		"issue":           {Key: "jira", KeySynonyms: []string{"Jira"}, Regex: "[A-Z]+-[0-9]+"},
		"breaking-change": {Key: "BREAKING CHANGE", KeySynonyms: []string{"BREAKING CHANGES"}},
		"refs":            {Key: "Refs", UseHash: true},
	},
}

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

func TestCommitMessageProcessorImpl_Parse(t *testing.T) {
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
			p := NewCommitMessageProcessor(cfg)
			if got := p.Parse(tt.subject, tt.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommitMessageProcessorImpl.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
