# Report

A package for generating reports from a list of packages `[]listen.Packages`.

It can generate reports in the following formats:

- JSON:  `report.NewJSONReport()`
- Markdown: `report.NewFullMarkdwonReport()`


## Example

```go
package main

import (
	"log"
	"os"

	"github.com/listendev/lstn/pkg/cmd/report"
	"github.com/listendev/lstn/pkg/listen"
)

func main() {
	// json report
	jsonReport := report.NewJSONReport()

	// json report

	jsonReportFile, err := os.Create("/tmp/report.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonReportFile.Close()
	jsonReport.WithOutput(jsonReportFile)

	// full markdown report
	fullMarkdownReport := report.NewFullMarkdwonReport()
	fullMDReportFile, err := os.Create("/tmp/report.md")
	if err != nil {
		log.Fatal(err)
	}
	fullMarkdownReport.WithOutput(fullMDReportFile)

	rb := report.NewBuilder()
	rb.RegisterReport(jsonReport)
	rb.RegisterReport(fullMarkdownReport)

	packages := []listen.Package{
		{
			Name:    "react",
			Version: "1.0.0",
			Verdicts: []listen.Verdict{
				{
					Message:  "unexpected outbound connection destination",
					Severity: "high",
					Metadata: map[string]interface{}{
						"commandline":      "/usr/local/bin/node",
						"file_descriptor:": "10.0.2.100:47326->142.251.111.128:0",
						"server_ip":        "142.251.111.128",
						"executable_path":  "/usr/local/bin/node",
					},
				},
			},
			Problems: []listen.Problem{},
		},
	}
	rb.Render(packages)
}

```
