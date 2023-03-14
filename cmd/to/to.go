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
package to

import (
	"context"
	"fmt"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/arguments"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/options"
	"github.com/listendev/lstn/pkg/cmd/packagesprinter"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var toCmd = &cobra.Command{
		Use:                   "to <name> [[version] [shasum] | [version constraint]]",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Get the verdicts of a package",
		Long: `Query listen.dev for the verdicts of a package.

Using this command, you can audit a single package version or all the versions of a package and obtain their verdicts.

Specifying the package name is mandatory.

It lists out the verdicts of all the versions of the input package name.`,
		Example: `  # Get the verdicts for all the chalk versions that listen.dev owns
  lstn to chalk
  lstn to debug 4.3.4
  lstn to react 18.0.0 b468736d1f4a5891f38585ba8e8fb29f91c3cb96

  # Get the verdicts for all the existing chalk versions
  lstn to chalk "*"
  # Get the verdicts for nock versions >= 13.2.0 and < 13.3.0
  lstn to nock "~13.2.x"
  # Get the verdicts for tap versions >= 16.3.0 and < 16.4.0
  lstn to tap "^16.3.0"
  # Get the verdicts for prettier versions >= 2.7.0 <= 3.0.0
  lstn to prettier ">=2.7.0 <=3.0.0"`,
		Args:              arguments.PackageTriple, // Executes before RunE
		ValidArgsFunction: arguments.PackageTripleActiveHelp,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, args []string) error {
			ctx = c.Context()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.ToKey)
			if err != nil {
				return err
			}
			toOpts, ok := opts.(*options.To)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			var res *listen.Response
			var resJSON []byte
			var resErr error

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			io.StartProgressIndicator()

			versions, multiple := ctx.Value(pkgcontext.VersionsCollection).(semver.Collection)
			if multiple {
				nv := len(versions)

				names := make([]string, nv)
				for i := 0; i < nv; i++ {
					names[i] = args[0]
				}

				// Create list of verdicts requests
				reqs, multipleErr := listen.NewBulkVerdictsRequests(names, versions)
				if multipleErr != nil {
					return multipleErr
				}

				// Query for verdicts about specific package versions...
				res, resJSON, resErr = listen.BulkPackages(reqs, listen.WithContext(ctx), listen.WithJSONOptions(toOpts.JSONFlags))

				goto EXIT
			}

			// Query for one single package version...
			// Or for all the package versions listen.dev owns of the target package
			{
				// New block so it's safe to skip variable declarations
				req, reqErr := listen.NewVerdictsRequest(args)
				if reqErr != nil {
					return reqErr
				}

				res, resJSON, resErr = listen.Packages(
					req,
					listen.WithContext(ctx),
					listen.WithJSONOptions(toOpts.JSONFlags),
				)
			}

		EXIT:
			if resErr != nil {
				return err
			}

			if resJSON != nil {
				fmt.Fprintf(io.Out, "%s", resJSON)
			}

			if res == nil {
				return nil
			}

			tablePrinter := packagesprinter.NewTablePrinter(io)

			return tablePrinter.RenderPackages(res)
		},
	}

	// Obtain the local options
	toOpts, err := options.NewTo()
	if err != nil {
		return nil, err
	}

	// Local flags will only run when this command is called directly
	toOpts.Attach(toCmd)

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.ToKey, toOpts)
	// Pass the registry option as a standalone to do not depend on the command
	ctx = context.WithValue(ctx, pkgcontext.RegistryKey, &toOpts.RegistryFlags)

	toCmd.SetContext(ctx)

	return toCmd, nil
}
