package sv

// CommitMessageConfig config a commit message.
type CommitMessageConfig struct {
	Types  []string
	Scope  CommitMessageScopeConfig
	Footer map[string]CommitMessageFooterConfig
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
	Regex       string
	UseHash     bool
}
