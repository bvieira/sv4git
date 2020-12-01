package sv

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// ValidateMessageProcessor interface.
type ValidateMessageProcessor interface {
	SkipBranch(branch string) bool
	Validate(message string) error
	Enhance(branch string, message string) (string, error)
}

// NewValidateMessageProcessor ValidateMessageProcessorImpl constructor
func NewValidateMessageProcessor(skipBranches, supportedTypes []string, issueKeyName, branchIssueRegex string) *ValidateMessageProcessorImpl {
	return &ValidateMessageProcessorImpl{
		skipBranches:     skipBranches,
		supportedTypes:   supportedTypes,
		issueKeyName:     issueKeyName,
		branchIssueRegex: branchIssueRegex,
	}
}

// ValidateMessageProcessorImpl process validate message hook.
type ValidateMessageProcessorImpl struct {
	skipBranches     []string
	supportedTypes   []string
	issueKeyName     string
	branchIssueRegex string
}

// SkipBranch check if branch should be ignored.
func (p ValidateMessageProcessorImpl) SkipBranch(branch string) bool {
	return contains(branch, p.skipBranches)
}

// Validate commit message.
func (p ValidateMessageProcessorImpl) Validate(message string) error {
	valid, err := regexp.MatchString("^("+strings.Join(p.supportedTypes, "|")+")(\\(.+\\))?!?: .*$", firstLine(message))
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("message should contain type: %v, and should be valid according with conventional commits", p.supportedTypes)
	}
	return nil
}

// Enhance add metadata on commit message.
func (p ValidateMessageProcessorImpl) Enhance(branch string, message string) (string, error) {
	if p.branchIssueRegex == "" || p.issueKeyName == "" || hasIssueID(message, p.issueKeyName) {
		return "", nil //enhance disabled
	}

	r, err := regexp.Compile(p.branchIssueRegex)
	if err != nil {
		return "", fmt.Errorf("could not compile issue regex: %s, error: %v", p.branchIssueRegex, err.Error())
	}

	groups := r.FindStringSubmatch(branch)
	if len(groups) != 4 {
		return "", fmt.Errorf("could not find issue id using configured regex")
	}

	footer := fmt.Sprintf("%s: %s", p.issueKeyName, groups[2])

	if !hasFooter(message) {
		return "\n" + footer, nil
	}

	return footer, nil
}

func hasFooter(message string) bool {
	r := regexp.MustCompile("^[a-zA-Z-]+: .*|^[a-zA-Z-]+ #.*|^BREAKING CHANGE: .*")

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
