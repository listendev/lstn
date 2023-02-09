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
	"os"
	"os/exec"
	"testing"

	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	behavior := os.Getenv("TEST_NPM_BEHAVIOR")
	switch behavior {
	case "":
		os.Exit(m.Run())
	case "npm-lt-6x":
		if err := internaltesting.StubNpm(internaltesting.NPM{Version: "4.6.1"}); err != nil {
			os.Exit(1)
		}
	case "npm-gt-6x":
		if err := internaltesting.StubNpm(internaltesting.NPM{Version: "8.19.3"}); err != nil {
			os.Exit(1)
		}
	case "npm-non-semver":
		if err := internaltesting.StubNpm(internaltesting.NPM{Version: "non-semver"}); err != nil {
			os.Exit(1)
		}
	default:
		os.Exit(1)
	}
}

func TestCheckNPMVersionLt6x(t *testing.T) {
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-lt-6x")
	testExe, err := os.Executable()
	assert.Nil(t, err)

	npmVersionCmd := exec.Command(testExe, "--version")
	err = checkNPMVersion(npmVersionCmd, ">= 6.x")
	if assert.Error(t, err) {
		assert.Equal(t, "the npm version is not >= 6.x", err.Error())
	}
}

func TestCheckNPMVersionGt6x(t *testing.T) {
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-gt-6x")
	testExe, err := os.Executable()
	assert.Nil(t, err)

	npmVersionCmd := exec.Command(testExe, "--version")
	assert.Nil(t, checkNPMVersion(npmVersionCmd, ">= 6.x"))
}

func TestCheckNPMVersionNotValidSemver(t *testing.T) {
	t.Setenv("TEST_NPM_BEHAVIOR", "npm-non-semver")
	testExe, err := os.Executable()
	assert.Nil(t, err)

	npmVersionCmd := exec.Command(testExe, "--version")
	err = checkNPMVersion(npmVersionCmd, ">= 6.x")
	if assert.Error(t, err) {
		assert.Equal(t, "the npm version is not a valid semantic version", err.Error())
	}
}

// func TestCheckNPMVersion(t *testing.T) {
// 	fs := memfs.New()

// 	ActiveFS = fs
// 	defer func() { ActiveFS = DefaultFS() }()

// 	err := stubNPM(fs, func() string {
// 		res, err := internaltesting.StubNpm(internaltesting.NPM{Version: "8.19.3"})
// 		assert.Nil(t, err)

// 		return res
// 	})
// 	assert.Nil(t, err)

// 	spew.Dump(fs)

// 	npmVersionCmd := exec.Command("/npm", "--version")
// 	spew.Dump(checkNPMVersion(npmVersionCmd, ">= 6.x"))
// }

// func stubNPM(fs billy.Filesystem, content func() string) error {
// 	if err := internaltesting.WriteFileContent(fs, "/npm", content(), true); err != nil {
// 		return err
// 	}

// 	return nil
// }
