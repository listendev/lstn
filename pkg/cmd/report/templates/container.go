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
	"io"
	"strings"
	"text/template"

	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/pkg/models"
	"github.com/listendev/pkg/models/severity"
)

// TODO No signs of suspicious behavior - should we first ensure there are no problems?!

type amounts struct {
	Map      map[string]uint
	Total    uint
	Problems uint
}

func newAmounts(packages []listen.Package) amounts {
	m := make(map[string]uint)
	var p uint
	var t uint
	for _, pkg := range packages {
		if len(pkg.Problems) > 0 {
			p++
			continue
		}

		for _, v := range pkg.Verdicts {
			m[v.Severity.String()]++
			t++
		}
	}
	return amounts{m, t, p}
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

	rProblems, err := r.Problems()
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
		Icons          map[string]string
		Amounts        amounts
		RenderHigh     string
		RenderMedium   string
		RenderLow      string
		RenderProblems string
	}{
		Icons:          icons,
		Amounts:        newAmounts(packages),
		RenderHigh:     rHigh,
		RenderMedium:   rMedium,
		RenderLow:      rLow,
		RenderProblems: rProblems,
	})
}
