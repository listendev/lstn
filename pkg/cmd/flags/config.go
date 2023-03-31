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
	"strings"

	"github.com/XANi/goneric"
	"github.com/creasty/defaults"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
)

var _ cmd.Options = (*ConfigFlags)(nil)

type Token struct {
	GitHub string `name:"GitHub token" flag:"gh-token" desc:"set the GitHub token" flagset:"Token" json:"gh-token"`
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

type Reporter struct {
	reporter *enumflag.EnumFlagValue[cmd.ReportType]

	Types []cmd.ReportType `json:"reporter" flag:"reporter"`
	GitHub
}

type Ignore struct {
	Packages []string `name:"ignore packages" flag:"ignore-packages" desc:"list of packages to not process" json:"ignore-packages" flagset:"Filtering"`
}

type ConfigFlags struct {
	LogLevel string `default:"info" name:"log level" flag:"loglevel" desc:"set the logging level" flagset:"Config" json:"loglevel"`                          // TODO > validator
	Timeout  int    `default:"60" name:"timeout" flag:"timeout" desc:"set the timeout, in seconds" validate:"number,min=30" flagset:"Config" json:"timeout"` // FIXME: change to time.Duration type
	Endpoint string `default:"https://npm.listen.dev" flag:"endpoint" name:"endpoint" desc:"the listen.dev endpoint emitting the verdicts" validate:"url,endpoint" transform:"tsuffix=/" flagset:"Config" json:"endpoint"`
	Token
	Registry
	Reporter
	Ignore
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
		if defaults.CanUpdate(o.Reporter.GitHub.Owner) {
			o.Reporter.GitHub.Owner = env.Owner
		}
		if defaults.CanUpdate(o.Reporter.GitHub.Repo) {
			o.Reporter.GitHub.Repo = env.Repo
		}
		if defaults.CanUpdate(o.Reporter.GitHub.Pull.ID) && env.IsPullRequest() {
			o.Reporter.GitHub.Pull.ID = env.Num
		}
	}
}

func (o *ConfigFlags) Define(c *cobra.Command) {
	// Create the enum flag value for --reporter
	enumValues := goneric.MapToSlice(func(t cmd.ReportType, v []string) string {
		return v[0]
	}, cmd.ReporterTypeIDs)
	sort.Strings(enumValues)
	o.Reporter.reporter = enumflag.NewSlice(&o.Reporter.Types, `(`+strings.Join(enumValues, ",")+`)`, cmd.ReporterTypeIDs, enumflag.EnumCaseInsensitive)

	// Manually define the --reporter enum flag
	c.Flags().VarP(o.Reporter.reporter, "reporter", "r", `set one or more reporters to use`)
	_ = c.Flags().SetAnnotation("reporter", flagusages.FlagGroupAnnotation, []string{"Reporting"})
}

func (o *ConfigFlags) Validate() []error {
	return Validate(o)
}

func (o *ConfigFlags) Transform(ctx context.Context) error {
	return Transform(ctx, o)
}
