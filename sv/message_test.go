package sv

import (
	"reflect"
	"testing"
)

var ccfg = CommitMessageConfig{
	Types: []string{"feat", "fix"},
	Scope: CommitMessageScopeConfig{},
	Footer: map[string]CommitMessageFooterConfig{
		"issue": {Key: "jira", KeySynonyms: []string{"Jira"}},
		"refs":  {Key: "Refs", UseHash: true},
	},
	Issue: CommitMessageIssueConfig{Regex: "[A-Z]+-[0-9]+"},
}

var ccfgEmptyIssue = CommitMessageConfig{
	Types: []string{"feat", "fix"},
	Scope: CommitMessageScopeConfig{},
	Footer: map[string]CommitMessageFooterConfig{
		"issue": {},
	},
	Issue: CommitMessageIssueConfig{Regex: "[A-Z]+-[0-9]+"},
}

var ccfgWithScope = CommitMessageConfig{
	Types: []string{"feat", "fix"},
	Scope: CommitMessageScopeConfig{Values: []string{"", "scope"}},
	Footer: map[string]CommitMessageFooterConfig{
		"issue": {Key: "jira", KeySynonyms: []string{"Jira"}},
		"refs":  {Key: "Refs", UseHash: true},
	},
	Issue: CommitMessageIssueConfig{Regex: "[A-Z]+-[0-9]+"},
}

func newBranchCfg(skipDetached bool) BranchesConfig {
	return BranchesConfig{
		PrefixRegex:  "([a-z]+\\/)?",
		SuffixRegex:  "(-.*)?",
		Skip:         []string{"develop", "master"},
		SkipDetached: &skipDetached,
	}
}

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

func TestMessageProcessorImpl_SkipBranch(t *testing.T) {
	tests := []struct {
		name     string
		bcfg     BranchesConfig
		branch   string
		detached bool
		want     bool
	}{
		{"normal branch", newBranchCfg(false), "JIRA-123", false, false},
		{"dont ignore detached branch", newBranchCfg(false), "JIRA-123", true, false},
		{"ignore branch on skip list", newBranchCfg(false), "master", false, true},
		{"ignore detached branch", newBranchCfg(true), "JIRA-123", true, true},
		{"null skip detached", BranchesConfig{Skip: []string{}}, "JIRA-123", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewMessageProcessor(ccfg, tt.bcfg)
			if got := p.SkipBranch(tt.branch, tt.detached); got != tt.want {
				t.Errorf("MessageProcessorImpl.SkipBranch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageProcessorImpl_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     CommitMessageConfig
		message string
		wantErr bool
	}{
		{"single line valid message", ccfg, "feat: add something", false},
		{"single line valid message with scope", ccfg, "feat(scope): add something", false},
		{"single line valid scope from list", ccfgWithScope, "feat(scope): add something", false},
		{"single line invalid scope from list", ccfgWithScope, "feat(invalid): add something", true},
		{"single line invalid type message", ccfg, "something: add something", true},
		{"single line invalid type message", ccfg, "feat?: add something", true},

		{"multi line valid message", ccfg, `feat: add something
		
		team: x`, false},

		{"multi line invalid message", ccfg, `feat add something
		
		team: x`, true},

		{"support ! for breaking change", ccfg, "feat!: add something", false},
		{"support ! with scope for breaking change", ccfg, "feat(scope)!: add something", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewMessageProcessor(tt.cfg, newBranchCfg(false))
			if err := p.Validate(tt.message); (err != nil) != tt.wantErr {
				t.Errorf("MessageProcessorImpl.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessageProcessorImpl_Enhance(t *testing.T) {
	p := NewMessageProcessor(ccfg, newBranchCfg(false))

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
	p := NewMessageProcessor(ccfg, newBranchCfg(false))

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

func Test_hasIssueID(t *testing.T) {
	cfgColon := CommitMessageFooterConfig{Key: "jira"}
	cfgHash := CommitMessageFooterConfig{Key: "jira", UseHash: true}
	cfgEmpty := CommitMessageFooterConfig{}

	tests := []struct {
		name     string
		message  string
		issueCfg CommitMessageFooterConfig
		want     bool
	}{
		{"single line without issue", "feat: something", cfgColon, false},
		{"multi line without issue", `feat: something
		
yay`, cfgColon, false},
		{"multi line without jira issue", `feat: something
		
jira1: JIRA-123`, cfgColon, false},
		{"multi line with issue", `feat: something
		
jira: JIRA-123`, cfgColon, true},
		{"multi line with issue and hash", `feat: something
		
jira #JIRA-123`, cfgHash, true},
		{"empty config", `feat: something
		
jira #JIRA-123`, cfgEmpty, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasIssueID(tt.message, tt.issueCfg); got != tt.want {
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
	tests := []struct {
		name    string
		cfg     CommitMessageConfig
		subject string
		body    string
		want    CommitMessage
	}{
		{"simple message", ccfg, "feat: something awesome", "", CommitMessage{Type: "feat", Scope: "", Description: "something awesome", Body: "", IsBreakingChange: false, Metadata: map[string]string{}}},
		{"message with scope", ccfg, "feat(scope): something awesome", "", CommitMessage{Type: "feat", Scope: "scope", Description: "something awesome", Body: "", IsBreakingChange: false, Metadata: map[string]string{}}},
		{"unmapped type", ccfg, "unkn: something unknown", "", CommitMessage{Type: "unkn", Scope: "", Description: "something unknown", Body: "", IsBreakingChange: false, Metadata: map[string]string{}}},
		{"jira and breaking change metadata", ccfg, "feat: something new", completeBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: completeBody, IsBreakingChange: true, Metadata: map[string]string{issueMetadataKey: "JIRA-123", breakingChangeMetadataKey: "this change breaks everything"}}},
		{"jira only metadata", ccfg, "feat: something new", issueOnlyBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: issueOnlyBody, IsBreakingChange: false, Metadata: map[string]string{issueMetadataKey: "JIRA-456"}}},
		{"jira synonyms metadata", ccfg, "feat: something new", issueSynonymsBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: issueSynonymsBody, IsBreakingChange: false, Metadata: map[string]string{issueMetadataKey: "JIRA-789"}}},
		{"breaking change with exclamation mark", ccfg, "feat!: something new", "", CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: "", IsBreakingChange: true, Metadata: map[string]string{}}},
		{"hash metadata", ccfg, "feat: something new", hashMetadataBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: hashMetadataBody, IsBreakingChange: false, Metadata: map[string]string{issueMetadataKey: "JIRA-999", "refs": "#123"}}},
		{"empty issue cfg", ccfgEmptyIssue, "feat: something new", hashMetadataBody, CommitMessage{Type: "feat", Scope: "", Description: "something new", Body: hashMetadataBody, IsBreakingChange: false, Metadata: map[string]string{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMessageProcessor(tt.cfg, newBranchCfg(false)).Parse(tt.subject, tt.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageProcessorImpl.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageProcessorImpl_Format(t *testing.T) {
	tests := []struct {
		name       string
		cfg        CommitMessageConfig
		msg        CommitMessage
		wantHeader string
		wantBody   string
		wantFooter string
	}{
		{"simple message", ccfg, NewCommitMessage("feat", "", "something", "", "", ""), "feat: something", "", ""},
		{"with issue", ccfg, NewCommitMessage("feat", "", "something", "", "JIRA-123", ""), "feat: something", "", "jira: JIRA-123"},
		{"with breaking change", ccfg, NewCommitMessage("feat", "", "something", "", "", "breaks"), "feat: something", "", "BREAKING CHANGE: breaks"},
		{"with scope", ccfg, NewCommitMessage("feat", "scope", "something", "", "", ""), "feat(scope): something", "", ""},
		{"with body", ccfg, NewCommitMessage("feat", "", "something", "body", "", ""), "feat: something", "body", ""},
		{"with multiline body", ccfg, NewCommitMessage("feat", "", "something", multilineBody, "", ""), "feat: something", multilineBody, ""},
		{"full message", ccfg, NewCommitMessage("feat", "scope", "something", multilineBody, "JIRA-123", "breaks"), "feat(scope): something", multilineBody, fullFooter},
		{"config without issue key", ccfgEmptyIssue, NewCommitMessage("feat", "", "something", "", "JIRA-123", ""), "feat: something", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := NewMessageProcessor(tt.cfg, newBranchCfg(false)).Format(tt.msg)
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

var expectedBodyFullMessage = `
see the issue for details

on typos fixed.

Reviewed-by: Z
Refs #133`

func Test_splitCommitMessageContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantSubject string
		wantBody    string
	}{
		{"single line commit", "feat: something", "feat: something", ""},
		{"multi line commit", fullMessage, "fix: correct minor typos in code", expectedBodyFullMessage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitCommitMessageContent(tt.content)
			if got != tt.wantSubject {
				t.Errorf("splitCommitMessageContent() subject got = %v, want %v", got, tt.wantSubject)
			}
			if got1 != tt.wantBody {
				t.Errorf("splitCommitMessageContent() body got1 = [%v], want [%v]", got1, tt.wantBody)
			}
		})
	}
}

//commitType, scope, description, hasBreakingChange
func Test_parseSubjectMessage(t *testing.T) {
	tests := []struct {
		name                  string
		message               string
		wantType              string
		wantScope             string
		wantDescription       string
		wantHasBreakingChange bool
	}{
		{"valid commit", "feat: something", "feat", "", "something", false},
		{"valid commit with scope", "feat(scope): something", "feat", "scope", "something", false},
		{"valid commit with breaking change", "feat(scope)!: something", "feat", "scope", "something", true},
		{"missing description", "feat: ", "feat", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctype, scope, description, hasBreakingChange := parseSubjectMessage(tt.message)
			if ctype != tt.wantType {
				t.Errorf("parseSubjectMessage() type got = %v, want %v", ctype, tt.wantType)
			}
			if scope != tt.wantScope {
				t.Errorf("parseSubjectMessage() scope got = %v, want %v", scope, tt.wantScope)
			}
			if description != tt.wantDescription {
				t.Errorf("parseSubjectMessage() description got = %v, want %v", description, tt.wantDescription)
			}
			if hasBreakingChange != tt.wantHasBreakingChange {
				t.Errorf("parseSubjectMessage() hasBreakingChange got = %v, want %v", hasBreakingChange, tt.wantHasBreakingChange)
			}
		})
	}
}
