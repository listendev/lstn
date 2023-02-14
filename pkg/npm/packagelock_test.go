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
	"strings"
	"testing"

	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestCheckNPMVersionLt6x(t *testing.T) {
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-lt-6x")
	testExe, err := os.Executable()
	assert.Nil(t, err)

	npmVersionCmd := exec.Command(testExe, "--version")
	v, err := getNPMVersion(npmVersionCmd)
	assert.Nil(t, err)
	err = checkNPMVersion(v, ">= 6.x")
	if assert.Error(t, err) {
		assert.Equal(t, "the npm version is not >= 6.x", err.Error())
	}
}

func TestCheckNPMVersionGt6x(t *testing.T) {
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-gt-6x")
	testExe, err := os.Executable()
	assert.Nil(t, err)

	npmVersionCmd := exec.Command(testExe, "--version")
	v, err := getNPMVersion(npmVersionCmd)
	assert.Nil(t, err)
	assert.Nil(t, checkNPMVersion(v, ">= 6.x"))
}

func TestCheckNPMVersionNotValidSemver(t *testing.T) {
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-non-semver")
	testExe, err := os.Executable()
	assert.Nil(t, err)

	npmVersionCmd := exec.Command(testExe, "--version")
	v, err := getNPMVersion(npmVersionCmd)
	assert.Nil(t, err)
	err = checkNPMVersion(v, ">= 6.x")
	if assert.Error(t, err) {
		assert.Equal(t, "the npm version is not a valid semantic version", err.Error())
	}
}

func TestGetNPMPackageLockOnlyGt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	newPath := binDir + string(filepath.ListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", newPath)
	dstNPM := filepath.Join(binDir, "npm")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstNPM))
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-gt-6x")

	c, e := getNPMPackageLockOnly(context.Background())
	assert.Nil(t, e)
	assert.Equal(t, fmt.Sprintf("%s install --package-lock-only --no-audit", dstNPM), c.String())
}

func TestGetNPMPackageLockOnlyLt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	newPath := binDir + string(filepath.ListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", newPath)
	dstNPM := filepath.Join(binDir, "npm")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstNPM))
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-lt-6x")

	c, e := getNPMPackageLockOnly(context.Background())
	if assert.Error(t, e) {
		assert.Nil(t, c)
		assert.Equal(t, "the npm version is not >= 6.x", e.Error())
	}
}

func TestGetNPMPackageLockOnlyNPMNotInPath(t *testing.T) {
	binDir := t.TempDir()
	newPath := binDir
	t.Setenv("PATH", newPath)

	c, e := getNPMPackageLockOnly(context.Background())
	if assert.Error(t, e) {
		assert.Nil(t, c)
		assert.Equal(t, "couldn't find the npm executable in the PATH", e.Error())
	}
}

func TestGetNPMPackageLockOnlyFromNVMMissingNVMDir(t *testing.T) {
	t.Setenv("NVM_DIR", "")

	c, e := getNPMPackageLockOnlyFromNVM(context.Background())
	if assert.Error(t, e) {
		assert.Nil(t, c)
		assert.Equal(t, "couldn't detect the nvm directory", e.Error())
	}
}

func TestGetNPMPackageLockOnlyFromNVMNoUseGt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	t.Setenv("NVM_DIR", binDir)
	t.Setenv("NVM_NO_USE", "true")
	newPath := binDir + string(filepath.ListSeparator)
	t.Setenv("PATH", newPath)
	dstBash := filepath.Join(binDir, "bash")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstBash))
	t.Setenv("TEST_NPM_BEHAVIOR", "nvm-gt-6x-no-use")

	c, e := getNPMPackageLockOnlyFromNVM(context.Background())
	assert.Nil(t, e)
	assert.Len(t, c.Args, 3)
	assert.True(t, strings.HasSuffix(c.Path, "bash"))
	assert.Contains(t, c.Args[1], "-c")
	assert.Equal(t, fmt.Sprintf("source %s/nvm.sh --no-use && npm install --package-lock-only --no-audit", binDir), strings.Join(c.Args[2:], " "))
}

func TestGetNPMPackageLockOnlyFromNVMGt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	t.Setenv("NVM_DIR", binDir)
	newPath := binDir + string(filepath.ListSeparator)
	t.Setenv("PATH", newPath)
	dstBash := filepath.Join(binDir, "bash")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstBash))
	t.Setenv("TEST_NPM_BEHAVIOR", "nvm-gt-6x")

	c, e := getNPMPackageLockOnlyFromNVM(context.Background())
	assert.Nil(t, e)
	assert.Len(t, c.Args, 3)
	assert.True(t, strings.HasSuffix(c.Path, "bash"))
	assert.Contains(t, c.Args[1], "-c")
	assert.Equal(t, fmt.Sprintf("source %s/nvm.sh && npm install --package-lock-only --no-audit", binDir), strings.Join(c.Args[2:], " "))
}

func TestGetNPMPackageLockOnlyFromNVMLt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	t.Setenv("NVM_DIR", binDir)
	newPath := binDir + string(filepath.ListSeparator)
	t.Setenv("PATH", newPath)
	dstBash := filepath.Join(binDir, "bash")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstBash))
	t.Setenv("TEST_NPM_BEHAVIOR", "nvm-lt-6x")

	c, e := getNPMPackageLockOnlyFromNVM(context.Background())
	if assert.Error(t, e) {
		assert.Nil(t, c)
		assert.Equal(t, "the npm version is not >= 6.x", e.Error())
	}
}
