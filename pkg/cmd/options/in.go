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

var _ cmd.Options = (*In)(nil)

type In struct {
	flags.JSONFlags
	flags.ConfigFlags
	flags.DebugFlags `flagset:"Debug"`
}

func NewIn() (*In, error) {
	o := &In{}

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return o, nil
}

func (o *In) Attach(c *cobra.Command) {
	flags.Define(c, o, "")
	flagusages.Set(c)
}

func (o *In) Validate() []error {
	return flags.Validate(o)
}

func (o *In) Transform(ctx context.Context) error {
	return flags.Transform(ctx, o)
}

func (o *In) AsJSON() string {
	return flags.AsJSON(o)
}
