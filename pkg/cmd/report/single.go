package report

import (
	"fmt"
	"io"

	"github.com/listendev/lstn/pkg/listen"
)

type SingleDependencyMarkdownReport struct {
	outputs map[int]io.Writer
}

func NewSingleDependencyMarkdownReport() *SingleDependencyMarkdownReport {
	return &SingleDependencyMarkdownReport{
		outputs: make(map[int]io.Writer),
	}
}

func (r *SingleDependencyMarkdownReport) WithOutputs(outputs map[int]io.WriteCloser) {
	for k, v := range outputs {
		r.outputs[k] = v
	}
}

func (r *SingleDependencyMarkdownReport) Render(packages []listen.Package) error {
	for k, p := range packages {
		depOutput, ok := r.outputs[k]
		if !ok {
			return fmt.Errorf("couldn't find output for dependency %s", p.Name)
		}
		_, err := depOutput.Write([]byte("## " + p.Name + "\n\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
