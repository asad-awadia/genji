package generator

const querySelectorTmpl = `
{{ define "query-selector" }}

{{ template "query-selector-Struct" . }}
{{ template "query-selector-New" . }}
{{ template "query-selector-Table" . }}
{{ template "query-selector-All" . }}
{{ end }}
`

const querySelectorStructTmpl = `
{{ define "query-selector-Struct" }}
{{- $fl := .FirstLetter -}}
{{- $structName := .Name -}}
// {{$structName}}QuerySelector provides helpers for selecting fields from the {{$structName}} structure.
type {{$structName}}QuerySelector struct{
{{- range $i, $a := .Fields }}
	{{$a.Name}} query.{{.Type}}Field
{{- end}}
}
{{ end }}
`

const querySelectorNewTmpl = `
{{ define "query-selector-New" }}
{{- $fl := .FirstLetter -}}
{{- $structName := .Name -}}
{{- if .IsExported }}
// New{{$structName}}QuerySelector creates a {{$structName}}QuerySelector.
func New{{$structName}}QuerySelector() {{$structName}}QuerySelector {
{{- else}}
// new{{$structName}}QuerySelector creates a {{$structName}}QuerySelector.
func new{{.ExportedName}}QuerySelector() {{$structName}}QuerySelector {
{{- end}}
	return {{$structName}}QuerySelector{
		{{- range $i, $a := .Fields }}
			{{$a.Name}}: query.New{{.Type}}Field("{{$a.Name}}"),
		{{- end}}
	}
}
{{ end }}
`

const querySelectorTableTmpl = `
{{ define "query-selector-Table" }}
{{- $fl := .FirstLetter -}}
{{- $structName := .Name -}}
// Table returns a query.TableSelector for {{$structName}}.
func (*{{$structName}}QuerySelector) Table() query.TableSelector {
	return query.Table("{{$structName}}")
}
{{ end }}
`

const querySelectorAllTmpl = `
{{ define "query-selector-All" }}
{{- $fl := .FirstLetter -}}
{{- $structName := .Name -}}
// All returns a list of all selectors for {{$structName}}.
func (s *{{$structName}}QuerySelector) All() []query.FieldSelector {
	return []query.FieldSelector{
		{{- range $i, $a := .Fields }}
		s.{{$a.Name}},
		{{- end}}
	}
}
{{ end }}
`