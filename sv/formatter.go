package sv

import (
	"bytes"
	"io/fs"
	"text/template"
)

type releaseNoteTemplateVariables struct {
	Release         string
	Date            string
	Sections        map[string]ReleaseNoteSection
	Order           []string
	BreakingChanges BreakingChangeSection
}

// OutputFormatter output formatter interface.
type OutputFormatter interface {
	FormatReleaseNote(releasenote ReleaseNote) (string, error)
	FormatChangelog(releasenotes []ReleaseNote) (string, error)
}

// OutputFormatterImpl formater for release note and changelog.
type OutputFormatterImpl struct {
	templates *template.Template
}

// NewOutputFormatter TemplateProcessor constructor.
func NewOutputFormatter(templatesFS fs.FS) *OutputFormatterImpl {
	tpls := template.Must(template.New("templates").ParseFS(templatesFS, "*"))
	return &OutputFormatterImpl{templates: tpls}
}

// FormatReleaseNote format a release note.
func (p OutputFormatterImpl) FormatReleaseNote(releasenote ReleaseNote) (string, error) {
	var b bytes.Buffer
	if err := p.templates.ExecuteTemplate(&b, "releasenotes-md.tpl", releaseNoteVariables(releasenote)); err != nil {
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
	if err := p.templates.ExecuteTemplate(&b, "changelog-md.tpl", templateVars); err != nil {
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
