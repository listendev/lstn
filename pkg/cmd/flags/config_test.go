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
package flags

import (
	"testing"

	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlagsConfigSuite struct {
	suite.Suite
}

func TestFlagsConfigSuite(t *testing.T) {
	suite.Run(t, new(FlagsConfigSuite))
}

func (suite *FlagsConfigSuite) TestNewConfigFlags() {
	i, err := NewConfigFlags()
	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), &ConfigFlags{}, i)
}

func (suite *FlagsConfigSuite) TestNewConfigFlagsDefaults() {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_PATH": "../../ci/testdata/github_event_pull_request.json",
	})
	suite.T().Cleanup(closer)

	i, err := NewConfigFlags()
	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), &ConfigFlags{}, i)

	assert.Equal(suite.T(), 285, i.Reporting.GitHub.Pull.ID)
	assert.Equal(suite.T(), "reviewdog", i.Reporting.GitHub.Owner)
	assert.Equal(suite.T(), "reviewdog", i.Reporting.GitHub.Repo)
	assert.Equal(suite.T(), "https://registry.npmjs.org", i.Registry.NPM)
	assert.Equal(suite.T(), "info", i.LogLevel)
	assert.Equal(suite.T(), "https://npm.listen.dev", i.Endpoint.Npm)
	assert.Equal(suite.T(), "https://pypi.listen.dev", i.Endpoint.PyPi)
	assert.Equal(suite.T(), 60, i.Timeout)
}

func (suite *FlagsConfigSuite) TestGetConfigFlagsNames() {
	m := GetNames(&ConfigFlags{})
	assert.Equal(suite.T(), 15, len(m))

	expected := make(map[string]string)
	expected["loglevel"] = "Verbosity.LogLevel"
	expected["npm-endpoint"] = "Endpoint.Npm"
	expected["pypi-endpoint"] = "Endpoint.PyPi"
	expected["timeout"] = "TimeFlags.Timeout"
	expected["gh-token"] = "Token.GitHub"
	expected["jwt-token"] = "Token.JWT"
	expected["gh-pull-id"] = "Reporting.GitHub.Pull.ID"
	expected["gh-repo"] = "Reporting.GitHub.Repo"
	expected["gh-owner"] = "Reporting.GitHub.Owner"
	expected["reporter"] = "Reporting.Types"
	expected["npm-registry"] = "Registry.NPM"
	expected["ignore-packages"] = "Filtering.Ignore.Packages"
	expected["ignore-deptypes"] = "Filtering.Ignore.Deptypes"
	expected["select"] = "Filtering.Expression"
	expected["lockfiles"] = "Lockfiles"

	for k, v := range m {
		e, ok := expected[k]
		assert.True(suite.T(), ok)
		assert.Equal(suite.T(), e, v)
	}
}

func (suite *FlagsConfigSuite) TestGetConfigFlagsDefaults() {
	m := GetDefaults(&ConfigFlags{})
	assert.Equal(suite.T(), 7, len(m))

	expected := make(map[string]string)
	expected["npm-endpoint"] = "https://npm.listen.dev"
	expected["pypi-endpoint"] = "https://pypi.listen.dev"
	expected["loglevel"] = "info"
	expected["timeout"] = "60"
	expected["npm-registry"] = "https://registry.npmjs.org"
	expected["ignore-packages"] = "[]"
	expected["lockfiles"] = "[\"package-lock.json\",\"poetry.lock\"]"

	for k, v := range m {
		e, ok := expected[k]
		assert.True(suite.T(), ok)
		assert.Equal(suite.T(), e, v)
	}
}
