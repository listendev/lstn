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
	"strings"

	"github.com/XANi/goneric"
	"github.com/creasty/defaults"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"golang.org/x/exp/maps"
)

type Scan struct {
	ignore *enumflag.EnumFlagValue[npm.DependencyType]

	Ignores []npm.DependencyType
	flags.JSONFlags
	flags.ConfigFlags `flagset:"Config"`
}

func NewScan() (*Scan, error) {
	o := &Scan{}

	ignoreValues := `(` + strings.Join(goneric.MapSlice(func(t npm.DependencyType) string { return t.String() }, maps.Keys(npm.DependencyTypeIDs)), ",") + `)`
	o.ignore = enumflag.NewSlice(&o.Ignores, ignoreValues, npm.DependencyTypeIDs, enumflag.EnumCaseInsensitive)

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting defaults for the scan options")
	}
	if err := o.ignore.Set(npm.BundleDependencies.String()); err != nil {
		return nil, fmt.Errorf("error setting defaults for the scan options")
	}

	return o, nil
}

func (o *Scan) Attach(c *cobra.Command) {
	flags.Define(c, o, "")
	c.Flags().VarP(o.ignore, "ignore", "i", "sets of dependencies to ignore")
	c.Flags().Lookup("ignore").DefValue = `"` + npm.DependencyTypeIDs[npm.BundleDependencies][0] + `"`
	flagusages.Set(c)
}

func (o *Scan) Validate() []error {
	return flags.Validate(o)
}

func (o *Scan) Transform(ctx context.Context) error {
	return flags.Transform(ctx, o)
}
