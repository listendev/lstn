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
package in

import (
	"context"
	"fmt"
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
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var inCmd = &cobra.Command{
		Use:                   "in [path]",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Inspect the verdicts for your dependencies tree",
		Long: `Query listen.dev for the verdicts of all the dependencies in your project.

Using this command, you can audit all the dependencies configured for a project and obtain their verdicts.
This requires a package.json file to fetch the package name and version of the project dependencies.

The verdicts it returns are listed by the name of each package and its specified version.`,
		Example: `  lstn in
  lstn in .
  lstn in /we/snitch
  lstn in sub/dir`,
		Args:              arguments.SingleDirectory, // Executes before RunE
		ValidArgsFunction: arguments.SingleDirectoryActiveHelp,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, args []string) error {
			ctx = c.Context()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.InKey)
			if err != nil {
				return err
			}
			inOpts, ok := opts.(*options.In)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			if inOpts.DebugOptions {
				c.Println(inOpts.AsJSON())

				return nil
			}

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			io.StartProgressIndicator()
			defer io.StopProgressIndicator()

			// Obtain the target directory that we want to listen in
			targetDir, err := arguments.GetDirectory(args)
			if err != nil {
				return fmt.Errorf("couldn't get to know on which directory you want me to listen in")
			}

			var packageLockJSON npm.PackageLockJSON
			var packageLockJSONErr error

			if inOpts.GenerateLock {
				// Try to generate a package-lock.json file on the fly
				packageLockJSON, packageLockJSONErr = npm.NewPackageLockJSONFromDir(ctx, targetDir)
			}
			// Fallback to reading the package-lock.json in the target directory
			if packageLockJSONErr != nil || !inOpts.GenerateLock {
				packageLockJSON, packageLockJSONErr = npm.GetPackageLockJSONFromDir(targetDir)
			}
			if packageLockJSONErr != nil {
				return packageLockJSONErr
			}

			// Ask listen.dev for an analysis
			req, err := listen.NewAnalysisRequest(packageLockJSON, listen.WithRequestContext())
			if err != nil {
				return err
			}

			res, resJSON, err := listen.Packages(
				req,
				listen.WithContext(ctx),
				listen.WithJSONOptions(inOpts.JSONFlags),
			)
			if err != nil {
				return err
			}

			if resJSON != nil {
				fmt.Fprintf(io.Out, "%s", resJSON)
			}

			if res == nil {
				return nil
			}

			tablePrinter := packagesprinter.NewTablePrinter(io)

			err = tablePrinter.RenderPackages(res)
			if err != nil {
				return err
			}

			return reporterfactory.Exec(c, inOpts.Reporting, *res)
		},
	}

	// Obtain the local options
	inOpts, err := options.NewIn()
	if err != nil {
		return nil, err
	}

	// Persistent flags here will work for this command and all subcommands
	// inCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags will only run when this command is called directly
	inOpts.Attach(inCmd, []string{"--ignore-packages", "--ignore-deptypes", "--select"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.InKey, inOpts)
	inCmd.SetContext(ctx)

	return inCmd, nil
}
