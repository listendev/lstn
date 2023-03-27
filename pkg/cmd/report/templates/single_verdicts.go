// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2023 The listen.dev team <engineering@garnet.ai>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package templates

import (
	"io"
	"text/template"

	"github.com/listendev/lstn/pkg/listen"
)

const singleVerdictsTpl = `
{{ if gt (len .Verdicts) 0 }}
## <b><a href="https://www.npmjs.com/package/{{ .Name }}/v/{{ .Version }}">{{ .Name }}@{{ .Version }}</a></b><br>

{{ range .Verdicts }}
{{ $priority := .Priority}}
{{ $priorityEmoji := ":large_blue_diamond:" }}
{{ if eq $priority "high" }}
	{{ $priorityEmoji = ":stop_sign:" }}
{{ else if eq $priority "medium" }}
	{{ $priorityEmoji = ":warning:" }}
{{ else if eq $priority "low" }}
	{{ $priorityEmoji = ":large_blue_diamond:" }}
{{ end }}
### {{ $priorityEmoji }} {{ .Message }}
<dl>
<dt>Dependency type</dt>
<dd>
{{ if and (eq (index .Metadata "npm_package_name") $.Name) (eq (index .Metadata "npm_package_version") $.Version) }}
Direct dependency
{{ else }}
{{ $transitivePackageName := index .Metadata "npm_package_name" }}
{{ $transitivePackageVersion := index .Metadata "npm_package_version" }}
Transitive dependency {{ if and $transitivePackageName $transitivePackageVersion }} (<a href="https://www.npmjs.com/package/{{ $transitivePackageName }}/v/{{ $transitivePackageVersion }}">{{ $transitivePackageName }}@{{ $transitivePackageVersion }}</a>){{ end }}
{{ end }}
</dd>
{{ $aiCtxMetadata := index .Metadata "ai_ctx" }}
{{ if $aiCtxMetadata }}
{{ $aiContext := index $aiCtxMetadata "context" }}
{{ $aiActions := index $aiCtxMetadata "actions" }}
{{ if $aiContext }}
<dt>Context</dt>
<dd>{{ $aiContext }}</dd>
{{ end }}
{{ if $aiActions }}
<dt>Suggested actions</dt>
<dd>
{{ range $action := $aiActions }}
- {{ $action }}
{{ end }}
{{ end}}
</dd>
{{ end }}
<dt>Metadata</dt>
<dd>
<table>
{{ range $key, $value := .Metadata }}
{{ if or (eq $key "npm_package_name")
        (eq $key "npm_package_version")
        (eq $key "ai_ctx")
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
</table>
</dd>
</dl>
{{ end }}
{{ else }}
Nothing to see here, lucky us! :tada:
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
