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
	"maps"
	"slices"
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
)

var _ cmd.Options = (*ConfigFlags)(nil)

type Token struct {
	GitHub string `desc:"set the GitHub token"          flag:"gh-token"  flagset:"Token" json:"gh-token"  name:"GitHub token" validate:"omitempty,notblank"`
	JWT    string `desc:"set the listen.dev auth token" flag:"jwt-token" flagset:"Token" json:"jwt-token" name:"JWT token"    validate:"omitempty,notblank"`
}

type TokenMandatory struct {
	GitHub string `desc:"set the GitHub token"          flag:"gh-token"  flagset:"Token" json:"gh-token"  name:"GitHub token" validate:"mandatory"`
	JWT    string `desc:"set the listen.dev auth token" flag:"jwt-token" flagset:"Token" json:"jwt-token" name:"JWT token"    validate:"mandatory"`
}

type Registry struct {
	NPM string `default:"https://registry.npmjs.org" desc:"set a custom NPM registry" flag:"npm-registry" flagset:"Registry" json:"npm-registry" name:"NPM registry" transform:"tsuffix=/" validate:"omitempty,url"`
}

type Pull struct {
	ID int `desc:"set the GitHub pull request ID" flag:"gh-pull-id" flagset:"Reporting" json:"gh-pull-id" name:"github pull request ID"`
}

type GitHub struct {
	Owner string `desc:"set the GitHub owner name (org|user)" flag:"gh-owner" flagset:"Reporting" json:"gh-owner" name:"github owner"`
	Repo  string `desc:"set the GitHub repository name"       flag:"gh-repo"  flagset:"Reporting" json:"gh-repo"  name:"github repository"`
	Pull
}

// NOTE > Struct can't have the same name of a flag.
type Reporting struct {
	reporter *enumflag.EnumFlagValue[cmd.ReportType]

	Types []cmd.ReportType `desc:"set one or more reporters to use" flag:"reporter" flagset:"Reporting" json:"reporter" shorthand:"r" transform:"unique"`
	GitHub
}

type Ignore struct {
	Packages []string          `default:"[]"                                         desc:"the list of packages to not process" flag:"ignore-packages" json:"ignore-packages"         name:"ignore packages" transform:"unique"`
	Deptypes []npmdeptype.Enum `desc:"the list of dependencies types to not process" flag:"ignore-deptypes"                     json:"ignore-deptypes" name:"ignore dependency types" transform:"unique"`

	types *enumflag.EnumFlagValue[npmdeptype.Enum]
}

type Filtering struct {
	Ignore     `flagset:"Filtering"`
	Expression string `desc:"filter the output verdicts using a jsonpath script expression (server-side)" flag:"select" flagset:"Filtering" json:"select" name:"filter verdicts" shorthand:"s"`
}

type Endpoint struct {
	Npm  string `default:"https://npm.listen.dev"  desc:"the listen.dev endpoint emitting the NPM verdicts"  flag:"npm-endpoint"  flagset:"Config" json:"npm"  name:"NPM endpoint"  transform:"tsuffix=/" validate:"url,endpoint"`
	PyPi string `default:"https://pypi.listen.dev" desc:"the listen.dev endpoint emitting the PyPi verdicts" flag:"pypi-endpoint" flagset:"Config" json:"pypi" name:"PyPi endpoint" transform:"tsuffix=/" validate:"url,endpoint"`
	Core string `default:"https://core.listen.dev" desc:"the listen.dev Core API endpoint"                   flag:"core-endpoint" flagset:"Config" json:"core" name:"Core API"      transform:"tsuffix=/" validate:"url"`
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
	LogLevel string   `default:"info"  desc:"set the logging level"       flag:"loglevel" flagset:"Config" json:"loglevel" name:"log level"`                          // TODO > validator
	Timeout  int      `default:"60"    desc:"set the timeout, in seconds" flag:"timeout"  flagset:"Config" json:"timeout"  name:"timeout"   validate:"number,min=30"` // FIXME: change to time.Duration type
	Endpoint Endpoint `json:"endpoint"`
	Token
	Registry
	Reporting
	Filtering
	Lockfiles []string `default:"[\"package-lock.json\",\"poetry.lock\"]" desc:"set one or more lock file paths (relative to the working dir) to lookup for" flag:"lockfiles" json:"lockfiles" shorthand:"l" transform:"unique"`
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
			slices.Collect(maps.Values(npmdeptype.IDs)),
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
