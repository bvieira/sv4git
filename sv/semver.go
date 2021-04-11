package sv

import "github.com/Masterminds/semver/v3"

type versionType int

const (
	none versionType = iota
	patch
	minor
	major
)

// ToVersion parse string to semver.Version
func ToVersion(value string) (semver.Version, error) {
	version := value
	if version == "" {
		version = "0.0.0"
	}
	v, err := semver.NewVersion(version)
	if err != nil {
		return semver.Version{}, err
	}
	return *v, nil
}

// SemVerCommitsProcessor interface
type SemVerCommitsProcessor interface {
	NextVersion(version semver.Version, commits []GitCommitLog) (semver.Version, bool)
}

// SemVerCommitsProcessorImpl process versions using commit log
type SemVerCommitsProcessorImpl struct {
	MajorVersionTypes         map[string]struct{}
	MinorVersionTypes         map[string]struct{}
	PatchVersionTypes         map[string]struct{}
	KnownTypes                []string
	IncludeUnknownTypeAsPatch bool
}

// NewSemVerCommitsProcessor SemanticVersionCommitsProcessorImpl constructor
func NewSemVerCommitsProcessor(vcfg VersioningConfig, mcfg CommitMessageConfig) *SemVerCommitsProcessorImpl {
	return &SemVerCommitsProcessorImpl{
		IncludeUnknownTypeAsPatch: !vcfg.IgnoreUnknown,
		MajorVersionTypes:         toMap(vcfg.UpdateMajor),
		MinorVersionTypes:         toMap(vcfg.UpdateMinor),
		PatchVersionTypes:         toMap(vcfg.UpdatePatch),
		KnownTypes:                mcfg.Types,
	}
}

// NextVersion calculates next version based on commit log
func (p SemVerCommitsProcessorImpl) NextVersion(version semver.Version, commits []GitCommitLog) (semver.Version, bool) {
	var versionToUpdate = none
	for _, commit := range commits {
		if v := p.versionTypeToUpdate(commit); v > versionToUpdate {
			versionToUpdate = v
		}
	}

	switch versionToUpdate {
	case major:
		return version.IncMajor(), true
	case minor:
		return version.IncMinor(), true
	case patch:
		return version.IncPatch(), true
	default:
		return version, false
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
