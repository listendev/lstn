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
package validate

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/listendev/lstn/pkg/jq"
)

func jqQueryCompiles(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() == reflect.String {
		_, err := jq.Compile(field.String())

		return err == nil
	}

	panic(fmt.Sprintf("bad field type: %T", field.Interface()))
}
