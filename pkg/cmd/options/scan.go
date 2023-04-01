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
package options

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/XANi/goneric"
	"github.com/creasty/defaults"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"golang.org/x/exp/maps"
)

var _ cmd.CommandOptions = (*Scan)(nil)

type Scan struct {
	exclude *enumflag.EnumFlagValue[npm.DependencyType]

	Excludes         []npm.DependencyType `json:"exclude"`
	flags.DebugFlags `flagset:"Debug"`
	flags.JSONFlags
	flags.ConfigFlags
}

func NewScan() (*Scan, error) {
	o := &Scan{}

	// Create the enum flag value for --exclude
	alwaysInSet := npm.BundleDependencies
	ignoreValues := goneric.MapSliceSkip(
		func(identifiers []string) (string, bool) {
			t := identifiers[0]
			if t == alwaysInSet.String() {
				return "", true
			}

			return t, false
		},
		maps.Values(npm.DependencyTypeIDs),
	)
	sort.Strings(ignoreValues)
	// Proxy values to o.Ignores
	o.exclude = enumflag.NewSlice(&o.Excludes, `(`+strings.Join(ignoreValues, ",")+`)`, npm.DependencyTypeIDs, enumflag.EnumCaseInsensitive)
	if err := o.exclude.Set(alwaysInSet.String()); err != nil {
		return nil, fmt.Errorf("error setting defaults for the scan options")
	}

	// Set defaults for other (normal) flags
	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting defaults for the scan options")
	}

	return o, nil
}

func (o *Scan) Attach(c *cobra.Command, exclusions []string) {
	flags.Define(c, o, "", exclusions)
	// Define --exclude enum flag manually
	c.Flags().VarP(o.exclude, "exclude", "e", `sets of dependencies to exclude (in addition to the default)`)
	flagusages.Set(c)
}

func (o *Scan) Validate() []error {
	return flags.Validate(o)
}

func (o *Scan) Transform(ctx context.Context) error {
	return flags.Transform(ctx, o)
}

func (o *Scan) AsJSON() string {
	return flags.AsJSON(o)
}
