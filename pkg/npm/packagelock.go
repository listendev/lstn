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
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-billy/v5/util"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/validate"
)

// generatePackageLock generates a package-lock.json by executing npm
// against the package.json file in the dir directory.
//
// It returns the package-lock.json file as a byte array.
//
// It assumes that the input directory exists and it already contains
// a package.json file.
func generatePackageLock(ctx context.Context, dir string) ([]byte, error) {
	// Get the npm command
	npmPackageLockOnly, err := getNPMPackageLockOnly(ctx)
	if err != nil {
		// Early exit if it's a timeout (or similar) error from the context
		if ctxErr := pkgcontext.Error(ctx, err); ctxErr != nil {
			return []byte{}, ctxErr
		}

		// Fallback to npm via nvm
		npmPackageLockOnlyFromNVM, nvmErr := getNPMPackageLockOnlyFromNVM(ctx)
		if nvmErr != nil {
			// FIXME > return more errors or a generic one
			return []byte{}, pkgcontext.OutputErrorf(ctx, nvmErr, "couldn't find the npm executable in any way")
		}

		npmPackageLockOnly = npmPackageLockOnlyFromNVM
	}

	// Create temporary directory
	tmp, err := util.TempDir(activeFS, "", "lstn-*")
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't create a temporary directory where to do the dirty work")
	}
	defer util.RemoveAll(activeFS, tmp)

	// Copy the package.json in the temporary directory
	packageJSONPath := filepath.Join(dir, "package.json")
	packageJSON, err := util.ReadFile(activeFS, packageJSONPath)
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't read the package.json file")
	}
	if err := util.WriteFile(activeFS, filepath.Join(tmp, "package.json"), packageJSON, 0644); err != nil {
		return []byte{}, fmt.Errorf("couldn't copy the package.json file")
	}

	// Generate the package-lock.json file
	// TODO(leodido) > Show progress?
	npmPackageLockOnly.Dir = tmp
	if err := npmPackageLockOnly.Run(); err != nil {
		return []byte{}, pkgcontext.OutputErrorf(ctx, err, "couldn't generate the package-lock.json file")
	}
	packageLockJSON, _ := util.ReadFile(activeFS, filepath.Join(tmp, "package-lock.json"))

	return packageLockJSON, nil
}

func checkNPMVersion(c *exec.Cmd, constraint string) error {
	// Obtain the npm version
	npmVersionOut, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("couldn't get the npm version")
	}
	npmVersionString := string(bytes.Trim(npmVersionOut, "\n"))

	// Check the npm version is valid
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

// getNPMPackageLockOnly returns the command to generate the package-lock.json file.
//
// It also checks that:
// - the npm executable is available in the PATH
// - its version is greater or equal than version "6.x".
func getNPMPackageLockOnly(ctx context.Context) (*exec.Cmd, error) {
	// Check the system has the npm executable
	exe, err := exec.LookPath("npm")
	if err != nil {
		return nil, fmt.Errorf("couldn't find the npm executable in the PATH")
	}

	npmVersionCmd := exec.CommandContext(ctx, exe, "--version")
	if err := checkNPMVersion(npmVersionCmd, ">= 6.x"); err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, exe, "install", "--package-lock-only", "--no-audit"), nil
}

// getNPMPackageLockOnlyFromNVM return the command to generate the package-lock.json file
// when the npm executable is behind nvm.
//
// In fact, it is likely that npm is not in the PATH because nvm is lazy-loading it.
//
// It also checks that:
// - the npm version is greater or equal than version "6.x".
func getNPMPackageLockOnlyFromNVM(ctx context.Context) (*exec.Cmd, error) {
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		return nil, fmt.Errorf("couldn't detect the nvm directory")
	}
	bashExe, err := exec.LookPath("bash")
	if err != nil {
		return nil, fmt.Errorf("couldn't find bash in the PATH")
	}

	cmdline := fmt.Sprintf("source %s/nvm.sh", nvmDir)

	nvmNoUse := os.Getenv("NVM_NO_USE")
	if nvmNoUse == "true" {
		cmdline += " --no-use"
	}

	// Obtain the npm version
	npmVersionCmd := exec.CommandContext(ctx, bashExe, "-c", fmt.Sprintf("%s && npm --version", cmdline))
	if err := checkNPMVersion(npmVersionCmd, ">= 6.x"); err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, bashExe, "-c", fmt.Sprintf("%s && npm install --package-lock-only --no-audit", cmdline)), nil
}
