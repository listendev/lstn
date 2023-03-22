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

func (suite *FlagsConfigSuite) TestGetConfigFlagsNames() {
	m := GetNames(&ConfigFlags{})
	assert.Equal(suite.T(), 9, len(m))

	expected := make(map[string]string)
	expected["loglevel"] = "LogLevel"
	expected["endpoint"] = "Endpoint"
	expected["timeout"] = "Timeout"
	expected["gh-token"] = "Token.GitHub"
	expected["gh-pull-id"] = "Reporter.GitHub.Pull.ID"
	expected["gh-repo"] = "Reporter.GitHub.Repo"
	expected["gh-owner"] = "Reporter.GitHub.Owner"
	expected["reporter"] = "Reporter.Types"
	expected["npm-registry"] = "Registry.NPM"

	for k, v := range m {
		e, ok := expected[k]
		assert.True(suite.T(), ok)
		assert.Equal(suite.T(), e, v)
	}
}

func (suite *FlagsConfigSuite) TestGetConfigFlagsDefaults() {
	m := GetDefaults(&ConfigFlags{})
	assert.Equal(suite.T(), 5, len(m))

	expected := make(map[string]string)
	expected["endpoint"] = "https://npm.listen.dev"
	expected["loglevel"] = "info"
	expected["timeout"] = "60"
	expected["npm-registry"] = "https://registry.npmjs.org"
	expected["gh-pull-id"] = "0"

	for k, v := range m {
		e, ok := expected[k]
		assert.True(suite.T(), ok)
		assert.Equal(suite.T(), e, v)
	}
}
