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
package arguments

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
)

// PackageTriple validates that the arguments are a triple package, version, and shasum.
//
// It checks that there's a package name as first arguments, at least.
// It checks that there are no more that 3 arguments.
// It accepts a single version or versions constraints as the second argument,
// in which case the third one is ignored.
func PackageTriple(c *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(1)(c, args); err != nil {
		return fmt.Errorf("requires at least 1 arg (package name)")
	}
	if err := cobra.RangeArgs(1, 3)(c, args); err != nil {
		return err
	}

	var constraints *semver.Constraints
	all := []error{}
	switch len(args) {
	case 3:
		// Validate third argument is a shasum
		if err := validate.Singleton.Var(args[2], "shasum"); err != nil {
			all = append(all, fmt.Errorf("%s is not a valid shasum", args[2]))
		}

		fallthrough
	case 2:
		// Validate second argument is a valid semver version
		if err := validate.Singleton.Var(args[1], "semver"); err != nil {
			// Then, check whether it is a valid version constraint
			if err := validate.Singleton.Var(args[1], "version_constraint"); err != nil {
				all = append(all, fmt.Errorf("%s is neither a valid version constraint not an exact valid semantic version", args[1]))
			} else {
				// Theoretically this should never error at this point
				constraints, _ = semver.NewConstraint(args[1])
				// Ignore shasum errors
				all = []error{}
			}
		}

		fallthrough
	case 1:
		if err := validate.Singleton.Var(args[0], "npm_package_name"); err != nil {
			// Check the first argument is a valid package name
			all = append(all, fmt.Errorf("%s is not a valid npm package name", args[0]))
		} else {
			if constraints != nil {
				versions, err := npm.GetVersionsFromRegistry(c.Context(), args[0], constraints)
				if err != nil {
					return err
				}
				// Store all of its versions matching the constraints
				c.SetContext(context.WithValue(c.Context(), pkgcontext.VersionsCollection, versions))
			}
		}
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

// PackageTripleActiveHelp generates the active help for a triple package, version (or version constraint), and shasumxs.
func PackageTripleActiveHelp(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO:  Double-check it's working.
	var comps []string

	switch len(args) {
	case 0:
		comps = cobra.AppendActiveHelp(comps, "Provide a package name")
	case 1:
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Provide the version of package %s, or a version constraint, or just execute the command like it is now", args[0]))
	case 2:
		// TODO: no active help if the previous argument was a version constraint
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Provide the shasum of package %s@%s or just execute the command like it is now", args[0], args[1]))
	default:
		comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments")
	}

	return comps, cobra.ShellCompDirectiveFilterDirs
}
