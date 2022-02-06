## {{if .Release}}{{.Release}}{{end}}{{if and .Date .Release}} ({{end}}{{.Date}}{{if and .Date .Release}}){{end}}
{{- $sections := .Sections }}
{{- range $key := .Order }}
{{- template "rn-md-section.tpl" (index $sections $key) }}
{{- end}}
{{- template "rn-md-section-breaking-changes.tpl" .BreakingChanges}}
