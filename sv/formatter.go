package sv

import (
	"bytes"
	"fmt"
	"text/template"
)

type releaseNoteTemplateVariables struct {
	Version         string
	Date            string
	Sections        map[string]ReleaseNoteSection
	BreakingChanges []string
}

const (
	cglTemplate = `# Changelog
{{- range .}}

{{template "rnTemplate" .}}
---
{{- end}}
`

	rnSectionItem = "- {{if .Scope}}**{{.Scope}}:** {{end}}{{.Subject}} ({{.Hash}}){{if .Metadata.issueid}} ({{.Metadata.issueid}}){{end}}"

	rnSection = `{{- if .}}

### {{.Name}}
{{range $k,$v := .Items}}
{{template "rnSectionItem" $v}}
{{- end}}
{{- end}}`

	rnSectionBreakingChanges = `{{- if .}}

### Breaking Changes
{{range $k,$v := .}}
- {{$v}}
{{- end}}
{{- end}}`

	rnTemplate = `## v{{.Version}}{{if .Date}} ({{.Date}}){{end}}
{{- template "rnSection" .Sections.feat}}
{{- template "rnSection" .Sections.fix}}
{{- template "rnSectionBreakingChanges" .BreakingChanges}}
`
)

// OutputFormatter output formatter interface.
type OutputFormatter interface {
	FormatReleaseNote(releasenote ReleaseNote) string
	FormatChangelog(releasenotes []ReleaseNote) string
}

// OutputFormatterImpl formater for release note and changelog.
type OutputFormatterImpl struct {
	releasenoteTemplate *template.Template
	changelogTemplate   *template.Template
}

// NewOutputFormatter TemplateProcessor constructor.
func NewOutputFormatter() *OutputFormatterImpl {
	cgl := template.Must(template.New("cglTemplate").Parse(cglTemplate))
	rn := template.Must(cgl.New("rnTemplate").Parse(rnTemplate))
	template.Must(rn.New("rnSectionItem").Parse(rnSectionItem))
	template.Must(rn.New("rnSection").Parse(rnSection))
	template.Must(rn.New("rnSectionBreakingChanges").Parse(rnSectionBreakingChanges))
	return &OutputFormatterImpl{releasenoteTemplate: rn, changelogTemplate: cgl}
}

// FormatReleaseNote format a release note.
func (p OutputFormatterImpl) FormatReleaseNote(releasenote ReleaseNote) string {
	var b bytes.Buffer
	p.releasenoteTemplate.Execute(&b, releaseNoteVariables(releasenote))
	return b.String()
}

// FormatChangelog format a changelog
func (p OutputFormatterImpl) FormatChangelog(releasenotes []ReleaseNote) string {
	var templateVars []releaseNoteTemplateVariables
	for _, v := range releasenotes {
		templateVars = append(templateVars, releaseNoteVariables(v))
	}

	var b bytes.Buffer
	p.changelogTemplate.Execute(&b, templateVars)
	return b.String()
}

func releaseNoteVariables(releasenote ReleaseNote) releaseNoteTemplateVariables {
	var date = ""
	if !releasenote.Date.IsZero() {
		date = releasenote.Date.Format("2006-01-02")
	}
	return releaseNoteTemplateVariables{
		Version:         fmt.Sprintf("%d.%d.%d", releasenote.Version.Major(), releasenote.Version.Minor(), releasenote.Version.Patch()),
		Date:            date,
		Sections:        releasenote.Sections,
		BreakingChanges: releasenote.BreakingChanges,
	}
}
