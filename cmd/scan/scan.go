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
	"runtime"

	"github.com/XANi/goneric"
	"github.com/davecgh/go-spew/spew"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/arguments"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var scanCmd = &cobra.Command{
		Use:                   "scan <path>",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Inspect the verdicts for your direct dependencies",
		Long: `Query listen.dev for the verdicts of the dependencies in your project.

Using this command, you can audit the first-level dependencies configured for a project and obtain their verdicts.
This requires a package.json file to fetch the package name and version of the project dependencies.

The verdicts it returns are listed by the name of each package and its specified version.`,
		Example: `  lstn scan
  lstn scan .
  lstn scan /we/snitch
  lstn scan /we/snitch -e peer
  lstn scan /we/snitch -e dev,peer
  lstn scan /we/snitch -e dev -e peer
  lstn scan sub/dir`,
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

			// Obtain the target directory that we want to listen in
			targetDir, err := arguments.GetDirectory(args)
			if err != nil {
				return fmt.Errorf("couldn't get to know which directory you want me to scan")
			}

			packageJSON, err := npm.GetPackageJSONFromDir(targetDir)
			if err != nil {
				return err
			}

			// Exclude dependencies sets
			dependenciesTypes, _ := goneric.SliceDiff(npm.AllDependencyTypes, scanOpts.Excludes)

			// We don't want to fallback to process all the dependencies sets by default when the user excluded all of them
			if len(dependenciesTypes) == 0 {
				return fmt.Errorf("no dependencies sets to process")
			}

			deps := packageJSON.Deps(ctx, npm.DefaultVersionResolutionStrategy, dependenciesTypes...)

			if len(deps) == 0 {
				return fmt.Errorf("there are no dependencies to process for the currently selected sets of dependencies")
			}

			// Process one dependency set at once
			for _, deps := range deps {
				names := goneric.MapSliceKey(deps)
				versions := goneric.MapSliceValue(deps)
				// Create list of verdicts requests
				reqs, err := listen.NewBulkVerdictsRequests(names, versions)
				if err != nil {
					return err
				}

				// Query for verdicts about specific package versions in bulk...
				res, resJSON, resErr := listen.BulkPackages(reqs, listen.WithContext(ctx), listen.WithJSONOptions(scanOpts.JSONFlags))

				if resErr != nil {
					return err
				}

				if resJSON != nil {
					fmt.Fprintf(os.Stdout, "%s", resJSON)
				}

				if res != nil {
					spew.Dump(res)
					// TODO > create visualization of the results
				}
			}

			return nil
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
	scanOpts.Attach(scanCmd)

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.ScanKey, scanOpts)
	// Pass the registry option as a standalone to do not depend on the command
	ctx = context.WithValue(ctx, pkgcontext.RegistryKey, &scanOpts.RegistryFlags)

	scanCmd.SetContext(ctx)

	return scanCmd, nil
}
