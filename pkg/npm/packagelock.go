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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-billy/v5/util"
	pkgcontext "github.com/listendev/lstn/pkg/context"
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
	//nolint:errcheck // no need to check the error
	defer util.RemoveAll(activeFS, tmp)

	// Copy the package.json in the temporary directory
	packageJSONPath := filepath.Join(dir, "package.json")
	packageJSON, err := util.ReadFile(activeFS, packageJSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, fmt.Errorf("directory %s does not contain a package.json file", dir)
		}
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

// getNPMPackageLockOnly returns the command to generate the package-lock.json file.
//
// It also checks that:
// - the npm executable is available in the PATH
// - its version is greater or equal than version "6.x".
func getNPMPackageLockOnly(ctx context.Context) (*exec.Cmd, error) {
	npm, err := getNPM(ctx)
	if err != nil {
		return nil, err
	}

	npmVersionCmd := exec.CommandContext(ctx, npm.Path, append(npm.Args[1:], "--version")...)
	npmVersionStr, err := getNPMVersion(npmVersionCmd)
	if err != nil {
		return nil, err
	}
	if err := checkNPMVersion(npmVersionStr, ">= 6.x"); err != nil {
		return nil, err
	}

	npmPackageLockOnlyCmd := exec.CommandContext(ctx, npm.Path, append(npm.Args[1:], "install", "--package-lock-only", "--no-audit", "--ignore-scripts")...)

	return npmPackageLockOnlyCmd, nil
}

// getNPMPackageLockOnlyFromNVM return the command to generate the package-lock.json file
// when the npm executable is behind nvm.
//
// In fact, it is likely that npm is not in the PATH because nvm is lazy-loading it.
//
// It also checks that:
// - the npm version is greater or equal than version "6.x".
func getNPMPackageLockOnlyFromNVM(ctx context.Context) (*exec.Cmd, error) {
	npm, err := getNPMFromNVM(ctx)
	if err != nil {
		return nil, err
	}

	npmVersionCmd := exec.CommandContext(ctx, npm.Path, npm.Args[1:]...)
	npmVersionCmd.Args[2] += " --version"
	npmVersionStr, err := getNPMVersion(npmVersionCmd)
	if err != nil {
		return nil, err
	}
	if err := checkNPMVersion(npmVersionStr, ">= 6.x"); err != nil {
		return nil, err
	}

	npmPackageLockOnlyCmd := exec.CommandContext(ctx, npm.Path, npm.Args[1:]...)
	npmPackageLockOnlyCmd.Args[2] += " install --package-lock-only --no-audit --ignore-scripts"

	return npmPackageLockOnlyCmd, nil
}
