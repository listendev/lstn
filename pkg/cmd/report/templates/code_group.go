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
	"fmt"
	"io"
	"text/template"

	"github.com/listendev/pkg/models"
	"github.com/listendev/pkg/models/severity"
)

//go:embed code_group.html
var tmpCodeGroup embed.FS

type cumulatedSeverities struct {
	High           []models.Verdict
	Medium         []models.Verdict
	Low            []models.Verdict
	TotalAmount    int
	SingleSeverity *singleSeverity
}

type singleSeverity struct {
	Severity severity.Severity
	Verdicts []models.Verdict
	Label    string
	Icon     string
}

func newCumulatedSeverities(severityGroups map[severity.Severity][]models.Verdict, icons icons) cumulatedSeverities {
	m := make(map[severity.Severity][]models.Verdict)
	t := 0
	for severity, verdicts := range severityGroups {
		t += len(verdicts)
		m[severity] = append(m[severity], verdicts...)
	}

	cs := cumulatedSeverities{
		High:           m[severity.High],
		Medium:         m[severity.Medium],
		Low:            m[severity.Low],
		TotalAmount:    t,
		SingleSeverity: nil,
	}

	for s, v := range severityGroups {
		if len(v) == t {
			i := icons.LowSeverity

			if s == severity.Medium {
				i = icons.MediumSeverity
			}

			l := fmt.Sprintf("%s severity", s)
			if s == severity.High {
				l = "Critical severity"
				i = icons.HighSeverity
			}

			cs.SingleSeverity = &singleSeverity{s, v, l, i}
			break
		}
	}

	return cs
}

type codeData struct {
	Label string
	Icon  string
}

var codeDataMap = map[string]codeData{
	"UNK": {"Unknown", "👽"},
	"FNI": {"Dynamic instrumentation", "📡"},
	"TSN": {"Typosquatting", "🔀"},
	"MDN": {"Metadata", "📑"},
	"STN": {"Static analysis", "🔎"},
	"DDN": {"Advisories", "🛡️"},
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func RenderCodeGroup(w io.Writer, code string, severitiesMap map[severity.Severity][]models.Verdict, i icons) error {
	tmplData, err := tmpCodeGroup.ReadFile("code_group.html")
	if err != nil {
		return err
	}

	tmpl, err := template.New("code_group").Funcs(template.FuncMap{
		"pluralize": pluralize,
	}).Parse(string(tmplData))
	if err != nil {
		return err
	}

	return tmpl.Execute(w, struct {
		Code                string
		CodeData            codeData
		Icons               icons
		CumulatedSeverities cumulatedSeverities
	}{
		Code:                code,
		CodeData:            codeDataMap[code],
		Icons:               i,
		CumulatedSeverities: newCumulatedSeverities(severitiesMap, i),
	})
}
