package sv

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

const (
	breakingKey = "breaking-change"
	// IssueIDKey key to issue id metadata
	issueKey = "issue"
)

// CommitMessage is a message using conventional commits.
type CommitMessage struct {
	Type             string            `json:"type,omitempty"`
	Scope            string            `json:"scope,omitempty"`
	Description      string            `json:"description,omitempty"`
	Body             string            `json:"body,omitempty"`
	IsBreakingChange bool              `json:"isBreakingChange,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// NewCommitMessage commit message constructor
func NewCommitMessage(ctype, scope, description, body, issue, breakingChanges string) CommitMessage {
	metadata := make(map[string]string)
	if issue != "" {
		metadata[issueKey] = issue
	}
	if breakingChanges != "" {
		metadata[breakingKey] = breakingChanges
	}
	return CommitMessage{Type: ctype, Scope: scope, Description: description, Body: body, IsBreakingChange: breakingChanges != "", Metadata: metadata}
}

// Issue return issue from metadata.
func (m CommitMessage) Issue() string {
	return m.Metadata[issueKey]
}

// BreakingMessage return breaking change message from metadata.
func (m CommitMessage) BreakingMessage() string {
	return m.Metadata[breakingKey]
}

// MessageProcessor interface.
type MessageProcessor interface {
	SkipBranch(branch string) bool
	Validate(message string) error
	Enhance(branch string, message string) (string, error)
	IssueID(branch string) (string, error)
	Format(msg CommitMessage) (string, string, string)
	Parse(subject, body string) CommitMessage
}

// NewMessageProcessor MessageProcessorImpl constructor
func NewMessageProcessor(mcfg CommitMessageConfig, bcfg BranchesConfig) *MessageProcessorImpl {
	return &MessageProcessorImpl{
		messageCfg:  mcfg,
		branchesCfg: bcfg,
	}
}

// MessageProcessorImpl process validate message hook.
type MessageProcessorImpl struct {
	messageCfg  CommitMessageConfig
	branchesCfg BranchesConfig
}

// SkipBranch check if branch should be ignored.
func (p MessageProcessorImpl) SkipBranch(branch string) bool {
	return contains(branch, p.branchesCfg.Skip)
}

// Validate commit message.
func (p MessageProcessorImpl) Validate(message string) error {
	valid, err := regexp.MatchString("^("+strings.Join(p.messageCfg.Types, "|")+")(\\(.+\\))?!?: .*$", firstLine(message))
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("message should contain type: %v, and should be valid according with conventional commits", p.messageCfg.Types)
	}
	return nil
}

// Enhance add metadata on commit message.
func (p MessageProcessorImpl) Enhance(branch string, message string) (string, error) {
	if p.branchesCfg.DisableIssue || p.messageCfg.IssueConfig().Key == "" || hasIssueID(message, p.messageCfg.IssueConfig().Key) {
		return "", nil //enhance disabled
	}

	issue, err := p.IssueID(branch)
	if err != nil {
		return "", err
	}
	if issue == "" {
		return "", fmt.Errorf("could not find issue id using configured regex")
	}

	footer := fmt.Sprintf("%s: %s", p.messageCfg.IssueConfig().Key, issue)

	if !hasFooter(message, p.messageCfg.Footer[breakingKey].Key) {
		return "\n" + footer, nil
	}

	return footer, nil
}

// IssueID try to extract issue id from branch, return empty if not found.
func (p MessageProcessorImpl) IssueID(branch string) (string, error) {
	rstr := fmt.Sprintf("^%s(%s)%s$", p.branchesCfg.PrefixRegex, p.messageCfg.Issue.Regex, p.branchesCfg.SuffixRegex)
	r, err := regexp.Compile(rstr)
	if err != nil {
		return "", fmt.Errorf("could not compile issue regex: %s, error: %v", rstr, err.Error())
	}

	groups := r.FindStringSubmatch(branch)
	if len(groups) != 4 {
		return "", nil
	}
	return groups[2], nil
}

// Format a commit message returning header, body and footer.
func (p MessageProcessorImpl) Format(msg CommitMessage) (string, string, string) {
	var header strings.Builder
	header.WriteString(msg.Type)
	if msg.Scope != "" {
		header.WriteString("(" + msg.Scope + ")")
	}
	header.WriteString(": ")
	header.WriteString(msg.Description)

	var footer strings.Builder
	if msg.BreakingMessage() != "" {
		footer.WriteString(fmt.Sprintf("%s: %s", p.messageCfg.BreakingChangeConfig().Key, msg.BreakingMessage()))
	}
	if issue, exists := msg.Metadata[issueKey]; exists {
		if footer.Len() > 0 {
			footer.WriteString("\n")
		}
		footer.WriteString(fmt.Sprintf("%s: %s", p.messageCfg.IssueConfig().Key, issue))
	}

	return header.String(), msg.Body, footer.String()
}

// Parse a commit message.
func (p MessageProcessorImpl) Parse(subject, body string) CommitMessage {
	commitType, scope, description, hasBreakingChange := parseSubjectMessage(subject)

	metadata := make(map[string]string)
	for key, mdCfg := range p.messageCfg.Footer {
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

func hasFooter(message, breakingChangeKey string) bool {
	r := regexp.MustCompile("^[a-zA-Z-]+: .*|^[a-zA-Z-]+ #.*|^" + breakingChangeKey + ": .*")

	scanner := bufio.NewScanner(strings.NewReader(message))
	lines := 0
	for scanner.Scan() {
		if lines > 0 && r.MatchString(scanner.Text()) {
			return true
		}
		lines++
	}

	return false
}

func hasIssueID(message, issueKeyName string) bool {
	r := regexp.MustCompile(fmt.Sprintf("(?m)^%s: .+$", issueKeyName))
	return r.MatchString(message)
}

func contains(value string, content []string) bool {
	for _, v := range content {
		if value == v {
			return true
		}
	}
	return false
}

func firstLine(value string) string {
	return strings.Split(value, "\n")[0]
}
