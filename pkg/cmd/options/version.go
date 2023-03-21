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

	"github.com/creasty/defaults"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	"github.com/spf13/cobra"
)

var _ cmd.Options = (*Version)(nil)

type Version struct {
	Verbosity        int  `name:"verbosity" shorthand:"v" desc:"increment the verbosity level" type:"count" validate:"gte=0" json:"verbosity"`
	Changelog        bool `name:"changelog" flag:"changelog" desc:"output the relase notes URL" json:"changelog"`
	flags.DebugFlags `flagset:"Debug"`
}

func NewVersion() (*Version, error) {
	v := &Version{}

	if err := defaults.Set(v); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return v, nil
}

func (v *Version) Validate() []error {
	return flags.Validate(v)
}

func (v *Version) Transform(ctx context.Context) error {
	return flags.Transform(ctx, v)
}

func (v *Version) Attach(c *cobra.Command) {
	flags.Define(c, v, "")
	flagusages.Set(c)
}

func (o *Version) AsJSON() string {
	return flags.AsJSON(o)
}
