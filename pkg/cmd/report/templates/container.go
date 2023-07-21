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
	"strings"
	"text/template"

	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/pkg/models"
	"github.com/listendev/pkg/models/severity"
	"github.com/listendev/pkg/verdictcode"
)

//go:embed container.html
var tmplContainer embed.FS

//go:embed severity.html
var tmpSeverity embed.FS

//go:embed codegroup.html
var tmpCodegroup embed.FS

//go:embed package.html
var tmpPackage embed.FS

//go:embed code.html
var tmpCode embed.FS

// TODO No signs of suspicious behavior - should we first ensure there are no problems?!

type amounts struct {
	Map   map[string]uint
	Total uint
}

func newAmounts(packages []listen.Package) amounts {
	m := make(map[string]uint)
	var t uint
	for _, p := range packages {
		for _, v := range p.Verdicts {
			m[v.Severity.String()]++
			t++
		}
	}
	return amounts{m, t}
}

// severity -> codeGroup -> name/version -> code -> verdicts
type nestedSeverityCodeGroupCode map[severity.Severity]map[string]map[string]map[verdictcode.Code][]models.Verdict

func nestSeverityCodeGroupCode(packages []listen.Package) nestedSeverityCodeGroupCode {
	m := make(nestedSeverityCodeGroupCode)

	for _, pkg := range packages {
		for _, v := range pkg.Verdicts {
			codeGroups, e := m[v.Severity]
			if !e {
				codeGroups = make(map[string]map[string]map[verdictcode.Code][]models.Verdict)
				m[v.Severity] = codeGroups
			}

			var foundCodeGroup string
			for codeGroup := range codeDataLabel {
				if strings.HasPrefix(v.Code.String(), codeGroup) {
					foundCodeGroup = codeGroup
					break
				}
			}
			if foundCodeGroup == "" {
				// codeGroup not found.
				continue
			}

			nameVersions, e := codeGroups[foundCodeGroup]
			if !e {
				nameVersions = make(map[string]map[verdictcode.Code][]models.Verdict)
				codeGroups[foundCodeGroup] = nameVersions
			}

			nameVersion := fmt.Sprintf("%s/%s", v.Pkg, v.Version)
			codes, e := nameVersions[nameVersion]
			if !e {
				codes = make(map[verdictcode.Code][]models.Verdict)
				nameVersions[nameVersion] = codes
			}

			verdicts, e := codes[v.Code]
			if !e {
				verdicts = []models.Verdict{}
			}

			verdicts = append(verdicts, v)
			codes[v.Code] = verdicts
		}
	}

	return m
}

var icons = map[string]string{
	"high":    "🚨",
	"medium":  "⚠️",
	"low":     "🔷",
	"package": "📦",
	"FNI":     "📡",
	"TSN":     "🔀",
	"MDN":     "📑",
	"STN":     "🔎",
	"DDN":     "🛡️",
}

var codeDataLabel = map[string]string{
	// Ignore UNK
	"FNI": "Dynamic instrumentation",
	"TSN": "Typosquatting",
	"MDN": "Metadata",
	"STN": "Static analysis",
	"DDN": "Advisories",
}

type nameVersion struct {
	Name    string
	Version string
}

var funcs = template.FuncMap{
	"pluralize": func(count int, singular, plural string) string {
		if count == 1 {
			return singular
		}
		return plural
	},
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
	"codeGroupLabel": func(codeGroup string) string {
		return codeDataLabel[codeGroup]
	},
	"getCodeMessage": func(verdict []models.Verdict) string {
		return verdict[0].Message
	},
	"getNameVersion": func(nameSlashVersion string) nameVersion {
		i := nameSlashVersion

		li := strings.LastIndex(i, "/")
		return nameVersion{i[:li], i[li+1:]}
	},
}

func RenderContainer(
	w io.Writer,
	packages []listen.Package,
) error {
	r := NewFromPackages(packages, icons, funcs)

	rHigh, err := r.Severity(severity.High)
	if err != nil {
		return err
	}
	rMedium, err := r.Severity(severity.Medium)
	if err != nil {
		return err
	}
	rLow, err := r.Severity(severity.Low)
	if err != nil {
		return err
	}

	tmplData, err := tmplContainer.ReadFile("container.html")
	if err != nil {
		return err
	}

	tmpl, err := template.New("container").Parse(string(tmplData))
	if err != nil {
		return err
	}

	return tmpl.Execute(w, struct {
		Icons   map[string]string
		High    string
		Medium  string
		Low     string
		Amounts amounts
		Debug   interface{}
	}{
		Icons:   icons,
		High:    rHigh,
		Medium:  rMedium,
		Low:     rLow,
		Amounts: newAmounts(packages),
		Debug:   packages,
	})
}
