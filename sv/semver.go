package sv

import "github.com/Masterminds/semver/v3"

type versionType int

const (
	none versionType = iota
	patch
	minor
	major
)

// ToVersion parse string to semver.Version.
func ToVersion(value string) (*semver.Version, error) {
	version := value
	if version == "" {
		version = "0.0.0"
	}
	return semver.NewVersion(version)
}

// SemVerCommitsProcessor interface.
type SemVerCommitsProcessor interface {
	NextVersion(version *semver.Version, commits []GitCommitLog) (*semver.Version, bool)
}

// SemVerCommitsProcessorImpl process versions using commit log.
type SemVerCommitsProcessorImpl struct {
	MajorVersionTypes         map[string]struct{}
	MinorVersionTypes         map[string]struct{}
	PatchVersionTypes         map[string]struct{}
	KnownTypes                []string
	IncludeUnknownTypeAsPatch bool
}

// NewSemVerCommitsProcessor SemanticVersionCommitsProcessorImpl constructor.
func NewSemVerCommitsProcessor(vcfg VersioningConfig, mcfg CommitMessageConfig) *SemVerCommitsProcessorImpl {
	return &SemVerCommitsProcessorImpl{
		IncludeUnknownTypeAsPatch: !vcfg.IgnoreUnknown,
		MajorVersionTypes:         toMap(vcfg.UpdateMajor),
		MinorVersionTypes:         toMap(vcfg.UpdateMinor),
		PatchVersionTypes:         toMap(vcfg.UpdatePatch),
		KnownTypes:                mcfg.Types,
	}
}

// NextVersion calculates next version based on commit log.
func (p SemVerCommitsProcessorImpl) NextVersion(version *semver.Version, commits []GitCommitLog) (*semver.Version, bool) {
	versionToUpdate := none
	for _, commit := range commits {
		if v := p.versionTypeToUpdate(commit); v > versionToUpdate {
			versionToUpdate = v
		}
	}

	updated := versionToUpdate != none
	if version == nil {
		return nil, updated
	}
	newVersion := updateVersion(*version, versionToUpdate)
	return &newVersion, updated
}

func updateVersion(version semver.Version, versionToUpdate versionType) semver.Version {
	switch versionToUpdate {
	case major:
		return version.IncMajor()
	case minor:
		return version.IncMinor()
	case patch:
		return version.IncPatch()
	default:
		return version
	}
}

func (p SemVerCommitsProcessorImpl) versionTypeToUpdate(commit GitCommitLog) versionType {
	if commit.Message.IsBreakingChange {
		return major
	}
	if _, exists := p.MajorVersionTypes[commit.Message.Type]; exists {
		return major
	}
	if _, exists := p.MinorVersionTypes[commit.Message.Type]; exists {
		return minor
	}
	if _, exists := p.PatchVersionTypes[commit.Message.Type]; exists {
		return patch
	}
	if !contains(commit.Message.Type, p.KnownTypes) && p.IncludeUnknownTypeAsPatch {
		return patch
	}
	return none
}

func toMap(values []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, v := range values {
		result[v] = struct{}{}
	}
	return result
}
