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

	"github.com/creasty/defaults"
	"github.com/spf13/cobra"
)

type ToOptions struct {
	JSONFlags

	baseOptions
}

func NewToOptions() (*ToOptions, error) {
	o := &ToOptions{}

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return o, nil
}

func (o *ToOptions) Attach(c *cobra.Command) {
	o.JSONFlags.Attach(c)
}

func (o *ToOptions) Validate() []error {
	return o.baseOptions.Validate(o)
}

func (o *ToOptions) Transform(ctx context.Context) error {
	return o.baseOptions.Transform(ctx, o)
}
