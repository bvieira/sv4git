package sv

import (
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
func NewValidateMessageProcessor(skipBranches, supportedTypes []string) *ValidateMessageProcessorImpl {
	return &ValidateMessageProcessorImpl{
		skipBranches:   skipBranches,
		supportedTypes: supportedTypes,
	}
}

// ValidateMessageProcessorImpl process validate message hook.
type ValidateMessageProcessorImpl struct {
	skipBranches   []string
	supportedTypes []string
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
	//TODO add issue id (branch format on varenv)
	return "", nil
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
