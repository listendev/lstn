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
	"context"
	"testing"

	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlagsBaseSuite struct {
	suite.Suite
}

func TestFlagsBaseSuite(t *testing.T) {
	suite.Run(t, new(FlagsBaseSuite))
}

func (suite *FlagsBaseSuite) TestGetNames() {
	type ScanOpts struct {
		ConfigFlags
		JSONFlags
	}
	res := GetNames(&ScanOpts{})

	// Expecting all the (sub)fields
	assert.Len(suite.T(), res, 7)
}

func (suite *FlagsBaseSuite) TestGetDefaults() {
	type ScanOpts struct {
		ConfigFlags
		JSONFlags
	}
	res := GetDefaults(&ScanOpts{})

	// Only 4 (sub)fields have default values
	assert.Len(suite.T(), res, 4)
}

func (suite *FlagsBaseSuite) TestGetField() {
	cfg := ConfigFlags{
		Token: Token{
			GitHub: "xxxx",
		},
		LogLevel: "info",
	}

	tokenGithubVal := GetField(cfg, "Token.GitHub")
	assert.True(suite.T(), tokenGithubVal.IsValid())
	assert.Equal(suite.T(), "xxxx", tokenGithubVal.Interface())

	logLevelVal := GetField(cfg, "LogLevel")
	assert.True(suite.T(), logLevelVal.IsValid())
	assert.Equal(suite.T(), "info", logLevelVal.Interface())
}

func (suite *FlagsBaseSuite) TestValidate() {
	cases := []struct {
		desc        string
		o           *ConfigFlags
		expectedStr []string
	}{
		{
			"empty config flags",
			&ConfigFlags{},
			[]string{"timeout must be 30 or greater", "endpoint must be a valid URL"},
		},
		{
			"invalid timeout",
			&ConfigFlags{Timeout: 29, Endpoint: "http://127.0.0.1:3000"},
			[]string{"timeout must be 30 or greater"},
		},
		{
			"invalid endpoint",
			&ConfigFlags{Timeout: 31, Endpoint: "http://invalid.endpoint"},
			[]string{"endpoint must be a valid listen.dev endpoint"},
		},
		{
			"valid config flags",
			&ConfigFlags{Timeout: 31, Endpoint: "http://127.0.0.1:3000"},
			[]string{},
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.desc, func(t *testing.T) {
			actual := Validate(tc.o)
			assert.Equal(suite.T(), len(tc.expectedStr), len(actual))
			for _, a := range actual {
				assert.Contains(suite.T(), tc.expectedStr, a.Error())
			}
		})
	}
}

func (suite *FlagsBaseSuite) TestTransform() {
	cases := []struct {
		desc     string
		o        cmd.Options
		expected cmd.Options
		wantErr  error
	}{
		{
			"empty config flags",
			&ConfigFlags{},
			&ConfigFlags{},
			nil,
		},
		{
			"endpoint config with leading slash",
			&ConfigFlags{
				Endpoint: "https://npm.listen.dev/",
			},
			&ConfigFlags{
				Endpoint: "https://npm.listen.dev",
			},
			nil,
		},
		{
			"custom registry with leading slash",
			&ConfigFlags{
				Registry: Registry{
					NPM: "https://registry.npm.org/",
				},
			},
			&ConfigFlags{
				Registry: Registry{
					NPM: "https://registry.npm.org",
				},
			},
			nil,
		},
	}

	ctx := context.Background()
	for _, tc := range cases {
		suite.T().Run(tc.desc, func(t *testing.T) {
			err := Transform(ctx, tc.o)
			if tc.wantErr == nil {
				assert.Equal(t, tc.expected, tc.o)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func (suite *FlagsBaseSuite) TestDefine() {
	type testFlags struct {
		ConfigFlags `flagset:"Config"`
		JSONFlags
	}

	cases := []struct {
		desc  string
		input interface{}
	}{
		{
			"flags definition from struct reference",
			&testFlags{},
		},
		{
			"flags definition from struct",
			testFlags{},
		},
	}

	expectedAnnotations := map[string][]string{flagusages.FlagGroupAnnotation: []string{"Config"}}

	for _, tc := range cases {
		suite.T().Run(tc.desc, func(t *testing.T) {
			c := &cobra.Command{}
			Define(c, tc.input, "")
			f := c.Flags()

			assert.NotNil(t, f.Lookup("loglevel"))
			assert.NotNil(t, f.Lookup("endpoint"))
			assert.NotNil(t, f.Lookup("timeout"))
			assert.Equal(t, expectedAnnotations, f.Lookup("loglevel").Annotations)
			assert.Equal(t, expectedAnnotations, f.Lookup("endpoint").Annotations)
			assert.Equal(t, expectedAnnotations, f.Lookup("timeout").Annotations)
			assert.Equal(t, "set the logging level", f.Lookup("loglevel").Usage)
			assert.Equal(t, "the listen.dev endpoint emitting the verdicts", f.Lookup("endpoint").Usage)
			assert.Equal(t, "set the timeout, in seconds", f.Lookup("timeout").Usage)

			assert.NotNil(t, f.Lookup("json"))
			assert.NotNil(t, f.Lookup("jq"))
			assert.NotNil(t, f.ShorthandLookup("q"))
			assert.Equal(t, "output the verdicts (if any) in JSON form", f.Lookup("json").Usage)
			assert.Equal(t, "filter the output using a jq expression", f.Lookup("jq").Usage)
		})
	}
}
