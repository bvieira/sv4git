package sv

// ==== Message ====

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

// ==== Branches ====

// BranchesConfig branches preferences.
type BranchesConfig struct {
	PrefixRegex string
	SuffixRegex string
	ExpectIssue bool
	Skip        []string
}

// ==== Versioning ====

// VersioningConfig versioning preferences.
type VersioningConfig struct {
	UpdateMajor        []string
	UpdateMinor        []string
	UpdatePatch        []string
	UnknownTypeAsPatch bool
}

// ==== Tag ====

// TagConfig tag preferences.
type TagConfig struct {
	Pattern string
}

// ==== Release Notes ====

// ReleaseNotesConfig release notes preferences.
type ReleaseNotesConfig struct {
	Headers map[string]string
}
