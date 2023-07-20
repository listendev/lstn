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
	"github.com/listendev/pkg/verdictcode"
)

//go:embed code_group.html
var tmpCodeGroup embed.FS

type cumulatedSeverities struct {
	Severities     map[severity.Severity]map[string]map[verdictcode.Code]TransitivesAndNon
	High           []models.Verdict
	Medium         []models.Verdict
	Low            []models.Verdict
	TotalAmount    int
	SingleSeverity *singleSeverity
}

type TransitivesAndNon struct {
	Transitive    []models.Verdict
	NonTransitive []models.Verdict
}

type singleSeverity struct {
	Severity severity.Severity
	Verdicts []models.Verdict
	Label    string
	Icon     string
}

func newCumulatedSeverities(severityGroups map[severity.Severity][]models.Verdict, icons map[string]string) cumulatedSeverities {
	m := make(map[severity.Severity][]models.Verdict)
	t := 0
	for severity, verdicts := range severityGroups {
		t += len(verdicts)
		m[severity] = append(m[severity], verdicts...)
	}

	M := make(map[severity.Severity]map[string]map[verdictcode.Code]TransitivesAndNon)
	M[severity.High] = make(map[string]map[verdictcode.Code]TransitivesAndNon)
	M[severity.Medium] = make(map[string]map[verdictcode.Code]TransitivesAndNon)
	M[severity.Low] = make(map[string]map[verdictcode.Code]TransitivesAndNon)
	for s, vs := range severityGroups {
		S := M[s]

		for _, v := range vs {
			nvKey := fmt.Sprintf("%s/%s", v.Pkg, v.Version)
			nv, e := S[nvKey]
			if !e {
				nv = make(map[verdictcode.Code]TransitivesAndNon)
				S[nvKey] = nv
			}
			c, e := nv[v.Code]
			if !e {
				c = TransitivesAndNon{
					Transitive:    []models.Verdict{},
					NonTransitive: []models.Verdict{},
				}
			}
			if _, ok := v.Metadata["npm_package_name"]; !ok {
				c.NonTransitive = append(c.NonTransitive, v)
				nv[v.Code] = c
				continue
			}
			if _, ok := v.Metadata["npm_package_version"]; !ok {
				c.NonTransitive = append(c.NonTransitive, v)
				nv[v.Code] = c // Save the updated c back to nv
				continue
			}
			if v.Metadata["npm_package_name"] != v.Pkg && v.Metadata["npm_package_version"] != v.Version {
				c.Transitive = append(c.Transitive, v)
				nv[v.Code] = c // Save the updated c back to nv
				continue
			}
			c.NonTransitive = append(c.NonTransitive, v)
			nv[v.Code] = c
		}
	}

	cs := cumulatedSeverities{
		Severities:     M,
		High:           m[severity.High],
		Medium:         m[severity.Medium],
		Low:            m[severity.Low],
		TotalAmount:    t,
		SingleSeverity: nil,
	}

	for s, v := range severityGroups {
		if len(v) == t {
			i := icons["low"]

			if s == severity.Medium {
				i = icons["medium"]
			}

			l := fmt.Sprintf("%s severity", s)
			if s == severity.High {
				l = "Critical severity"
				i = icons["high"]
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

func RenderCodeGroup(w io.Writer, code string, severitiesMap map[severity.Severity][]models.Verdict, icons map[string]string) error {
	tmplData, err := tmpCodeGroup.ReadFile("code_group.html")
	if err != nil {
		return err
	}

	tmpl, err := template.New("code_group").Funcs(template.FuncMap{
		"pluralize": pluralize,
		"icon": func(key string) string {
			return icons[key]
		},
		"severityLabel": func(s severity.Severity) string {
			if s == severity.Low {
				return "Low severity"
			}
			if s == severity.Medium {
				return "Medium severity"
			}
			return "Critical severity"
		},
		"extractMessage": func(grouped TransitivesAndNon) string {
			if len(grouped.Transitive) > 0 {
				return grouped.Transitive[0].Message
			}
			return grouped.NonTransitive[0].Message
		},
		"transitiveAndNonAmount": func(grouped TransitivesAndNon) int {
			return len(grouped.NonTransitive) + len(grouped.Transitive)
		},
	}).Parse(string(tmplData))
	if err != nil {
		return err
	}

	return tmpl.Execute(w, struct {
		Code                string
		CodeData            codeData
		Icons               map[string]string
		CumulatedSeverities cumulatedSeverities
	}{
		Code:                code,
		CodeData:            codeDataMap[code],
		Icons:               icons,
		CumulatedSeverities: newCumulatedSeverities(severitiesMap, icons),
	})
}
