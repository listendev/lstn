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
package viper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
	"golang.org/x/exp/constraints"
)

func HandleEnumFlagPrecedence[T constraints.Integer](v reflect.Value, flag *pflag.Flag, defaultVal string) error {
	if flag == nil {
		return fmt.Errorf("cannot handle nil flag")
	}

	flagName := flag.Name

	// Store the flag value (it equals to the default when no flag)
	enumFlag, _ := flag.Value.(*enumflag.EnumFlagValue[T])
	flagValue := enumFlag.String()

	// Set the value coming from environment variable or config file (viper)
	value := viper.GetString(flagName)
	if value != "[]" && value != defaultVal {
		reportTypeErr := enumFlag.Set(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
		if reportTypeErr != nil {
			return fmt.Errorf("%s %s; got %s", flagName, reportTypeErr.Error(), value)
		}
		// Substitute the slice
		v.Set(reflect.ValueOf(enumFlag.Get()))
	}

	// Use the flag slice value when it's not empty
	if len(flagValue) > 0 && flagValue != "[]" {
		reportTypeErr := enumFlag.Set(strings.TrimSuffix(strings.TrimPrefix(flagValue, "["), "]"))
		if reportTypeErr != nil {
			return fmt.Errorf("%s %s; got %s", flagName, reportTypeErr.Error(), flagValue)
		}
		// Substitute the slice
		v.Set(reflect.ValueOf(enumFlag.Get()))
	}

	return nil
}
