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
	"io"

	"github.com/listendev/lstn/pkg/cmd/report/templates"
	"github.com/listendev/lstn/pkg/listen"
)

type FullMarkdwonReport struct {
	output io.Writer
}

func NewFullMarkdwonReport() *FullMarkdwonReport {
	return &FullMarkdwonReport{}
}

func (r *FullMarkdwonReport) WithOutput(w io.Writer) {
	r.output = w
}

func (r *FullMarkdwonReport) Render(packages []listen.Package) error {
	return templates.RenderContainer(r.output, packages)
}
