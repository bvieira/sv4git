package sv

import (
	"bytes"
	"io/fs"
	"sort"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
)

type releaseNoteTemplateVariables struct {
	Release     string
	Tag         string
	Version     *semver.Version
	Date        time.Time
	Sections    []ReleaseNoteSection
	AuthorNames []string
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
	templateFNs := map[string]interface{}{
		"timefmt": timeFormat,
	}
	tpls := template.Must(template.New("templates").Funcs(templateFNs).ParseFS(templatesFS, "*"))
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
	release := releasenote.Tag
	if releasenote.Version != nil {
		release = "v" + releasenote.Version.String()
	}
	return releaseNoteTemplateVariables{
		Release:     release,
		Tag:         releasenote.Tag,
		Version:     releasenote.Version,
		Date:        releasenote.Date,
		Sections:    releasenote.Sections,
		AuthorNames: toSortedArray(releasenote.AuthorsNames),
	}
}

func toSortedArray(input map[string]struct{}) []string {
	result := make([]string, len(input))
	i := 0
	for k := range input {
		result[i] = k
		i++
	}
	sort.Strings(result)
	return result
}

func timeFormat(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(format)
}
