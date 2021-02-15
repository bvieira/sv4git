package sv

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	breakingChangeFooterKey   = "BREAKING CHANGE"
	breakingChangeMetadataKey = "breaking-change"
	issueMetadataKey          = "issue"
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
		metadata[issueMetadataKey] = issue
	}
	if breakingChanges != "" {
		metadata[breakingChangeMetadataKey] = breakingChanges
	}
	return CommitMessage{Type: ctype, Scope: scope, Description: description, Body: body, IsBreakingChange: breakingChanges != "", Metadata: metadata}
}

// Issue return issue from metadata.
func (m CommitMessage) Issue() string {
	return m.Metadata[issueMetadataKey]
}

// BreakingMessage return breaking change message from metadata.
func (m CommitMessage) BreakingMessage() string {
	return m.Metadata[breakingChangeMetadataKey]
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
	subject, body := splitCommitMessageContent(message)
	msg := p.Parse(subject, body)

	if !regexp.MustCompile("^[a-z+]+(\\(.+\\))?!?: .+$").MatchString(subject) {
		return errors.New("message should be valid according with conventional commits")
	}

	if msg.Type == "" || !contains(msg.Type, p.messageCfg.Types) {
		return fmt.Errorf("message type should be one of [%v]", strings.Join(p.messageCfg.Types, ", "))
	}

	if len(p.messageCfg.Scope.Values) > 0 && !contains(msg.Scope, p.messageCfg.Scope.Values) {
		return fmt.Errorf("message scope should one of [%v]", strings.Join(p.messageCfg.Scope.Values, ", "))
	}

	return nil
}

// Enhance add metadata on commit message.
func (p MessageProcessorImpl) Enhance(branch string, message string) (string, error) {
	if p.branchesCfg.DisableIssue || p.messageCfg.IssueFooterConfig().Key == "" || hasIssueID(message, p.messageCfg.IssueFooterConfig()) {
		return "", nil //enhance disabled
	}

	issue, err := p.IssueID(branch)
	if err != nil {
		return "", err
	}
	if issue == "" {
		return "", fmt.Errorf("could not find issue id using configured regex")
	}

	footer := fmt.Sprintf("%s: %s", p.messageCfg.IssueFooterConfig().Key, issue)

	if !hasFooter(message) {
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
		footer.WriteString(fmt.Sprintf("%s: %s", breakingChangeFooterKey, msg.BreakingMessage()))
	}
	if issue, exists := msg.Metadata[issueMetadataKey]; exists {
		if footer.Len() > 0 {
			footer.WriteString("\n")
		}
		if p.messageCfg.IssueFooterConfig().UseHash {
			footer.WriteString(fmt.Sprintf("%s #%s", p.messageCfg.IssueFooterConfig().Key, strings.TrimPrefix(issue, "#")))
		} else {
			footer.WriteString(fmt.Sprintf("%s: %s", p.messageCfg.IssueFooterConfig().Key, issue))
		}
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
	if tagValue := extractFooterMetadata(breakingChangeFooterKey, body, false); tagValue != "" {
		metadata[breakingChangeMetadataKey] = tagValue
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

func hasFooter(message string) bool {
	r := regexp.MustCompile("^[a-zA-Z-]+: .*|^[a-zA-Z-]+ #.*|^" + breakingChangeFooterKey + ": .*")

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

func hasIssueID(message string, issueConfig CommitMessageFooterConfig) bool {
	var r *regexp.Regexp
	if issueConfig.UseHash {
		r = regexp.MustCompile(fmt.Sprintf("(?m)^%s #.+$", issueConfig.Key))
	} else {
		r = regexp.MustCompile(fmt.Sprintf("(?m)^%s: .+$", issueConfig.Key))
	}
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

func splitCommitMessageContent(content string) (string, string) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	scanner.Scan()
	subject := scanner.Text()

	var body strings.Builder
	first := true
	for scanner.Scan() {
		if !first {
			body.WriteString("\n")
		}
		body.WriteString(scanner.Text())
		first = false
	}

	return subject, body.String()
}
