package sv

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

const (
	logSeparator       = "##"
	endLine            = "~~"
	breakingChangesTag = "BREAKING CHANGE:"
	issueIDTag         = "jira:"
)

// Git commands
type Git interface {
	Describe() string
	Log(initialTag, endTag string) ([]GitCommitLog, error)
	Tag(version semver.Version) error
	Tags() ([]GitTag, error)
}

// GitCommitLog description of a single commit log
type GitCommitLog struct {
	Hash     string            `json:"hash,omitempty"`
	Type     string            `json:"type,omitempty"`
	Scope    string            `json:"scope,omitempty"`
	Subject  string            `json:"subject,omitempty"`
	Body     string            `json:"body,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GitTag git tag info
type GitTag struct {
	Name string
	Date time.Time
}

// GitImpl git command implementation
type GitImpl struct {
	messageMetadata map[string]string
	tagPattern      string
}

// NewGit constructor
func NewGit(messageMetadata map[string]string, tagPattern string) *GitImpl {
	return &GitImpl{messageMetadata: messageMetadata, tagPattern: tagPattern}
}

// Describe runs git describe, it no tag found, return empty
func (GitImpl) Describe() string {
	cmd := exec.Command("git", "describe", "--abbrev=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(string(out), "\n"))
}

// Log return git log
func (g GitImpl) Log(initialTag, endTag string) ([]GitCommitLog, error) {
	format := "--pretty=format:\"%h" + logSeparator + "%s" + logSeparator + "%b" + endLine + "\""
	var cmd *exec.Cmd
	if initialTag == "" && endTag == "" {
		cmd = exec.Command("git", "log", format)
	} else if endTag == "" {
		cmd = exec.Command("git", "log", initialTag+"..HEAD", format)
	} else if initialTag == "" {
		cmd = exec.Command("git", "log", endTag, format)
	} else {
		cmd = exec.Command("git", "log", initialTag+".."+endTag, format)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseLogOutput(g.messageMetadata, string(out)), nil
}

// Tag create a git tag
func (g GitImpl) Tag(version semver.Version) error {
	tag := fmt.Sprintf(g.tagPattern, version.Major(), version.Minor(), version.Patch())
	tagMsg := fmt.Sprintf("Version %d.%d.%d", version.Major(), version.Minor(), version.Patch())

	tagCommand := exec.Command("git", "tag", "-a", tag, "-m", tagMsg)
	if err := tagCommand.Run(); err != nil {
		return err
	}

	pushCommand := exec.Command("git", "push", "origin", tag)
	return pushCommand.Run()
}

// Tags list repository tags
func (g GitImpl) Tags() ([]GitTag, error) {
	cmd := exec.Command("git", "tag", "-l", "--format", "%(taggerdate:iso8601)#%(refname:short)")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseTagsOutput(string(out))
}

func parseTagsOutput(input string) ([]GitTag, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var result []GitTag
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			values := strings.Split(line, "#")
			date, err := time.Parse("2006-01-02 15:04:05 -0700", values[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse tag data, message: %v", err)
			}
			result = append(result, GitTag{Name: values[1], Date: date})
		}
	}
	return result, nil
}

func parseLogOutput(messageMetadata map[string]string, log string) []GitCommitLog {
	scanner := bufio.NewScanner(strings.NewReader(log))
	scanner.Split(splitAt([]byte(endLine)))
	var logs []GitCommitLog
	for scanner.Scan() {
		if text := strings.TrimSpace(strings.Trim(scanner.Text(), "\"")); text != "" {
			logs = append(logs, parseCommitLog(messageMetadata, text))
		}
	}
	return logs
}

func parseCommitLog(messageMetadata map[string]string, commit string) GitCommitLog {
	content := strings.Split(strings.Trim(commit, "\""), logSeparator)
	commitType, scope, subject := parseCommitLogMessage(content[1])

	metadata := make(map[string]string)
	for k, v := range messageMetadata {
		if tagValue := extractTag(v, content[2]); tagValue != "" {
			metadata[k] = tagValue
		}
	}

	return GitCommitLog{
		Hash:     content[0],
		Type:     commitType,
		Scope:    scope,
		Subject:  subject,
		Body:     content[2],
		Metadata: metadata,
	}
}

func parseCommitLogMessage(message string) (string, string, string) {
	regex := regexp.MustCompile("([a-z]+)(\\((.*)\\))?: (.*)")
	result := regex.FindStringSubmatch(message)
	if len(result) != 5 {
		return "", "", message
	}
	return result[1], result[3], strings.TrimSpace(result[4])
}

func extractTag(tag, text string) string {
	regex := regexp.MustCompile(tag + ": (.*)")
	result := regex.FindStringSubmatch(text)
	if len(result) < 2 {
		return ""
	}
	return result[1]
}

func splitAt(b []byte) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		dataLen := len(data)

		if atEOF && dataLen == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, b); i >= 0 {
			return i + len(b), data[0:i], nil
		}

		if atEOF {
			return dataLen, data, nil
		}

		return 0, nil, nil
	}
}
