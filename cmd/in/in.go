/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package in

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/listendev/lstn/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
)

func New(ctx context.Context) (*cobra.Command, error) {
	var inCmd = &cobra.Command{
		Use:   "in",
		Short: "Inspect the verdicts of your dependencies",
		Long: `Query listen.dev for the verdicts of all the dependencies in your project.
		
Using this command, you can audit all of the dependencies configured for a project and obtain their verdicts.
This requires a package.json file to fetch the package name and version for each of the dependencies specified for the project.

The verdicts it returns are listed by the name of each package and its specified version.`,
		Example: `  lstn in
  lstn in .
  lstn in /we/snitch
  lstn in sub/dir`,
		Args:              validateInArgs, // Executes before RunE
		ValidArgsFunction: activeHelpIn,
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.InKey)
			if err != nil {
				return err
			}
			inOpts, ok := opts.(*flags.InOptions)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current subcommand")
			}

			// Obtain the target directory that we want to listen in
			targetDir, err := getTargetDirectory(args)
			if err != nil {
				return fmt.Errorf("couldn't get to know on which directory you want me to listen in")
			}
			// Check that the target directory contains a package.json file
			packageJSONErrors := validate.Singleton.Var(filepath.Join(targetDir, "package.json"), "file")
			// NOTE > In the future, we can try to detect other package managers here rather than erroring out
			if packageJSONErrors != nil {
				return fmt.Errorf("couldn't find a package.json in %s", targetDir)
			}

			// Unmarshal the package-lock.json file contents to a struct
			packageLockJSON, err := npm.NewPackageLockJSONFrom(ctx, targetDir)
			if err != nil {
				return err
			}

			packagesWithShasum, err := packageLockJSON.Shasums(ctx)
			if err != nil {
				return err
			}
			if len(packagesWithShasum) != len(packageLockJSON.Deps()) {
				return fmt.Errorf("couldn't find all the dependencies as per package-lock.json file")
			}

			// Ask listen.dev for an analysis
			req := &listen.AnalysisRequest{
				PackageLockJSON: packageLockJSON,
				Packages:        packagesWithShasum,
			}

			res, resJSON, err := listen.PackageLockAnalysis(ctx, req, inOpts.Json)
			if err != nil {
				return err
			}

			if inOpts.Json && resJSON != nil {
				fmt.Println(string(resJSON))
			}

			if !inOpts.Json && res != nil {
				spew.Dump(res)
				// TODO > create visualization of the results
			}

			return nil
		},
	}

	// Obtain the local options
	inOpts, err := flags.NewInOptions()
	if err != nil {
		return nil, err
	}

	// Persistent flags here will work for this command and all subcommands
	// inCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags will only run when this command is called directly
	inCmd.Flags().BoolVar(&inOpts.Json, "json", inOpts.Json, "output the verdicts (if any) in JSON form")

	// Pass the configuration options through the context
	ctx = context.WithValue(ctx, pkgcontext.InKey, inOpts)
	inCmd.SetContext(ctx)

	return inCmd, nil
}

// getTargetDirectory computes the absolute
// path from the input arguments.
//
// When no argument has been specified,
// it return the current working directory.
func getTargetDirectory(args []string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if len(args) > 0 {
		dir = args[0]
	}

	return filepath.Abs(dir)
}

// validateInArgs validates the input arguments.
//
// It checks that there is maximum one argument.
// It checks that the argument is an existing directory, too.
func validateInArgs(c *cobra.Command, args []string) error {
	if err := cobra.MaximumNArgs(1)(c, args); err != nil {
		return err
	}
	// No further validation left if there are no arguments at all
	if len(args) == 0 {
		return nil
	}
	if errs := validate.Singleton.Var(args[0], "dir"); errs != nil {
		return fmt.Errorf("requires the argument to be an existing directory")
	}

	return nil
}

// TODO(leodido) > Double-check it's working
func activeHelpIn(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var comps []string
	if len(args) == 0 {
		comps = cobra.AppendActiveHelp(comps, "Executing against the current working directory")
	} else if len(args) == 1 {
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Executing against directory '%s'", args[0]))
	} else {
		comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments")
	}
	return comps, cobra.ShellCompDirectiveFilterDirs
}
