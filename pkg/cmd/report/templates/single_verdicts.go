package templates

import (
	"io"
	"text/template"

	"github.com/listendev/lstn/pkg/listen"
)

const singleVerdictsTpl = `
{{ if gt (len .Verdicts) 0 }}
<b><a href="https://www.npmjs.com/package/{{ .Name }}/v/{{ .Version }}">{{ .Name }}@{{ .Version }}</a></b><br>
<details>
<summary>:stop_sign:
    <span style="color: red">
        {{ len .Verdicts }} alert{{ if gt (len .Verdicts) 1 }}s{{end}} found
    </span> <i>(click to expand)</i>
</summary>

{{ range .Verdicts }}
{{ if eq .Priority "high" }}
## :stop_sign: {{ .Message }}
{{ else if eq .Priority "medium" }}
## :warning: {{ .Message }}
{{ else if eq .Priority "low" }}
## :large_blue_diamond: {{ .Message }}
{{ end }}
<dl>
<dt>Dependency type</dt>
<dd>
{{ if and (eq (index .Metadata "npm_package_name") $.Name) (eq (index .Metadata "npm_package_version") $.Version) }}
Direct dependency
{{ else }}
Transitive dependency ({{ index .Metadata "npm_package_name" }}@{{ index .Metadata "npm_package_version" }})
{{ end }}
</dd>
{{ if index .Metadata "ai_context" }}
<dt>Context</dt>
<dd>{{ index .Metadata "ai_context" }}</dd>
{{ end }}
{{ if index .Metadata "ai_action" }}
<dt>Action</dt>
<dd>{{ index .Metadata "ai_action" }}</dd>
{{ end }}
<dt>Metadata</dt>
<dd>
<table>
{{ range $key, $value := .Metadata }}
{{ if or (eq $key "npm_package_name")
        (eq $key "npm_package_version")
        (eq $key "ai_comment")
        (eq $key "ai_action")
}}
    {{ continue }}
{{ end }}
<tr>
<td>{{ $key }}:</td><td>{{ $value }}</td>
</tr>
{{ end }}
</table
</dd>
</dl>
{{ end }}

</details>
{{ end }}
`

func RenderSingleVerdictsPackage(w io.Writer, p listen.Package) error {
	ct := template.Must(template.New("single_verdict").Parse(singleVerdictsTpl))
	err := ct.Execute(w, p)
	if err != nil {
		return err
	}
	return nil
}
