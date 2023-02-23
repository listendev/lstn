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
)

type ConfigFlags struct {
	LogLevel string `default:"info" name:"log level" flag:"loglevel" desc:"set the logging level"`                           // TODO > validator
	Timeout  int    `default:"60" name:"timeout" flag:"timeout" desc:"set the timeout, in seconds" validate:"number,min=30"` // FIXME: change to time.Duration type
	Endpoint string `default:"https://npm.listen.dev" flag:"endpoint" name:"endpoint" desc:"the listen.dev endpoint emitting the verdicts" validate:"url,endpoint" transform:"tsuffix=/"`
}

func NewConfigFlags() (*ConfigFlags, error) {
	o := &ConfigFlags{}

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return o, nil
}

func (o *ConfigFlags) Validate() []error {
	return Validate(o)
}

func (o *ConfigFlags) Transform(ctx context.Context) error {
	return Transform(ctx, o)
}
