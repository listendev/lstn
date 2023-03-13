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
	"os"
	"path/filepath"

	"github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
)

// SingleDirectory validates the input arguments is a single directory.
//
// It checks that there is maximum one argument.
// It checks that the argument is an existing directory, too.
func SingleDirectory(c *cobra.Command, args []string) error {
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
	// Check that the target directory contains a package.json file
	packageJSONErrors := validate.Singleton.Var(filepath.Join(args[0], "package.json"), "file")
	// NOTE > In the future, we can try to detect other package managers here rather than erroring out
	if packageJSONErrors != nil {
		return fmt.Errorf("couldn't find a package.json in %s", args[0])
	}

	return nil
}

// GetDirectory computes the absolute path from the input arguments.
//
// When no argument has been specified, it return the current working directory.
func GetDirectory(args []string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if len(args) > 0 {
		dir = args[0]
	}

	return filepath.Abs(dir)
}

// SingleDirectoryActiveHelp generates the active help for a single directory.
func SingleDirectoryActiveHelp(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO:  Double-check it's working.
	var comps []string
	switch len(args) {
	case 0:
		comps = cobra.AppendActiveHelp(comps, "Executing against the current working directory")
	case 1:
		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("Executing against directory '%s'", args[0]))
	default:
		comps = cobra.AppendActiveHelp(comps, "This command does not take any more arguments")
	}

	return comps, cobra.ShellCompDirectiveFilterDirs
}
