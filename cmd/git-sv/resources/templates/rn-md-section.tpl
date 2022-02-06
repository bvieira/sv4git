{{- if .}}{{- if ne .Name ""}}

### {{.Name}}
{{range $k,$v := .Items}}
{{template "rn-md-section-item.tpl" $v}}
{{- end}}
{{- end}}{{- end}}