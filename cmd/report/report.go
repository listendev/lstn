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
package report

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/arguments"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/options"
	"github.com/listendev/lstn/pkg/cmd/report"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var reportCmd = &cobra.Command{
		Use:                   "report <output_path>",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "TODO",
		Long:                  `TODO`,
		Example:               `  lstn scan | lstn report /tmp/myreport`,
		Args:                  arguments.SingleDirectory, // Executes before RunE
		ValidArgsFunction:     arguments.SingleDirectoryActiveHelp,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, args []string) error {
			ctx = c.Context()

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			io.StartProgressIndicator()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.ReportKey)
			if err != nil {
				return err
			}
			reportOpts, ok := opts.(*options.Report)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			// Obtain the target directory that we want to listen in
			targetDir, err := arguments.GetDirectory(args)
			if err != nil {
				return fmt.Errorf("couldn't get to know which directory you want me to write the report to")
			}

			fmt.Println("report, bla bla bla", targetDir, reportOpts)

			packages, err := readPackagesFromReader(io.In)
			if err != nil {
				return err
			}

			// json report
			jsonReport := report.NewJSONReport()
			reportFile, err := createReportFile(targetDir, "report.json")
			if err != nil {
				return err
			}
			jsonReport.WithOutput(reportFile)

			rb := report.NewReportBuilder()
			rb.RegisterReport(jsonReport)

			return rb.Render(packages)
		},
	}

	// Obtain the local options
	reportOpts, err := options.NewReport()
	if err != nil {
		return nil, err
	}

	// Persistent flags here will work for this command and all subcommands
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags will only run when this command is called directly
	reportOpts.Attach(reportCmd)

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.ReportKey, reportOpts)

	reportCmd.SetContext(ctx)

	return reportCmd, nil
}

func readPackagesFromReader(reader io.Reader) ([]listen.Package, error) {
	combinedResponse := []listen.Package{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {

		ret := []listen.Package{}
		inputStr := scanner.Bytes()
		err := json.Unmarshal(inputStr, &ret)
		if err != nil {
			return nil, fmt.Errorf("couldn't decode input JSON command: %w", err)
		}

		combinedResponse = append(combinedResponse, ret...)

	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("couldn't read input JSON: %w", err)
	}
	return combinedResponse, nil
}

func createReportFile(targetDir string, filename string) (*os.File, error) {
	reportFile, err := os.Create(path.Join(targetDir, filename))

	if err != nil {
		return nil, fmt.Errorf("couldn't open the report file: %w", err)
	}

	err = reportFile.Truncate(0)
	if err != nil {
		return nil, fmt.Errorf("couldn't truncate the report file: %w", err)
	}
	_, err = reportFile.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("couldn't seek the report file: %w", err)
	}
	return reportFile, nil
}
