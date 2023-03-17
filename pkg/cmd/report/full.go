package report

import (
	"fmt"
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
	for _, p := range packages {
		err := templates.RenderSingleVerdictsPackage(r.output, p)
		if err != nil {
			return fmt.Errorf("couldn't render package %s: %w", p.Name, err)
		}
	}
	return nil
}
