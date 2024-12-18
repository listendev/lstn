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

	"github.com/ghetzel/testify/require"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	npmdeptype "github.com/listendev/lstn/pkg/npm/deptype"
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
	assert.Len(suite.T(), res, 18)
}

func (suite *FlagsBaseSuite) TestGetDefaults() {
	type ScanOpts struct {
		ConfigFlags
		JSONFlags
	}
	res := GetDefaults(&ScanOpts{})

	assert.Len(suite.T(), res, 8)
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

func (suite *FlagsBaseSuite) TestGetFieldTag() {
	cfg := ConfigFlags{}

	tokenGithubTag, tokenGithubTagFound := GetFieldTag(cfg, "Token.GitHub")
	require.True(suite.T(), tokenGithubTagFound)
	tokeGithubNameTag, tokeGithubNameTagFound := tokenGithubTag.Lookup("name")
	require.True(suite.T(), tokeGithubNameTagFound)
	require.Equal(suite.T(), "GitHub token", tokeGithubNameTag)
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
			[]string{"timeout must be 30 or greater", "NPM endpoint must be a valid URL", "PyPi endpoint must be a valid URL", "Core API must be a valid URL"},
		},
		{
			"invalid timeout",
			&ConfigFlags{Timeout: 29, Endpoint: Endpoint{Npm: "http://127.0.0.1:3000", PyPi: "http://127.0.0.1:3001", Core: "http://127.0.0.1:3002"}},
			[]string{"timeout must be 30 or greater"},
		},
		{
			"invalid NPM endpoint",
			&ConfigFlags{Timeout: 31, Endpoint: Endpoint{Npm: "http://invalid.endpoint", PyPi: "http://127.0.0.1:3001", Core: "http://127.0.0.1:3002"}},
			[]string{"NPM endpoint must be a valid listen.dev endpoint"},
		},
		{
			"invalid PyPi endpoint",
			&ConfigFlags{Timeout: 31, Endpoint: Endpoint{PyPi: "http://invalid.endpoint", Npm: "http://127.0.0.1:3001", Core: "http://127.0.0.1:3002"}},
			[]string{"PyPi endpoint must be a valid listen.dev endpoint"},
		},
		{
			"valid config flags",
			&ConfigFlags{Timeout: 31, Endpoint: Endpoint{Npm: "http://127.0.0.1:3000", PyPi: "http://127.0.0.1:3000", Core: "http://127.0.0.1:3002"}},
			[]string{},
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.desc, func(_ *testing.T) {
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
			&ConfigFlags{
				Reporting: Reporting{
					Types: []cmd.ReportType{},
				},
				Filtering: Filtering{
					Ignore: Ignore{
						Deptypes: []npmdeptype.Enum{},
					},
				},
			},
			nil,
		},
		{
			"npm endpoint config with leading slash",
			&ConfigFlags{
				Endpoint: Endpoint{Npm: "https://npm.listen.dev/"},
			},
			&ConfigFlags{
				Endpoint: Endpoint{Npm: "https://npm.listen.dev"},
				Reporting: Reporting{
					Types: []cmd.ReportType{},
				},
				Filtering: Filtering{
					Ignore: Ignore{
						Deptypes: []npmdeptype.Enum{},
					},
				},
			},
			nil,
		},
		{
			"pypi endpoint config with leading slash",
			&ConfigFlags{
				Endpoint: Endpoint{PyPi: "https://pypi.listen.dev/"},
			},
			&ConfigFlags{
				Endpoint: Endpoint{PyPi: "https://pypi.listen.dev"},
				Reporting: Reporting{
					Types: []cmd.ReportType{},
				},
				Filtering: Filtering{
					Ignore: Ignore{
						Deptypes: []npmdeptype.Enum{},
					},
				},
			},
			nil,
		},
		{
			"endpoints config with leading slash",
			&ConfigFlags{
				Endpoint: Endpoint{PyPi: "https://pypi.listen.dev/", Npm: "https://npm.listen.dev/"},
			},
			&ConfigFlags{
				Endpoint: Endpoint{PyPi: "https://pypi.listen.dev", Npm: "https://npm.listen.dev"},
				Reporting: Reporting{
					Types: []cmd.ReportType{},
				},
				Filtering: Filtering{
					Ignore: Ignore{
						Deptypes: []npmdeptype.Enum{},
					},
				},
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
				Reporting: Reporting{
					Types: []cmd.ReportType{},
				},
				Filtering: Filtering{
					Ignore: Ignore{
						Deptypes: []npmdeptype.Enum{},
					},
				},
			},
			nil,
		},
		{
			"duplicate packages to ignore",
			&ConfigFlags{
				Filtering: Filtering{
					Ignore: Ignore{
						Packages: []string{"a", "a"},
					},
				},
			},
			&ConfigFlags{
				Filtering: Filtering{
					Ignore: Ignore{
						Packages: []string{"a"},
						Deptypes: []npmdeptype.Enum{},
					},
				},
				Reporting: Reporting{
					Types: []cmd.ReportType{},
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

	expectedAnnotations := map[string][]string{flagusages.FlagGroupAnnotation: {"Config"}}

	for _, tc := range cases {
		suite.T().Run(tc.desc, func(t *testing.T) {
			c := &cobra.Command{}
			Define(c, tc.input, "", []string{})
			f := c.Flags()

			assert.NotNil(t, f.Lookup("loglevel"))
			assert.NotNil(t, f.Lookup("npm-endpoint"))
			assert.NotNil(t, f.Lookup("timeout"))
			assert.Equal(t, expectedAnnotations, f.Lookup("loglevel").Annotations)
			assert.Equal(t, expectedAnnotations, f.Lookup("npm-endpoint").Annotations)
			assert.Equal(t, expectedAnnotations, f.Lookup("timeout").Annotations)
			assert.Equal(t, "set the logging level", f.Lookup("loglevel").Usage)
			assert.Equal(t, "the listen.dev endpoint emitting the NPM verdicts", f.Lookup("npm-endpoint").Usage)
			assert.Equal(t, "set the timeout, in seconds", f.Lookup("timeout").Usage)

			assert.NotNil(t, f.Lookup("json"))
			assert.NotNil(t, f.Lookup("jq"))
			assert.NotNil(t, f.ShorthandLookup("q"))
			assert.Equal(t, "output the verdicts (if any) in JSON form", f.Lookup("json").Usage)
			assert.Equal(t, "filter the output verdicts using a jq expression (requires --json)", f.Lookup("jq").Usage)
		})
	}
}
