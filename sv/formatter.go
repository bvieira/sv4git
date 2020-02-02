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

const rnTemplate = `## v{{.Version}} ({{.Date}})
{{- if .Sections.feat}}

### {{.Sections.feat.Name}}
{{range $k,$v := .Sections.feat.Items}}
- {{if $v.Scope}}**{{$v.Scope}}:** {{end}}{{$v.Subject}} ({{$v.Hash}}){{if $v.Metadata.issueid}} ({{$v.Metadata.issueid}}){{end}}{{end}}
{{- end}}

{{- if .Sections.fix}}

### {{.Sections.fix.Name}}
{{range $k,$v := .Sections.fix.Items}}
- {{if $v.Scope}}**{{$v.Scope}}:** {{end}}{{$v.Subject}} ({{$v.Hash}}){{if $v.Metadata.issueid}} ({{$v.Metadata.issueid}}){{end}}{{end}}
{{- end}}

{{- if .BreakingChanges}}

### Breaking Changes
{{range $k,$v := .BreakingChanges}}
- {{$v}}{{end}}
{{- end}}
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
	template := template.Must(template.New("releasenotes").Parse(rnTemplate))
	return &OutputFormatterImpl{releasenoteTemplate: template}
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
