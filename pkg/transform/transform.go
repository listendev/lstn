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
package transform

import (
	"context"
	"reflect"

	"github.com/XANi/goneric"
	"github.com/go-playground/mold/v4"
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/listendev/lstn/pkg/cmd"
	npmdeptype "github.com/listendev/lstn/pkg/npm/deptype"
)

// Singleton is the mold.Transformer singleton instance.
var Singleton *mold.Transformer

func init() {
	Singleton = modifiers.New()
	Singleton.SetTagName("transform")

	Singleton.Register("unique", func(_ context.Context, fl mold.FieldLevel) error {
		if fl.Field().Kind() == reflect.Slice {
			unique := fl.Field().Interface()
			switch fl.Field().Type().Elem().String() {
			case "string":
				unique = goneric.SliceDedupe(unique.([]string))
			case "cmd.ReportType":
				intermediate := []string{}
				for _, i := range unique.([]cmd.ReportType) {
					intermediate = append(intermediate, i.String())
				}
				unique, _ = cmd.ParseReportTypes(goneric.SliceDedupe(intermediate))
			case "deptype.Enum":
				intermediate := []string{}
				for _, i := range unique.([]npmdeptype.Enum) {
					intermediate = append(intermediate, i.String())
				}
				unique, _ = npmdeptype.ParseMultiple(goneric.SliceDedupe(intermediate)...)
			}

			fl.Field().Set(reflect.ValueOf(unique))
		}

		return nil
	})
}
