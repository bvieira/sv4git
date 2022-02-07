## {{if .Release}}{{.Release}}{{end}}{{if and .Date .Release}} ({{end}}{{.Date}}{{if and .Date .Release}}){{end}}
{{- range $section := .Sections }}
{{- template "rn-md-section.tpl" $section }}
{{- end}}
{{- template "rn-md-section-breaking-changes.tpl" .BreakingChanges}}
