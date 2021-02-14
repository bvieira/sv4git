package sv

// CommitMessageConfig config a commit message.
type CommitMessageConfig struct {
	Types  []string
	Scope  CommitMessageScopeConfig
	Footer map[string]CommitMessageFooterConfig
	Issue  CommitMessageIssueConfig
}

// IssueConfig config for issue.
func (c CommitMessageConfig) IssueConfig() CommitMessageFooterConfig {
	if v, exists := c.Footer[issueKey]; exists {
		return v
	}
	return CommitMessageFooterConfig{}
}

// BreakingChangeConfig config for breaking changes.
func (c CommitMessageConfig) BreakingChangeConfig() CommitMessageFooterConfig {
	if v, exists := c.Footer[breakingKey]; exists {
		return v
	}
	return CommitMessageFooterConfig{}
}

// CommitMessageScopeConfig config scope preferences.
type CommitMessageScopeConfig struct {
	Mandatory bool
	Values    []string
}

// CommitMessageFooterConfig config footer metadata.
type CommitMessageFooterConfig struct {
	Key         string
	KeySynonyms []string
	UseHash     bool
}

// CommitMessageIssueConfig issue preferences.
type CommitMessageIssueConfig struct {
	Regex string
}

// BranchesConfig branches preferences.
type BranchesConfig struct {
	PrefixRegex string
	SuffixRegex string
	ExpectIssue bool
	Skip        []string
}
