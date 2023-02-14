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
package npm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/Masterminds/semver/v3"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/validate"
)

// Version returns the npm version.
//
// Or it errors out if unable to find the npm executable.
func Version(ctx context.Context) (string, error) {
	exe, err := getNPM(ctx)
	if err != nil {
		// Early exit if it's a timeout (or similar) error from the context
		if ctxErr := pkgcontext.Error(ctx, err); ctxErr != nil {
			return "", ctxErr
		}

		// Fallback to npm via nvm
		var nvmErr error
		exe, nvmErr = getNPMFromNVM(ctx)
		if nvmErr != nil {
			return "", pkgcontext.OutputErrorf(ctx, nvmErr, "couldn't find the npm executable via nvm")
		}
		exe.Args[2] += " --version"
	} else {
		exe.Args = append(exe.Args, "--version")
	}

	return getNPMVersion(exe)
}

func getNPMVersion(c *exec.Cmd) (string, error) {
	npmVersionOut, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("couldn't get the npm version")
	}

	return string(bytes.Trim(npmVersionOut, "\n")), nil
}

// checkNPMVersion checks the npm version is valid and fits the constraint.
func checkNPMVersion(npmVersionString, constraint string) error {
	npmVersionErrors := validate.Singleton.Var(npmVersionString, "semver")
	if npmVersionErrors != nil {
		return fmt.Errorf("the npm version is not a valid semantic version")
	}
	npmVersion, err := semver.NewVersion(npmVersionString)
	if err != nil {
		return fmt.Errorf("couldn't get a valid npm version")
	}

	// Check the npm is at least version 6.x
	npmVersionConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		return fmt.Errorf("couldn't compare the npm version")
	}
	npmVersionValid, _ := npmVersionConstraint.Validate(npmVersion)
	if !npmVersionValid {
		return fmt.Errorf("the npm version is not %s", constraint)
	}

	return nil
}
