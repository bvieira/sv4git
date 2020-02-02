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

const rnSectionItem = "- {{if .Scope}}**{{.Scope}}:** {{end}}{{.Subject}} ({{.Hash}}){{if .Metadata.issueid}} ({{.Metadata.issueid}}){{end}}"
const rnSection = `{{- if .}}

### {{.Name}}
{{range $k,$v := .Items}}
{{template "rnSectionItem" $v}}
{{- end}}
{{- end}}`
const rnSectionBreakingChanges = `{{- if .}}

### Breaking Changes
{{range $k,$v := .}}
- {{$v}}
{{- end}}
{{- end}}`
const rnTemplate = `## v{{.Version}} ({{.Date}})
{{- template "rnSection" .Sections.feat}}
{{- template "rnSection" .Sections.fix}}
{{- template "rnSectionBreakingChanges" .BreakingChanges}}
`

// OutputFormatter output formatter interface.
type OutputFormatter interface {
	FormatReleaseNote(releasenote ReleaseNote) string
}

// OutputFormatterImpl formater for release note and changelog.
type OutputFormatterImpl struct {
	releasenoteTemplate *template.Template
}

// NewOutputFormatter TemplateProcessor constructor.
func NewOutputFormatter() *OutputFormatterImpl {
	t := template.Must(template.New("releasenotes").Parse(rnTemplate))
	template.Must(t.New("rnSectionItem").Parse(rnSectionItem))
	template.Must(t.New("rnSection").Parse(rnSection))
	template.Must(t.New("rnSectionBreakingChanges").Parse(rnSectionBreakingChanges))
	return &OutputFormatterImpl{releasenoteTemplate: t}
}

// FormatReleaseNote format a release note.
func (p OutputFormatterImpl) FormatReleaseNote(releasenote ReleaseNote) string {
	templateVars := releaseNoteTemplateVariables{
		Version:         fmt.Sprintf("%d.%d.%d", releasenote.Version.Major(), releasenote.Version.Minor(), releasenote.Version.Patch()),
		Date:            releasenote.Date.Format("2006-01-02"),
		Sections:        releasenote.Sections,
		BreakingChanges: releasenote.BreakingChanges,
	}

	var b bytes.Buffer
	p.releasenoteTemplate.Execute(&b, templateVars)
	return b.String()
}
