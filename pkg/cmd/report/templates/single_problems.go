package templates

import (
	"io"
	"text/template"

	"github.com/listendev/lstn/pkg/listen"
)

const singleProblemsTpl = `
{{ if gt (len .Problems) 0 }}
### <b><a href="https://www.npmjs.com/package/{{ .Name }}/v/{{ .Version }}">{{ .Name }}@{{ .Version }}</a></b><br>

{{ range .Problems }}
{{ $title := .Title}}
{{ $url := .Type }}
- {{ $title }} (<a href="{{ $url }}">learn more :link:</a>)
{{ end }}
{{ end }}
`

func RenderSingleProblemsPackage(w io.Writer, p listen.Package) error {
	ct := template.Must(template.New("single_problem").Parse(singleProblemsTpl))
	err := ct.Execute(w, p)
	if err != nil {
		return err
	}
	return nil
}
