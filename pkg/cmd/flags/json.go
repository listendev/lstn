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
	"io"

	"github.com/listendev/lstn/pkg/jq"
)

type JSONFlags struct {
	JSON bool   `name:"json" flag:"json" desc:"output the verdicts (if any) in JSON form" json:"json"`
	JQ   string `flagset:"Filtering" name:"jq" flag:"jq" shorthand:"q" desc:"filter the output verdicts using a jq expression (requires --json)" validate:"excluded_without=JSON,jq" json:"jq"`
}

func (o *JSONFlags) IsJSON() bool {
	return o.JSON
}

func (o *JSONFlags) GetQuery() string {
	return o.JQ
}

func (o *JSONFlags) GetOutput(ctx context.Context, input io.Reader, output io.Writer) error {
	if o.IsJSON() {
		return jq.Eval(ctx, input, output, o.GetQuery())
	}

	return fmt.Errorf("cannot output JSON")
}
