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
package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/listendev/lstn/pkg/listen"
)

type JSONReport struct {
	output io.Writer
}

func NewJSONReport() *JSONReport {
	return &JSONReport{}
}

func (r *JSONReport) WithOutput(w io.Writer) {
	r.output = w
}

func (r *JSONReport) Render(packages []listen.Package) error {
	enc := json.NewEncoder(r.output)
	enc.SetIndent("", "  ")
	err := enc.Encode(packages)
	if err != nil {
		return fmt.Errorf("couldn't encode the JSON report: %w", err)
	}

	return nil
}
