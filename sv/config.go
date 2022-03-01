package sv

// ==== Message ====

// CommitMessageConfig config a commit message.
type CommitMessageConfig struct {
	Types  []string                             `yaml:"types,flow"`
	Scope  CommitMessageScopeConfig             `yaml:"scope"`
	Footer map[string]CommitMessageFooterConfig `yaml:"footer"`
	Issue  CommitMessageIssueConfig             `yaml:"issue"`
}

// IssueFooterConfig config for issue.
func (c CommitMessageConfig) IssueFooterConfig() CommitMessageFooterConfig {
	if v, exists := c.Footer[issueMetadataKey]; exists {
		return v
	}
	return CommitMessageFooterConfig{}
}

// CommitMessageScopeConfig config scope preferences.
type CommitMessageScopeConfig struct {
	Values []string `yaml:"values"`
}

// CommitMessageFooterConfig config footer metadata.
type CommitMessageFooterConfig struct {
	Key            string   `yaml:"key"`
	KeySynonyms    []string `yaml:"key-synonyms,flow"`
	UseHash        bool     `yaml:"use-hash"`
	AddValuePrefix string   `yaml:"add-value-prefix"`
}

// CommitMessageIssueConfig issue preferences.
type CommitMessageIssueConfig struct {
	Regex string `yaml:"regex"`
}

// ==== Branches ====

// BranchesConfig branches preferences.
type BranchesConfig struct {
	Prefix       string   `yaml:"prefix"`
	Suffix       string   `yaml:"suffix"`
	DisableIssue bool     `yaml:"disable-issue"`
	Skip         []string `yaml:"skip,flow"`
	SkipDetached *bool    `yaml:"skip-detached"`
}

// ==== Versioning ====

// VersioningConfig versioning preferences.
type VersioningConfig struct {
	UpdateMajor   []string `yaml:"update-major,flow"`
	UpdateMinor   []string `yaml:"update-minor,flow"`
	UpdatePatch   []string `yaml:"update-patch,flow"`
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
	Headers  map[string]string           `yaml:"headers,omitempty"`
	Sections []ReleaseNotesSectionConfig `yaml:"sections"`
}

func (cfg ReleaseNotesConfig) sectionConfig(sectionType string) *ReleaseNotesSectionConfig {
	for _, sectionCfg := range cfg.Sections {
		if sectionCfg.SectionType == sectionType {
			return &sectionCfg
		}
	}
	return nil
}

// ReleaseNotesSectionConfig preferences for a single section on release notes.
type ReleaseNotesSectionConfig struct {
	Name        string   `yaml:"name"`
	SectionType string   `yaml:"section-type"`
	CommitTypes []string `yaml:"commit-types,flow,omitempty"`
}

const (
	// ReleaseNotesSectionTypeCommits ReleaseNotesSectionConfig.SectionType value.
	ReleaseNotesSectionTypeCommits = "commits"
	// ReleaseNotesSectionTypeBreakingChanges ReleaseNotesSectionConfig.SectionType value.
	ReleaseNotesSectionTypeBreakingChanges = "breaking-changes"
)
