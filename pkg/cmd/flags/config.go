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
	"reflect"

	"github.com/creasty/defaults"
)

type ConfigOptions struct {
	LogLevel string `default:"info" name:"log level" flag:"loglevel"` // TODO > validator
	Timeout  int    `default:"60" name:"timeout" flag:"timeout" validate:"number,min=30"`
	Endpoint string `default:"http://127.0.0.1:3000" flag:"endpoint" name:"endpoint" validate:"url,endpoint" transform:"tsuffix=/"`

	baseOptions
}

func NewConfigOptions() (*ConfigOptions, error) {
	o := &ConfigOptions{}

	if err := defaults.Set(o); err != nil {
		return nil, fmt.Errorf("error setting configuration defaults")
	}

	return o, nil
}

func (o *ConfigOptions) Validate() []error {
	return o.baseOptions.Validate(o)
}

func (o *ConfigOptions) Transform(ctx context.Context) error {
	return o.baseOptions.Transform(ctx, o)
}

func (o *ConfigOptions) GetField(name string) reflect.Value {
	return reflect.ValueOf(o).Elem().FieldByName(name)
}

func GetConfigFlagsNames() map[string]string {
	ret := make(map[string]string)
	t := reflect.TypeOf(ConfigOptions{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("flag")
		if tag != "" {
			ret[tag] = field.Name
		}
	}

	return ret
}

func GetConfigFlagsDefaults() map[string]string {
	ret := make(map[string]string)
	e := reflect.TypeOf(ConfigOptions{})
	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)
		tag := field.Tag.Get("default")
		if tag != "" {
			ret[field.Tag.Get("flag")] = tag
		}
	}

	return ret
}
