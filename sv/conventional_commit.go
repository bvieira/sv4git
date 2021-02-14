package sv

import (
	"regexp"
	"strings"
)

const (
	breakingKey = "breaking-change"
	// IssueIDKey key to issue id metadata
	issueKey = "issue"
)

// CommitMessageConfig config a commit message
type CommitMessageConfig struct {
	Types  []string
	Scope  ScopeConfig
	Footer map[string]FooterMetadataConfig
}

// ScopeConfig config scope preferences
type ScopeConfig struct {
	Mandatory bool
	Values    []string
}

// FooterMetadataConfig config footer metadata
type FooterMetadataConfig struct {
	Key         string
	KeySynonyms []string
	Regex       string
	UseHash     bool
}

// CommitMessage is a message using conventional commits.
type CommitMessage struct {
	Type             string            `json:"type,omitempty"`
	Scope            string            `json:"scope,omitempty"`
	Description      string            `json:"description,omitempty"`
	Body             string            `json:"body,omitempty"`
	IsBreakingChange bool              `json:"isBreakingChange,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Issue return issue from metadata.
func (m CommitMessage) Issue() string {
	return m.Metadata[issueKey]
}

// BreakingMessage return breaking change message from metadata.
func (m CommitMessage) BreakingMessage() string {
	return m.Metadata[breakingKey]
}

// CommitMessageProcessor parse commit messages.
type CommitMessageProcessor interface {
	Parse(subject, body string) CommitMessage
}

// CommitMessageProcessorImpl commit message processor implementation
type CommitMessageProcessorImpl struct {
	cfg CommitMessageConfig
}

// NewCommitMessageProcessor CommitMessageProcessorImpl constructor
func NewCommitMessageProcessor(cfg CommitMessageConfig) CommitMessageProcessor {
	return &CommitMessageProcessorImpl{cfg: cfg}
}

// Parse parse a commit message
func (p CommitMessageProcessorImpl) Parse(subject, body string) CommitMessage {
	commitType, scope, description, hasBreakingChange := parseSubjectMessage(subject)

	metadata := make(map[string]string)
	for key, mdCfg := range p.cfg.Footer {
		prefixes := append([]string{mdCfg.Key}, mdCfg.KeySynonyms...)
		for _, prefix := range prefixes {
			if tagValue := extractFooterMetadata(prefix, body, mdCfg.UseHash); tagValue != "" {
				metadata[key] = tagValue
				break
			}
		}
	}

	if _, exists := metadata[breakingKey]; exists {
		hasBreakingChange = true
	}

	return CommitMessage{
		Type:             commitType,
		Scope:            scope,
		Description:      description,
		Body:             body,
		IsBreakingChange: hasBreakingChange,
		Metadata:         metadata,
	}
}

func parseSubjectMessage(message string) (string, string, string, bool) {
	regex := regexp.MustCompile("([a-z]+)(\\((.*)\\))?(!)?: (.*)")
	result := regex.FindStringSubmatch(message)
	if len(result) != 6 {
		return "", "", message, false
	}
	return result[1], result[3], strings.TrimSpace(result[5]), result[4] == "!"
}

func extractFooterMetadata(key, text string, useHash bool) string {
	var regex *regexp.Regexp
	if useHash {
		regex = regexp.MustCompile(key + " (#.*)")
	} else {
		regex = regexp.MustCompile(key + ": (.*)")
	}

	result := regex.FindStringSubmatch(text)
	if len(result) < 2 {
		return ""
	}
	return result[1]
}
