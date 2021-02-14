package sv

// ==== Message ====

// CommitMessageConfig config a commit message.
type CommitMessageConfig struct {
	Types  []string                             `yaml:"types"`
	Scope  CommitMessageScopeConfig             `yaml:"scope"`
	Footer map[string]CommitMessageFooterConfig `yaml:"footer"`
	Issue  CommitMessageIssueConfig             `yaml:"issue"`
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
	Mandatory bool     `yaml:"mandatory"`
	Values    []string `yaml:"values"`
}

// CommitMessageFooterConfig config footer metadata.
type CommitMessageFooterConfig struct {
	Key         string   `yaml:"key"`
	KeySynonyms []string `yaml:"key-synonyms"`
	UseHash     bool     `yaml:"use-hash"`
}

// CommitMessageIssueConfig issue preferences.
type CommitMessageIssueConfig struct {
	Regex string `yaml:"regex"`
}

// ==== Branches ====

// BranchesConfig branches preferences.
type BranchesConfig struct {
	PrefixRegex  string   `yaml:"prefix"`
	SuffixRegex  string   `yaml:"sufix"`
	DisableIssue bool     `yaml:"disable-issue"`
	Skip         []string `yaml:"skip"`
}

// ==== Versioning ====

// VersioningConfig versioning preferences.
type VersioningConfig struct {
	UpdateMajor   []string `yaml:"update-major"`
	UpdateMinor   []string `yaml:"update-minor"`
	UpdatePatch   []string `yaml:"update-patch"`
	IgnoreUnknown bool     `yaml:"ignore-unknown"`
}

// ==== Tag ====

// TagConfig tag preferences.
type TagConfig struct {
	Pattern string `yaml:"pattern"`
}

// ==== Release Notes ====

// ReleaseNotesConfig release notes preferences.
type ReleaseNotesConfig struct {
	Headers map[string]string `yaml:"headers"`
}
