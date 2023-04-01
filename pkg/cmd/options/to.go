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

var _ cmd.CommandOptions = (*To)(nil)

type To struct {
	flags.DebugFlags `flagset:"Debug"`
	flags.JSONFlags
	flags.ConfigFlags
}

func NewTo() (*To, error) {
	o := &To{}

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return o, nil
}

func (o *To) Attach(c *cobra.Command, exclusions []string) {
	flags.Define(c, o, "", exclusions)
	flagusages.Set(c)
}

func (o *To) Validate() []error {
	return flags.Validate(o)
}

func (o *To) Transform(ctx context.Context) error {
	return flags.Transform(ctx, o)
}

func (o *To) AsJSON() string {
	return flags.AsJSON(o)
}
