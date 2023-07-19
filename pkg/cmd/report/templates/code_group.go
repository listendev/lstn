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
package templates

import (
	"embed"
	"io"
	"text/template"

	"github.com/listendev/pkg/models"
	"github.com/listendev/pkg/models/severity"
)

//go:embed code_group.html
var tmpCodeGroup embed.FS

type cumulatedSeverities struct {
	High   int
	Medium int
	Low    int
}

func RenderCodeGroup(w io.Writer, code string, severitiesMap map[severity.Severity][]models.Verdict, i icons) error {
	tmplData, err := tmpCodeGroup.ReadFile("code_group.html")
	if err != nil {
		return err
	}

	tmpl, err := template.New("code_group").Parse(string(tmplData))
	if err != nil {
		return err
	}

	severitiesCounter := make(map[severity.Severity]int)
	for severity, verdicts := range severitiesMap {
		severitiesCounter[severity] += len(verdicts)
	}

	return tmpl.Execute(w, struct {
		Code                string
		Icons               icons
		CumulatedSeverities cumulatedSeverities
	}{
		Code:  code,
		Icons: i,
		CumulatedSeverities: cumulatedSeverities{
			High:   severitiesCounter[severity.High],
			Medium: severitiesCounter[severity.Medium],
			Low:    severitiesCounter[severity.Low],
		},
	})
}
