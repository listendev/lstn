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
{{ $priority := index .Metadata "ai_rank" }}
{{ $staticPriority := .Priority}}
{{ if not $priority }}
	{{ $priority := $staticPriority }}
{{ end }}
{{ $priorityEmoji := ":large_blue_diamond:" }}
{{ if eq $priority "high" }}
	{{ $priorityEmoji = ":stop_sign:" }}
{{ else if eq $priority "medium" }}
	{{ $priorityEmoji = ":warning:" }}
{{ else if eq $priority "low" }}
	{{ $priorityEmoji = ":large_blue_diamond:" }}
{{ end }}
## {{ $priorityEmoji }} {{ .Message }}
<dl>
<dt>Dependency type</dt>
<dd>
{{ if and (eq (index .Metadata "npm_package_name") $.Name) (eq (index .Metadata "npm_package_version") $.Version) }}
Direct dependency
{{ else }}
{{ $transitivePackageName := index .Metadata "npm_package_name" }}
{{ $transitivePackageVersion := index .Metadata "npm_package_version" }}
Transitive dependency (<a href="https://www.npmjs.com/package/{{ $transitivePackageName }}/v/{{ $transitivePackageVersion }}">{{ $transitivePackageName }}@{{ $transitivePackageVersion }}</a>)
{{ end }}
</dd>
{{ if index .Metadata "ai_context" }}
<dt>Context</dt>
<dd>{{ index .Metadata "ai_context" }}</dd>
{{ end }}
{{ if index .Metadata "ai_actions" }}
<dt>Suggested actions</dt>
<dd>
{{ range $action := index .Metadata "ai_actions" }}
- {{ $action }}
{{ end }}
</dd>
{{ end }}
<dt>Metadata</dt>
<dd>
<table>
{{ range $key, $value := .Metadata }}
{{ if or (eq $key "npm_package_name")
        (eq $key "npm_package_version")
        (eq $key "ai_context")
        (eq $key "ai_actions")
        (eq $key "ai_concern")
        (eq $key "ai_rank")
}}
    {{ continue }}
{{ end }}
{{ if not $value }}
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
