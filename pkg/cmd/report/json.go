package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/listendev/lstn/pkg/listen"
)

type JSONReport struct {
	output io.Writer
}

func NewJSONReport() *JSONReport {
	return &JSONReport{}
}

func (r *JSONReport) WithOutput(w io.Writer) {
	r.output = w
}

func (r *JSONReport) Render(packages []listen.Package) error {
	enc := json.NewEncoder(r.output)
	enc.SetIndent("", "  ")
	err := enc.Encode(packages)
	if err != nil {
		return fmt.Errorf("couldn't encode the JSON report: %w", err)
	}
	return nil
}
