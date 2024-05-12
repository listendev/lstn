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
	"fmt"

	"github.com/listendev/lstn/pkg/validate"
	"github.com/listendev/pkg/ecosystem"
	"github.com/spf13/cobra"
)

// PackageTuple validates that the arguments are a tuple ecosystem, package, version, and digest.
//
// It checks that there's an ecosystem as first argument.
// It checks that there's a package name as the second argument.
// It checks that there are no more that 4 arguments.
// It accepts a single version or versions constraints as the third argument,
// in which case the fourth (the digest) one is ignored.
// It requires at least 2 arguments.
func PackageTuple(c *cobra.Command, args []string) error {
	if err := cobra.MinimumNArgs(2)(c, args); err != nil {
		return fmt.Errorf("requires at least 2 args (ecosystem and package name)")
	}
	if err := cobra.RangeArgs(2, 4)(c, args); err != nil {
		return err
	}

	all := []error{}
	for idx, arg := range args {
		switch idx {
		case 0:
			// Check the first argument is a valid ecosystem
			if _, err := ecosystem.FromString(arg); err != nil {
				all = append(all, fmt.Errorf("%s is not a valid ecosystem", arg))
				goto checkall
			}

		case 1:
			// Assuming the previous validator works as meant, we can't have an error here
			eco, _ := ecosystem.FromString(args[0])
			switch eco {
			case ecosystem.Npm:
				// Check the first argument is a valid package name
				if err := validate.Singleton.Var(arg, "npm_package_name"); err != nil {
					all = append(all, fmt.Errorf("%s is not a valid npm package name", arg))
					goto checkall
				}
			case ecosystem.Pypi:
				// TODO: pypi package name validation
				break
			default:
				all = append(all, fmt.Errorf("%s ecosystem is not supported", eco.Case()))
				goto checkall
			}

		case 2:
			// Validate third argument is a valid semver version
			if err := validate.Singleton.Var(arg, "semver"); err != nil {
				// Then, check whether it is a valid version constraint
				if err := validate.Singleton.Var(arg, "version_constraint"); err != nil {
					all = append(all, fmt.Errorf("%s is neither a valid version constraint not an exact valid semantic version", arg))
				}
			}

		case 3:
			// Assuming the previous validator works as meant, we can't have an error here
			eco, _ := ecosystem.FromString(args[0])
			switch eco {
			case ecosystem.Npm:
				// Validate fourth argument is a shasum
				if err := validate.Singleton.Var(arg, "shasum"); err != nil {
					all = append(all, fmt.Errorf("%s is not a valid shasum digest", arg))
				}
			case ecosystem.Pypi:
				// Validate fourth argument is a blake2b_256
				if err := validate.Singleton.Var(arg, "digest"); err != nil {
					all = append(all, fmt.Errorf("%s is not a valid blake2b_256 digest", arg))
				}
			}
		}
	}

	// all := []error{}
	// switch len(args) {
	// case 4:
	// 	// Validate fourth argument is a digest
	// 	if err := validate.Singleton.Var(args[3], "digest"); err != nil {
	// 		all = append(all, fmt.Errorf("%s is not a valid digest", args[3]))
	// 	}

	// 	fallthrough
	// case 3:
	// 	// Validate third argument is a valid semver version
	// 	if err := validate.Singleton.Var(args[2], "semver"); err != nil {
	// 		// Then, check whether it is a valid version constraint
	// 		if err := validate.Singleton.Var(args[2], "version_constraint"); err != nil {
	// 			all = append(all, fmt.Errorf("%s is neither a valid version constraint not an exact valid semantic version", args[2]))
	// 		} else {
	// 			all = []error{}
	// 		}
	// 	}

	// 	fallthrough
	// case 2:
	// 	// FIXME: per python? condizionale?
	// 	if err := validate.Singleton.Var(args[1], "npm_package_name"); err != nil {
	// 		// Check the first argument is a valid package name
	// 		all = append(all, fmt.Errorf("%s is not a valid npm package name", args[1]))
	// 	}
	// case 1:
	// 	// TODO: ...
	// }

	// Format errors (if any)
checkall:
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

// PackageTupleActiveHelp generates the active help for a triple package, version (or version constraint), and shasumxs.
func PackageTupleActiveHelp(_ *cobra.Command, args []string, _ /*toComplete*/ string) ([]string, cobra.ShellCompDirective) {
	// TODO:  Double-check it's working.
	var comps []string

	switch len(args) {
	case 0:
		comps = cobra.AppendActiveHelp(comps, "Provide an ecosystem")
	case 1:
		comps = cobra.AppendActiveHelp(comps, "Provide a package name")
	case 2:
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Provide the version of package %s, or a version constraint, or just execute the command like it is now", args[0]))
	case 3:
		// TODO: no active help if the previous argument was a version constraint
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Provide the shasum of package %s@%s or just execute the command like it is now", args[0], args[1]))
	default:
		comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments")
	}

	return comps, cobra.ShellCompDirectiveFilterDirs
}
