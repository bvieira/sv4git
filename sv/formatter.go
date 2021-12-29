package sv

import (
	"bytes"
	"text/template"
)

type releaseNoteTemplateVariables struct {
	Release         string
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

	rnTemplate = `## {{if .Release}}{{.Release}}{{end}}{{if and .Date .Release}} ({{end}}{{.Date}}{{if and .Date .Release}}){{end}}
{{- $sections := .Sections }}
{{- range $key := .Order }}
{{- template "rnSection" (index $sections $key) }}
{{- end}}
{{- template "rnSectionBreakingChanges" .BreakingChanges}}
`
)

// OutputFormatter output formatter interface.
type OutputFormatter interface {
	FormatReleaseNote(releasenote ReleaseNote) (string, error)
	FormatChangelog(releasenotes []ReleaseNote) (string, error)
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
func (p OutputFormatterImpl) FormatReleaseNote(releasenote ReleaseNote) (string, error) {
	var b bytes.Buffer
	if err := p.releasenoteTemplate.Execute(&b, releaseNoteVariables(releasenote)); err != nil {
		return "", err
	}
	return b.String(), nil
}

// FormatChangelog format a changelog.
func (p OutputFormatterImpl) FormatChangelog(releasenotes []ReleaseNote) (string, error) {
	templateVars := make([]releaseNoteTemplateVariables, len(releasenotes))
	for i, v := range releasenotes {
		templateVars[i] = releaseNoteVariables(v)
	}

	var b bytes.Buffer
	if err := p.changelogTemplate.Execute(&b, templateVars); err != nil {
		return "", err
	}
	return b.String(), nil
}

func releaseNoteVariables(releasenote ReleaseNote) releaseNoteTemplateVariables {
	date := ""
	if !releasenote.Date.IsZero() {
		date = releasenote.Date.Format("2006-01-02")
	}

	release := ""
	if releasenote.Version != nil {
		release = "v" + releasenote.Version.String()
	} else if releasenote.Tag != "" {
		release = releasenote.Tag
	}
	return releaseNoteTemplateVariables{
		Release:         release,
		Date:            date,
		Sections:        releasenote.Sections,
		Order:           []string{"feat", "fix", "refactor", "perf", "test", "build", "ci", "chore", "docs", "style"},
		BreakingChanges: releasenote.BreakingChanges,
	}
}
