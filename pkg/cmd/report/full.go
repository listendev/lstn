package report

import (
	"io"

	"github.com/listendev/lstn/pkg/cmd/report/templates"
	"github.com/listendev/lstn/pkg/listen"
)

type FullMarkdwonReport struct {
	output io.Writer
}

func NewFullMarkdwonReport() *FullMarkdwonReport {
	return &FullMarkdwonReport{}
}

func (r *FullMarkdwonReport) WithOutput(w io.Writer) {
	r.output = w
}

func (r *FullMarkdwonReport) Render(packages []listen.Package) error {
	return templates.RenderContainer(r.output, packages)
}