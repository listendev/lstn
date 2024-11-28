// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2024 The listen.dev team <engineering@garnet.ai>
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
package context

import (
	"context"
	"fmt"
	"reflect"

	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flags"
)

// GetFromContext returns a Options instance.
//
// It also validates and transforms the Options instance it obtained.
//
// It errors out if the context key does not refer an Options instance
// or if the validation and trasformation process errored out.
func GetOptionsFromContext(ctx context.Context, key any) (cmd.Options, error) {
	o, ok := ctx.Value(key).(cmd.Options)
	if !ok {
		return nil, fmt.Errorf("the key does not refer an Options instance")
	}

	// Update the inner config options into the current option set
	// This makes the config options work as global options
	val := reflect.ValueOf(o).Elem()
	fld := val.FieldByName("ConfigFlags")
	if fld.IsValid() && fld.CanSet() {
		cfgOpts, ok := ctx.Value(ConfigKey).(*flags.ConfigFlags)
		if !ok {
			return nil, fmt.Errorf("couldn't obtain the config options to update the current option set")
		}
		fld.Set(reflect.ValueOf(cfgOpts).Elem())
	}

	if errors := o.Validate(); errors != nil {
		ret := "invalid options"
		for _, e := range errors {
			ret += "\n       "
			ret += e.Error()
		}

		return nil, fmt.Errorf("%s", ret)
	}

	// Transform the config options values
	if err := o.Transform(ctx); err != nil {
		return nil, err
	}

	return o, nil
}
