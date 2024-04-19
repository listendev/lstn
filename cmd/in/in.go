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
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/XANi/goneric"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/arguments"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/options"
	"github.com/listendev/lstn/pkg/cmd/packagesprinter"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	listentype "github.com/listendev/lstn/pkg/listen/type"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/listendev/lstn/pkg/pypi"
	reporterfactory "github.com/listendev/lstn/pkg/reporter/factory"
	"github.com/listendev/pkg/ecosystem"
	"github.com/listendev/pkg/lockfile"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

//nolint:gocyclo // ignore
func New(ctx context.Context) (*cobra.Command, error) {
	var inCmd = &cobra.Command{
		Use:                   "in [path]",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Inspect the verdicts for your dependencies tree",
		Long: `Query listen.dev for the verdicts of all the dependencies in your project.

Using this command, you can audit all the dependencies of a project and obtain their verdicts.
Given a project directory containing manifest files (package-lock.json, poetry.lock, etc),
it fetches the package names and versions of the project dependencies.

The verdicts it returns are listed by the name of each package and its specified version.`,
		Example: `  lstn in
  lstn in .
  lstn in /we/snitch
  lstn in sub/dir
  lstn in --lockfiles poetry.lock,package-lock.json
  lstn in /pyproj --lockfiles poetry.lock`,
		Args:              arguments.SingleDirectory, // Executes before RunE
		ValidArgsFunction: arguments.SingleDirectoryActiveHelp,
		Annotations: map[string]string{
			"source":   project.GetSourceURL(filename),
			"subgroup": groups.WithDirectory.String(),
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

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			cs := io.ColorScheme()

			// For debugging/testing reasons...
			if inOpts.DebugOptions {
				c.Println(inOpts.AsJSON())

				return nil
			}

			// Obtain the target directory that we want to listen in
			targetDir, err := arguments.GetDirectory(args)
			if err != nil {
				return fmt.Errorf("couldn't get to know on which directory you want me to listen in")
			}

			// Lookup the lock files (relative to the working directory)
			foundLockfiles, notFoundLockfiles := arguments.GetLockfiles(targetDir, inOpts.Lockfiles)
			if len(notFoundLockfiles) > 0 {
				notFoundErrors := []string{}
				for _, errs := range notFoundLockfiles {
					notFoundErrors = append(notFoundErrors, goneric.MapSlice(func(e error) string {
						return e.Error()
					}, errs)...)
				}
				sort.SliceStable(notFoundErrors, func(i, j int) bool {
					return notFoundErrors[i] < notFoundErrors[j]
				})
				c.PrintErrln(cs.WarningIcon(), strings.Join(notFoundErrors, fmt.Sprintf("\n%s ", cs.WarningIcon())))
			}
			if len(foundLockfiles) == 0 {
				return fmt.Errorf("directory %s does not contain any lock file", targetDir)
			}

			numIterations := len(foundLockfiles)
			for lp, lf := range foundLockfiles {
				// TODO: check that targetDir == filepath.Dir(lp) for extra safety?
				dir := filepath.Dir(lp)
				eco := lockfile.Ecosystem(lf)

				var toAnalyse listentype.AnalysisRequester
				var lockfileErr error
				switch eco {
				case ecosystem.Npm:
					switch lf {
					case lockfile.PackageLockJSON:
						toAnalyse, lockfileErr = npm.GetPackageLockJSONFromDir(dir)

					default:
						err := fmt.Errorf("could not process %s yet", lp)
						if numIterations == 1 {
							return err
						}
						c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), err.Error())

						continue
					}

				case ecosystem.Pypi:
					switch lf {
					case lockfile.PoetryLock:
						toAnalyse, lockfileErr = pypi.GetPoetryLockFromDir(dir)

					default:
						err := fmt.Errorf("could not process %s yet", lp)
						if numIterations == 1 {
							return err
						}
						c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), err.Error())

						continue
					}

				case ecosystem.None:
					err := fmt.Errorf("couldn't retrieve the ecosystem relative to the %s lock file", lp)
					if numIterations == 1 {
						return err
					}
					c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), err.Error())

					continue
				}

				if lockfileErr != nil {
					err := fmt.Errorf("could not process %s yet: %s", lp, cs.Red(lockfileErr.Error()))
					if numIterations == 1 {
						return err
					}
					c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), err.Error())

					continue
				}

				io.StartProgressIndicator()

				// Prepare analysis request for <eco>.listen.dev/api/analysis
				req, err := listen.NewAnalysisRequest(toAnalyse, listen.WithRequestContext())
				if err != nil {
					io.StopProgressIndicator()
					if numIterations == 1 {
						return err
					}
					c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), fmt.Sprintf("got an error requesting the analysis: %s", cs.Red(err.Error())))

					continue
				}

				// Ask listen.dev to analyze the lockfile
				res, resJSON, err := listen.Packages(
					req,
					listen.WithContext(ctx),
					listen.WithEcosystem(eco),
					listen.WithJSONOptions(inOpts.JSONFlags),
				)
				if err != nil {
					io.StopProgressIndicator()
					if numIterations == 1 {
						return err
					}
					c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), fmt.Sprintf("got an error from the analysis endpoint: %s", cs.Red(err.Error())))

					continue
				}
				if resJSON != nil {
					fmt.Fprintf(io.Out, "%s", resJSON)
				}
				if res == nil {
					io.StopProgressIndicator()
					c.PrintErrln(cs.WarningIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), "couldn't obtain the verdicts but got no error")

					if numIterations == 1 {
						return nil
					}

					continue
				}
				io.StopProgressIndicator()

				c.Println(cs.SuccessIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), fmt.Sprintf("showing verdicts for %s...\n", lp))

				tablePrinter := packagesprinter.NewTablePrinter(io)
				err = tablePrinter.RenderPackages(res)
				if err != nil {
					if numIterations == 1 {
						return err
					}
					c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), fmt.Sprintf("got an error printing the verdicts: %s", cs.Red(err.Error())))

					continue
				}

				errExec := reporterfactory.Exec(c, inOpts.Reporting, *res, &lp)
				if errExec != nil {
					if numIterations == 1 {
						return errExec
					}
					c.PrintErrln(cs.FailureIcon(), cs.Blue(fmt.Sprintf("[%s ecosystem]", eco.Case())), fmt.Sprintf("got an error executing the reporter: %s", cs.Red(errExec.Error())))

					continue
				}
			}

			return nil
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
