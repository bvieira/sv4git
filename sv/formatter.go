package sv

import (
	"bytes"
	"text/template"
)

type releaseNoteTemplateVariables struct {
	Version         string
	Date            string
	Sections        map[string]ReleaseNoteSection
	Order           []string
	BreakingChanges BreakingChangeSection
}

const (
	cglTemplate = `# Changelog
{{- range .}}

{{template "rnTemplate" .}}
---
{{- end}}
`

	rnSectionItem = "- {{if .Message.Scope}}**{{.Message.Scope}}:** {{end}}{{.Message.Description}} ({{.Hash}}){{if .Message.Metadata.issue}} ({{.Message.Metadata.issue}}){{end}}"

	rnSection = `{{- if .}}{{- if ne .Name ""}}

### {{.Name}}
{{range $k,$v := .Items}}
{{template "rnSectionItem" $v}}
{{- end}}
{{- end}}{{- end}}`

	rnSectionBreakingChanges = `{{- if ne .Name ""}}

### {{.Name}}
{{range $k,$v := .Messages}}
- {{$v}}
{{- end}}
{{- end}}`

	rnTemplate = `## {{if .Version}}v{{.Version}}{{end}}{{if and .Date .Version}} ({{end}}{{.Date}}{{if and .Version .Date}}){{end}}
{{- $sections := .Sections }}
{{- range $key := .Order }}
{{- template "rnSection" (index $sections $key) }}
{{- end}}
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

	var version = ""
	if releasenote.Version != nil {
		version = releasenote.Version.String()
	}
	return releaseNoteTemplateVariables{
		Version:         version,
		Date:            date,
		Sections:        releasenote.Sections,
		Order:           []string{"feat", "fix", "refactor", "perf", "test", "build", "ci", "chore", "docs", "style"},
		BreakingChanges: releasenote.BreakingChanges,
	}
}
