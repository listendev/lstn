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
package to

import (
	"context"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/groups"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
)

func New(ctx context.Context) (*cobra.Command, error) {
	var toCmd = &cobra.Command{
		Use:                   "to <name> [version] [shasum]",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Get the verdicts of a package",
		Long: `Query listen.dev for the verdicts of a package.

Using this command, you can audit a single package version or all the versions of a package and obtain their verdicts.

Specifying the package name is mandatory.

The verdicts it returns are listed by the shasum of each version belonging to that package name.
If you're a hairsplitting person, you can also query for the verdicts specific to a package version's shasum.`,
		Example: `  lstn to chalk
  lstn to debug 4.3.4`,
		Args:              validateInArgs, // Executes before RunE
		ValidArgsFunction: activeHelpIn,
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()

			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.ToKey)
			if err != nil {
				return err
			}
			toOpts, ok := opts.(*flags.ToOptions)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current subcommand")
			}

			// Query for the package verdicts
			res, resJSON, err := listen.PackageVerdicts(ctx, listen.NewVerdictsRequest(args), &toOpts.JsonFlags)
			if err != nil {
				return err
			}

			if resJSON != nil {
				fmt.Fprintf(os.Stdout, "%s", resJSON)
			}

			if res != nil {
				spew.Dump(res)
				// TODO > create visualization of the results
			}

			return nil
		},
	}

	// Obtain the local options
	toOpts, err := flags.NewToOptions()
	if err != nil {
		return nil, err
	}

	// Persistent flags here will work for this command and all subcommands
	// toCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags will only run when this command is called directly
	toOpts.Attach(toCmd)

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.ToKey, toOpts)
	toCmd.SetContext(ctx)

	return toCmd, nil
}

func validateInArgs(c *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(1)(c, args); err != nil {
		return fmt.Errorf("requires at least 1 arg (package name)")
	}
	if err := cobra.RangeArgs(1, 3)(c, args); err != nil {
		return err
	}

	// Validate first argument is a valid package name
	all := []error{}
	switch len(args) {
	case 3:
		if err := validate.Singleton.Var(args[2], "len=40"); err != nil {
			all = append(all, fmt.Errorf("%s is not a valid shasum", args[2]))
		}
		fallthrough
	case 2:
		if err := validate.Singleton.Var(args[1], "semver"); err != nil {
			all = append(all, fmt.Errorf("%s is not a valid semantic version", args[1]))
		}
		fallthrough
	default:
		// TODO > validate the package name
	}

	// Format errors (if any)
	if len(all) > 0 {
		ret := "invalid arguments"
		for _, e := range all {
			ret += "\n       "
			ret += e.Error()
		}

		return fmt.Errorf(ret)
	}

	return nil
}

// TODO(leodido) > Double-check it's working
func activeHelpIn(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var comps []string

	if len(args) == 0 {
		comps = cobra.AppendActiveHelp(comps, "Provide a package name")
	} else if len(args) == 1 {
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Provide the version of package %s or just execute the command like it is now", args[0]))
	} else if len(args) == 2 {
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Provide the shasum of package %s@%s or just execute the command like it is now", args[0], args[1]))
	} else {
		comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments")
	}

	return comps, cobra.ShellCompDirectiveFilterDirs
}
