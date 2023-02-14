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
	"os"
	"path/filepath"
	"testing"

	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestVersionLt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	newPath := binDir + string(filepath.ListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", newPath)
	dstNPM := filepath.Join(binDir, "npm")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstNPM))
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-lt-6x")

	v, e := Version(context.TODO())
	assert.Nil(t, e)
	assert.Equal(t, "4.6.1", v)
}

func TestVersionGt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	newPath := binDir + string(filepath.ListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", newPath)
	dstNPM := filepath.Join(binDir, "npm")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstNPM))
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-gt-6x")

	v, e := Version(context.TODO())
	assert.Nil(t, e)
	assert.Equal(t, "8.19.3", v)
}

func TestVersionNVMNoUseGt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	t.Setenv("NVM_DIR", binDir)
	t.Setenv("NVM_NO_USE", "true")
	newPath := binDir + string(filepath.ListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", newPath)
	dstBash := filepath.Join(binDir, "bash")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstBash))
	t.Setenv("TEST_NPM_BEHAVIOR", "nvm-gt-6x-no-use")

	v, e := Version(context.TODO())
	assert.Nil(t, e)
	assert.Equal(t, "8.19.3", v)
}

func TestVersionNVMGt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	t.Setenv("NVM_DIR", binDir)
	newPath := binDir + string(filepath.ListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", newPath)
	dstBash := filepath.Join(binDir, "bash")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstBash))
	t.Setenv("TEST_NPM_BEHAVIOR", "nvm-gt-6x")

	v, e := Version(context.TODO())
	assert.Nil(t, e)
	assert.Equal(t, "8.19.3", v)
}

func TestVersionNVMLt6x(t *testing.T) {
	testExe, err := os.Executable()
	assert.Nil(t, err)

	// Prepend the path of this test binary to PATH
	binDir := t.TempDir()
	t.Setenv("NVM_DIR", binDir)
	newPath := binDir + string(filepath.ListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", newPath)
	dstBash := filepath.Join(binDir, "bash")
	assert.Nil(t, internaltesting.CopyExecutable(testExe, dstBash))
	t.Setenv("TEST_NPM_BEHAVIOR", "nvm-lt-6x")

	v, e := Version(context.TODO())
	assert.Nil(t, e)
	assert.Equal(t, "4.6.1", v)
}
