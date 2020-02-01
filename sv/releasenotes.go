package sv

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/Masterminds/semver"
)

type releaseNoteTemplate struct {
	Version         string
	Date            string
	Sections        map[string]ReleaseNoteSection
	BreakingChanges []string
}

const markdownTemplate = `# v{{.Version}} ({{.Date}})

{{if .Sections.feat}}## {{.Sections.feat.Name}}
{{range $k,$v := .Sections.feat.Items}}
- {{if $v.Scope}}**{{$v.Scope}}:** {{end}}{{$v.Subject}} ({{$v.Hash}}) {{if $v.Metadata.issueid}}({{$v.Metadata.issueid}}){{end}}{{end}}{{end}}

{{if .Sections.fix}}## {{.Sections.fix.Name}}
{{range $k,$v := .Sections.fix.Items}}
- {{if $v.Scope}}**{{$v.Scope}}:** {{end}}{{$v.Subject}} ({{$v.Hash}}) {{if $v.Metadata.issueid}}({{$v.Metadata.issueid}}){{end}}{{end}}{{end}}

{{if .BreakingChanges}}## Breaking Changes
{{range $k,$v := .BreakingChanges}}
- {{$v}}{{end}}
{{end}}`

// ReleaseNoteProcessor release note processor interface.
type ReleaseNoteProcessor interface {
	Get(date time.Time, commits []GitCommitLog) ReleaseNote
	Format(releasenote ReleaseNote, version semver.Version) string
}

// ReleaseNoteProcessorImpl release note based on commit log.
type ReleaseNoteProcessorImpl struct {
	tags     map[string]string
	template *template.Template
}

// NewReleaseNoteProcessor ReleaseNoteProcessor constructor.
func NewReleaseNoteProcessor(tags map[string]string) *ReleaseNoteProcessorImpl {
	template := template.Must(template.New("markdown").Parse(markdownTemplate))
	return &ReleaseNoteProcessorImpl{tags: tags, template: template}
}

// Get generate a release note based on commits.
func (p ReleaseNoteProcessorImpl) Get(date time.Time, commits []GitCommitLog) ReleaseNote {
	sections := make(map[string]ReleaseNoteSection)
	var breakingChanges []string
	for _, commit := range commits {
		if name, exists := p.tags[commit.Type]; exists {
			section, sexists := sections[commit.Type]
			if !sexists {
				section = ReleaseNoteSection{Name: name}
			}
			section.Items = append(section.Items, commit)
			sections[commit.Type] = section
		}
		if value, exists := commit.Metadata[BreakingChangeTag]; exists {
			breakingChanges = append(breakingChanges, value)
		}
	}

	return ReleaseNote{Date: date.Truncate(time.Minute), Sections: sections, BreakingChanges: breakingChanges}
}

// Format format a release note.
func (p ReleaseNoteProcessorImpl) Format(releasenote ReleaseNote, version semver.Version) string {
	templateVars := releaseNoteTemplate{
		Version:         fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), version.Patch()),
		Date:            releasenote.Date.Format("2006-01-02"),
		Sections:        releasenote.Sections,
		BreakingChanges: releasenote.BreakingChanges,
	}

	var b bytes.Buffer
	p.template.Execute(&b, templateVars)
	return b.String()
}

// ReleaseNote release note.
type ReleaseNote struct {
	Date            time.Time
	Sections        map[string]ReleaseNoteSection
	BreakingChanges []string
}

// ReleaseNoteSection release note section.
type ReleaseNoteSection struct {
	Name  string
	Items []GitCommitLog
}
