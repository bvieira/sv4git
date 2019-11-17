package sv

import "github.com/Masterminds/semver"

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

// BreakingChangeTag breaking change tag from commit metadata
const BreakingChangeTag string = "breakingchange"

// SemVerCommitsProcessor interface
type SemVerCommitsProcessor interface {
	NextVersion(version semver.Version, commits []GitCommitLog) semver.Version
}

// SemVerCommitsProcessorImpl process versions using commit log
type SemVerCommitsProcessorImpl struct {
	MajorVersionTypes         map[string]struct{}
	MinorVersionTypes         map[string]struct{}
	PatchVersionTypes         map[string]struct{}
	IncludeUnknownTypeAsPatch bool
}

// NewSemVerCommitsProcessor SemanticVersionCommitsProcessorImpl constructor
func NewSemVerCommitsProcessor(unknownAsPatch bool, majorTypes, minorTypes, patchTypes []string) *SemVerCommitsProcessorImpl {
	return &SemVerCommitsProcessorImpl{
		IncludeUnknownTypeAsPatch: unknownAsPatch,
		MajorVersionTypes:         toMap(majorTypes),
		MinorVersionTypes:         toMap(minorTypes),
		PatchVersionTypes:         toMap(patchTypes),
	}
}

// NextVersion calculates next version based on commit log
func (p SemVerCommitsProcessorImpl) NextVersion(version semver.Version, commits []GitCommitLog) semver.Version {
	var versionToUpdate = none
	for _, commit := range commits {
		if v := p.versionTypeToUpdate(commit); v > versionToUpdate {
			versionToUpdate = v
		}
	}

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
	if _, exists := commit.Metadata[BreakingChangeTag]; exists {
		return major
	}
	if _, exists := p.MajorVersionTypes[commit.Type]; exists {
		return major
	}
	if _, exists := p.MinorVersionTypes[commit.Type]; exists {
		return minor
	}
	if _, exists := p.PatchVersionTypes[commit.Type]; exists {
		return patch
	}
	if p.IncludeUnknownTypeAsPatch {
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
