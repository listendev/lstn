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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/XANi/goneric"
	"github.com/creasty/defaults"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd"
	npmdeptype "github.com/listendev/lstn/pkg/npm/deptype"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"golang.org/x/exp/maps"
)

var _ cmd.Options = (*ConfigFlags)(nil)

type Token struct {
	GitHub string `name:"GitHub token" flag:"gh-token" desc:"set the GitHub token" flagset:"Token" json:"gh-token" validate:"omitempty,notblank"`
	JWT    string `name:"JWT token" flag:"jwt-token" desc:"set the listen.dev auth token" flagset:"Token" json:"jwt-token" validate:"omitempty,notblank"`
}

type TokenMandatory struct {
	GitHub string `name:"GitHub token" flag:"gh-token" desc:"set the GitHub token" flagset:"Token" json:"gh-token" validate:"mandatory"`
	JWT    string `name:"JWT token" flag:"jwt-token" desc:"set the listen.dev auth token" flagset:"Token" json:"jwt-token" validate:"mandatory"`
}

type Registry struct {
	NPM string `name:"NPM registry" flag:"npm-registry" desc:"set a custom NPM registry" validate:"omitempty,url" default:"https://registry.npmjs.org" transform:"tsuffix=/" flagset:"Registry" json:"npm-registry"`
}

type Pull struct {
	ID int `name:"github pull request ID" flag:"gh-pull-id" desc:"set the GitHub pull request ID" flagset:"Reporting" json:"gh-pull-id"`
}

type GitHub struct {
	Owner string `name:"github owner" flag:"gh-owner" desc:"set the GitHub owner name (org|user)" flagset:"Reporting" json:"gh-owner"`
	Repo  string `name:"github repository" flag:"gh-repo" desc:"set the GitHub repository name" flagset:"Reporting" json:"gh-repo"`
	Pull
}

// NOTE > Struct can't have the same name of a flag.
type Reporting struct {
	reporter *enumflag.EnumFlagValue[cmd.ReportType]

	Types []cmd.ReportType `json:"reporter" flag:"reporter" shorthand:"r" transform:"unique" desc:"set one or more reporters to use" flagset:"Reporting"`
	GitHub
}

type Ignore struct {
	Packages []string          `name:"ignore packages" flag:"ignore-packages" desc:"the list of packages to not process" transform:"unique" default:"[]" json:"ignore-packages"`
	Deptypes []npmdeptype.Enum `name:"ignore dependency types" flag:"ignore-deptypes" desc:"the list of dependencies types to not process" transform:"unique" json:"ignore-deptypes"`

	types *enumflag.EnumFlagValue[npmdeptype.Enum]
}

type Filtering struct {
	Ignore     `flagset:"Filtering"`
	Expression string `flagset:"Filtering" name:"filter verdicts" flag:"select" shorthand:"s" desc:"filter the output verdicts using a jsonpath script expression (server-side)" json:"select"`
}

type Endpoint struct {
	Npm  string `default:"https://npm.listen.dev" flag:"npm-endpoint" name:"NPM endpoint" desc:"the listen.dev endpoint emitting the NPM verdicts" validate:"url,endpoint" transform:"tsuffix=/" flagset:"Config" json:"npm"`
	PyPi string `default:"https://pypi.listen.dev" flag:"pypi-endpoint" name:"PyPi endpoint" desc:"the listen.dev endpoint emitting the PyPi verdicts" validate:"url,endpoint" transform:"tsuffix=/" flagset:"Config" json:"pypi"`
	Core string `default:"https://core.listen.dev" flag:"core-endpoint" name:"Core API" desc:"the listen.dev Core API endpoint" validate:"url" transform:"tsuffix=/" flagset:"Config" json:"core"`
}

// IsLocalCore returns true if the Core API endpoint is a local one.
// It is considered local and endpoint with fixed IP address and http scheme.
func (e Endpoint) IsLocalCore() bool {
	isHTTP := strings.HasPrefix(e.Core, "http://")
	fixedIP := false

	address := strings.TrimPrefix(e.Core, "http://")

	nums := strings.Split(address, ".")
	if len(nums) == 4 {
		for _, num := range nums {
			n, err := strconv.Atoi(num)
			if err != nil {
				return false
			}

			if n >= 0 && n <= 255 {
				fixedIP = true
			} else {
				return false
			}
		}
	}

	return isHTTP && fixedIP
}

// ConfigFlags are the options that the CLI also reads from the YAML configuration file.
type ConfigFlags struct {
	LogLevel string   `default:"info" name:"log level" flag:"loglevel" desc:"set the logging level" flagset:"Config" json:"loglevel"`                          // TODO > validator
	Timeout  int      `default:"60" name:"timeout" flag:"timeout" desc:"set the timeout, in seconds" validate:"number,min=30" flagset:"Config" json:"timeout"` // FIXME: change to time.Duration type
	Endpoint Endpoint `json:"endpoint"`
	Token
	Registry
	Reporting
	Filtering
	Lockfiles []string `json:"lockfiles" flag:"lockfiles" shorthand:"l" transform:"unique" desc:"set one or more lock file paths (relative to the working dir) to lookup for" default:"[\"package-lock.json\",\"poetry.lock\"]"`
}

func NewConfigFlags() (*ConfigFlags, error) {
	o := &ConfigFlags{}

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return o, nil
}

func (o *ConfigFlags) SetDefaults() {
	// Attempt to dynamically set the defaults for the GitHub reporting flags from the environment
	env, err := ci.NewInfo()
	if err == nil && env != nil {
		if defaults.CanUpdate(o.Reporting.GitHub.Owner) {
			o.Reporting.GitHub.Owner = env.Owner
		}
		if defaults.CanUpdate(o.Reporting.GitHub.Repo) {
			o.Reporting.GitHub.Repo = env.Repo
		}
		if defaults.CanUpdate(o.Reporting.GitHub.Pull.ID) && env.IsGitHubPullRequest() {
			o.Reporting.GitHub.Pull.ID = env.Num
		}
	}
	if defaults.CanUpdate(o.Reporting.Types) {
		// Create the enum flag value for --reporter
		enumValues := goneric.MapToSlice(func(_ cmd.ReportType, v []string) string {
			return v[0]
		}, cmd.ReporterTypeIDs)
		sort.Strings(enumValues)
		o.Reporting.reporter = enumflag.NewSlice(&o.Reporting.Types, `(`+strings.Join(enumValues, ",")+`)`, cmd.ReporterTypeIDs, enumflag.EnumCaseInsensitive)
	}
	if defaults.CanUpdate(o.Filtering.Ignore.types) {
		// Create the enum flag value for --ignore-deptypes
		alwaysInSet := npmdeptype.BundleDependencies
		ignoreValues := goneric.MapSliceSkip(
			func(identifiers []string) (string, bool) {
				t := identifiers[0]
				if t == alwaysInSet.String() {
					return "", true
				}

				return t, false
			},
			maps.Values(npmdeptype.IDs),
		)
		sort.Strings(ignoreValues)
		o.Filtering.Ignore.types = enumflag.NewSlice(&o.Filtering.Ignore.Deptypes, `(`+strings.Join(ignoreValues, ",")+`)`, npmdeptype.IDs, enumflag.EnumCaseInsensitive)
		_ = o.Filtering.Ignore.types.Set(alwaysInSet.String())
	}
}

func (o *ConfigFlags) Define(c *cobra.Command, exclusions []string) {
	if !goneric.SliceIn(exclusions, "reporter") {
		// Manually define the --reporter enum flag
		fld, found := getValue(o.Reporting).Type().FieldByName("Types")
		if found && o.Reporting.reporter != nil {
			desc := fld.Tag.Get("desc")
			flag := fld.Tag.Get("flag")
			shrt := fld.Tag.Get("shorthand")
			c.Flags().VarP(o.Reporting.reporter, flag, shrt, desc)
		}
	}
	if !goneric.SliceIn(exclusions, "ignore-deptypes") {
		// Manually define the --ignore-deptypes enum flag
		fld, found := getValue(o.Filtering.Ignore).Type().FieldByName("Deptypes")
		if found && o.Filtering.Ignore.types != nil {
			desc := fld.Tag.Get("desc")
			flag := fld.Tag.Get("flag")
			shrt := fld.Tag.Get("shorthand")
			c.Flags().VarP(o.Filtering.Ignore.types, flag, shrt, desc)
		}
	}
}

func (o *ConfigFlags) Validate() []error {
	return Validate(o)
}

func (o *ConfigFlags) Transform(ctx context.Context) error {
	return Transform(ctx, o)
}
