/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package flags

import (
	"context"
	"fmt"

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
)

type InOptions struct {
	Json bool   `name:"json"`
	JQ   string `name:"jq" validate:"excluded_without=Json,jq"` // TODO > set default to empty string (valid JQ query) to obtain JSON pretty print for free?

	baseOptions
}

func NewInOptions() (*InOptions, error) {
	o := &InOptions{}

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return o, nil
}

func (o *InOptions) Attach(c *cobra.Command) {
	c.Flags().BoolVar(&o.Json, "json", o.Json, "output the verdicts (if any) in JSON form")
	c.Flags().StringVarP(&o.JQ, "jq", "q", o.JQ, "filter the output using a jq expression")

	// TODO > There's no need to append a pre-run function if we are not actually using one at the moment
	// previousPreRun := c.PreRunE
	// c.PreRunE = func(c *cobra.Command, args []string) error {
	// 	// Run existing pre run (if any)
	// 	if previousPreRun != nil {
	// 		if err := previousPreRun(c, args); err != nil {
	// 			return err
	// 		}
	// 	}

	// 	// TODO > Check
	// 	// Assuming the JQ query will run on a struct listen.Response (parametrize this),
	// 	// we should get the field names of such struct (to which depth?) and check that the JQ query only uses them.
	// 	// Otherwise return an error here interrupting the command execution.

	// 	return nil
	// }
}

func (o *InOptions) Validate() []error {
	return o.baseOptions.Validate(o)
}

func (o *InOptions) Transform(ctx context.Context) error {
	return o.baseOptions.Transform(ctx, o)
}
