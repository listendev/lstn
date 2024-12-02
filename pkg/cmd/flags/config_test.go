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

func (suite *FlagsConfigSuite) TestIsLocalCore() {
	tests := []struct {
		name     string
		endpoint Endpoint
		expected bool
	}{
		{
			name: "Local core with HTTP and fixed IP",
			endpoint: Endpoint{
				Core: "http://192.168.1.125",
			},
			expected: true,
		},
		{
			name: "Local core with HTTP and non-fixed IP",
			endpoint: Endpoint{
				Core: "http://example.com",
			},
			expected: false,
		},
		{
			name: "Local core with HTTPS and fixed IP",
			endpoint: Endpoint{
				Core: "https://192.168.1.1",
			},
			expected: false,
		},
		{
			name: "Local core with HTTPS and non-fixed IP",
			endpoint: Endpoint{
				Core: "https://example.com",
			},
			expected: false,
		},
		{
			name: "Local core with HTTP and invalid IP",
			endpoint: Endpoint{
				Core: "http://999.999.999.999",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := tt.endpoint.IsLocalCore()
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
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
	assert.Equal(suite.T(), 16, len(m))

	expected := make(map[string]string)
	expected["loglevel"] = "LogLevel"
	expected["npm-endpoint"] = "Endpoint.Npm"
	expected["pypi-endpoint"] = "Endpoint.PyPi"
	expected["core-endpoint"] = "Endpoint.Core"
	expected["timeout"] = "Timeout"
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
	assert.Equal(suite.T(), 8, len(m))

	expected := make(map[string]string)
	expected["npm-endpoint"] = "https://npm.listen.dev"
	expected["pypi-endpoint"] = "https://pypi.listen.dev"
	expected["core-endpoint"] = "https://core.listen.dev"
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
