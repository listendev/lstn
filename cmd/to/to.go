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
	"github.com/listendev/lstn/pkg/jsonpath"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/listendev/pkg/ecosystem"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	// Obtain the local options
	toOpts, err := options.NewTo()
	if err != nil {
		return nil, err
	}

	var toCmd = &cobra.Command{
		Use:                   "to <ecosystem> <name> [[version] [shasum] | [version constraint]]",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Get the verdicts of a package",
		Long: `Query listen.dev for the verdicts of a package.

Using this command, you can audit a single package version or all the versions of a package and obtain their verdicts.

Specifying the ecosystem and the package name is mandatory.

It lists out the verdicts of all the versions of the input package name.`,
		Example: `  # Get the verdicts for all the chalk versions that listen.dev owns
  lstn to npm chalk
  # Get the verdicts for version 4.3.4 of the debug package on NPM
  lstn to npm debug 4.3.4
  # Get the listen.dev verdicts for react 18.0.0 with digest b468736d1f4a5891f38585ba8e8fb29f91c3cb96
  lstn to npm react 18.0.0 b468736d1f4a5891f38585ba8e8fb29f91c3cb96
  # Get the verdicts for all the existing chalk versions
  lstn to npm chalk "*"
  # Get the verdicts for nock versions >= 13.2.0 and < 13.3.0
  lstn to npm nock "~13.2.x"
  # Get the verdicts for tap versions >= 16.3.0 and < 16.4.0
  lstn to npm tap "^16.3.0"
  # Get the verdicts for prettier versions >= 2.7.0 <= 3.0.0
  lstn to npm prettier ">=2.7.0 <=3.0.0"
  # Get the verdicts for all the flask versions that listen.dev analysed
  lstn to pypi flask`,
		// Executes before RunE
		Args: func(c *cobra.Command, args []string) error {
			// Do not enforce arguments validation when users uses --debug-options
			if toOpts.DebugOptions {
				return nil
			}

			return arguments.PackageTuple(c, args)
		},
		ValidArgsFunction: arguments.PackageTupleActiveHelp,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		PreRunE: func(c *cobra.Command, args []string) error {
			// Theoretically, it's impossible args[0] is not a valid ecosystem at this point (because of the Args function)
			eco, _ := ecosystem.FromString(args[0])

			if len(args) > 2 {
				// Theoretically, it's impossible args[2] is not a valid semver constraint at this point (because of the Args function)
				constraints, _ := semver.NewConstraint(args[2])

				var versions semver.Collection
				var err error
				switch eco {
				case ecosystem.Npm:
					versions, err = npm.GetVersionsFromRegistry(c.Context(), args[1], constraints)
				case ecosystem.Pypi:
					versions, err = pypi.GetVersionsFromRegistry(c.Context(), args[1], constraints) // FIXME: implement for PyPi
				}
				if err != nil {
					return err
				}

				// Store all of its versions matching the constraints
				c.SetContext(context.WithValue(c.Context(), pkgcontext.VersionsCollection, versions))
			}

			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			ctx = c.Context()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.ToKey)
			if err != nil {
				return err
			}
			var ok bool
			toOpts, ok = opts.(*options.To)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			if toOpts.DebugOptions {
				c.Println(toOpts.AsJSON())

				return nil
			}

			var res *listen.Response
			var resJSON []byte
			var resErr error

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			io.StartProgressIndicator()

			// Theoretically, it's impossible args[0] is not a valid ecosystem at this point (because of the Args function)
			eco, _ := ecosystem.FromString(args[0])

			versions, multiple := ctx.Value(pkgcontext.VersionsCollection).(semver.Collection)
			if multiple {
				nv := len(versions)

				names := make([]string, nv)
				for i := 0; i < nv; i++ {
					names[i] = args[1]
				}

				// Create list of verdicts requests
				reqs, multipleErr := listen.NewBulkVerdictsRequests(names, versions, toOpts.ConfigFlags.Filtering.Expression)
				if multipleErr != nil {
					return multipleErr
				}

				// Query for verdicts about specific package versions...
				res, resJSON, resErr = listen.BulkPackages(reqs, listen.WithContext(ctx), listen.WithJSONOptions(toOpts.JSONFlags), listen.WithEcosystem(eco))

				goto EXIT
			}

			// Query for one single package version...
			// Or for all the package versions listen.dev owns of the target package
			{
				// New block so it's safe to skip variable declarations
				req, reqErr := listen.NewVerdictsRequest(args[1:])
				if reqErr != nil {
					return reqErr
				}
				req.Select = jsonpath.Make(toOpts.ConfigFlags.Filtering.Expression)

				res, resJSON, resErr = listen.Packages(
					req,
					listen.WithContext(ctx),
					listen.WithEcosystem(eco),
					listen.WithJSONOptions(toOpts.JSONFlags),
				)
			}

		EXIT:
			io.StopProgressIndicator()
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

	// Local flags will only run when this command is called directly
	toOpts.Attach(toCmd, []string{"--reporter", "--gh-owner", "--gh-repo", "--gh-pull-id", "--gh-token", "--jwt-token", "--ignore-packages", "--ignore-deptypes", "--lockfiles"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.ToKey, toOpts)
	toCmd.SetContext(ctx)

	return toCmd, nil
}
