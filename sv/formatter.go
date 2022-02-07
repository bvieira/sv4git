package sv

import (
	"bytes"
	"io/fs"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
)

type releaseNoteTemplateVariables struct {
	Release         string
	Version         *semver.Version
	Date            string
	Sections        []ReleaseNoteSection
	BreakingChanges BreakingChangeSection
	AuthorNames     []string
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

	release := releasenote.Tag
	if releasenote.Version != nil {
		release = "v" + releasenote.Version.String()
	}
	return releaseNoteTemplateVariables{
		Release:         release,
		Version:         releasenote.Version,
		Date:            date,
		Sections:        toTemplateSections(releasenote.Sections),
		BreakingChanges: releasenote.BreakingChanges,
		AuthorNames:     toSortedArray(releasenote.AuthorsNames),
	}
}

func toTemplateSections(sections map[string]ReleaseNoteSection) []ReleaseNoteSection {
	result := make([]ReleaseNoteSection, len(sections))
	i := 0
	for _, section := range sections {
		result[i] = section
		i++
	}

	order := map[string]int{"feat": 0, "fix": 1, "refactor": 2, "perf": 3, "test": 4, "build": 5, "ci": 6, "chore": 7, "docs": 8, "style": 9}
	sort.SliceStable(result, func(i, j int) bool {
		priority1, disambiguity1 := priority(result[i].Types, order)
		priority2, disambiguity2 := priority(result[j].Types, order)
		if priority1 == priority2 {
			return disambiguity1 < disambiguity2
		}
		return priority1 < priority2
	})
	return result
}

func priority(types []string, order map[string]int) (int, string) {
	sort.Strings(types)
	p := -1
	for _, t := range types {
		if p == -1 || order[t] < p {
			p = order[t]
		}
	}
	return p, strings.Join(types, "-")
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
