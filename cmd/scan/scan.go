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
package scan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/arguments"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/options"
	"github.com/listendev/lstn/pkg/cmd/packagesprinter"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/npm"
	reporterfactory "github.com/listendev/lstn/pkg/reporter/factory"
	"github.com/listendev/pkg/ecosystem"
	"github.com/spf13/cobra"
)

var _, filename, _, _ = runtime.Caller(0)

func New(ctx context.Context) (*cobra.Command, error) {
	scanCmd := &cobra.Command{
		Use:                   "scan [path]",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Inspect the verdicts for your direct dependencies",
		Long: `Query listen.dev for the verdicts of the dependencies in your project.

Using this command, you can audit the first-level dependencies configured for a project and obtain their verdicts.
This requires a package.json file to fetch the package name and version of the project dependencies.

The verdicts it returns are listed by the name of each package and its specified version.`,
		Example: `  lstn scan
  lstn scan .
  lstn scan sub/dir
  lstn scan /we/snitch
  lstn scan /we/snitch --ignore-deptypes peer
  lstn scan /we/snitch --ignore-deptypes dev,peer
  lstn scan /we/snitch --ignore-deptypes dev --ignore-deptypes peer
  lstn scan /we/snitch --ignore-packages react,glob --ignore-deptypes peer
  lstn scan /we/snitch --ignore-packages react --ignore-packages glob,@vue/devtools`,
		Args:              arguments.SingleDirectory, // Executes before RunE
		ValidArgsFunction: arguments.SingleDirectoryActiveHelp,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, args []string) error {
			ctx = c.Context()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.ScanKey)
			if err != nil {
				return err
			}
			scanOpts, ok := opts.(*options.Scan)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			if scanOpts.DebugOptions {
				c.Println(scanOpts.AsJSON())

				return nil
			}

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			io.StartProgressIndicator()

			// Obtain the target directory that we want to listen in
			targetDir, err := arguments.GetDirectory(args)
			if err != nil {
				return fmt.Errorf("couldn't get to know which directory you want me to scan")
			}

			packageJSON, err := npm.GetPackageJSONFromDir(targetDir)
			if err != nil {
				return err
			}

			// Exclude dependencies
			packageJSON.FilterOutByTypes(scanOpts.Deptypes...)
			packageJSON.FilterOutByNames(scanOpts.Packages...)

			// Retrieve dependencies to process
			deps := packageJSON.Deps(ctx, npm.DefaultVersionResolutionStrategy)
			if len(deps) == 0 {
				return fmt.Errorf("there are no dependencies to process")
			}

			// Process one dependency set at once
			tablePrinter := packagesprinter.NewTablePrinter(io)
			combinedResponse := listen.Response{}
			for _, deps := range deps {
				// Create list of verdicts requests
				reqs, bulkErr := listen.NewBulkVerdictsRequestsFromMap(deps, scanOpts.Expression)
				if bulkErr != nil {
					return bulkErr
				}

				// Query for verdicts about the current dependencies set in parallel...
				res, resJSON, resErr := listen.BulkPackages(
					reqs,
					listen.WithContext(ctx),
					listen.WithEcosystem(ecosystem.Npm), // FIXME: only NPM at the moment
					listen.WithJSONOptions(scanOpts.JSONFlags),
				)

				if resErr != nil {
					return resErr
				}

				if resJSON != nil {
					fmt.Fprintf(os.Stdout, "%s", resJSON)
				}

				// Appending the results of the current dependency set
				if res != nil {
					combinedResponse = append(combinedResponse, *res...)
				}
			}

			if scanOpts.JSON {
				return nil
			}

			err = tablePrinter.RenderPackages(&combinedResponse)
			if err != nil {
				return err
			}
			src := filepath.Join(targetDir, "package.json")

			return reporterfactory.Exec(c, scanOpts.Reporting, combinedResponse, &src)
		},
	}

	// Obtain the local options
	scanOpts, err := options.NewScan()
	if err != nil {
		return nil, err
	}

	// Persistent flags here will work for this command and all subcommands
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags will only run when this command is called directly
	scanOpts.Attach(scanCmd, []string{"jwt-token", "lockfiles", "core-endpoint"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.ScanKey, scanOpts)
	scanCmd.SetContext(ctx)

	return scanCmd, nil
}
