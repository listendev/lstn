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
	"bytes"
	"embed"
	"fmt"
	"sort"
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

//go:embed problems.html
var tmpProblems embed.FS

type render struct {
	packages []listen.Package
	data     nestedSeverityCodeGroupCode
	icons    map[string]string
	funcs    template.FuncMap
}

// Mapping severity -> codeGroup -> name/version -> code -> verdicts.
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

			nameVersion := strings.TrimPrefix(fmt.Sprintf("%s/%s@%s", v.Org, v.Pkg, v.Version), "/")
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

//nolint:revive // this is a constructor returning a pointer to a struct not meant for external usage at the moment
func NewFromPackages(packages []listen.Package, icons map[string]string, funcs template.FuncMap) *render {
	data := nestSeverityCodeGroupCode(packages)

	return &render{packages, data, icons, funcs}
}

func (r *render) Severity(s severity.Severity) (string, error) {
	var render bytes.Buffer

	tmplData, err := tmpSeverity.ReadFile("severity.html")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("severity").Funcs(r.funcs).Parse(string(tmplData))
	if err != nil {
		return "", err
	}

	codeGroups, e := r.data[s]
	if !e {
		return "", nil
	}

	type holder struct {
		codeGroup    string
		nameVersions map[string]map[verdictcode.Code][]models.Verdict
	}
	cg := make([]holder, 0, len(codeGroups))
	for codeGroup, nameVersions := range codeGroups {
		cg = append(cg, holder{codeGroup, nameVersions})
	}
	sort.SliceStable(cg, func(i, j int) bool {
		return cg[i].codeGroup < cg[j].codeGroup
	})

	codeGroupsIcons := []string{}
	rCodeGroups := []string{}
	for _, element := range cg {
		rCodeGroup, err := r.CodeGroup(element.codeGroup, element.nameVersions)
		if err != nil {
			return "", err
		}

		rCodeGroups = append(rCodeGroups, rCodeGroup)
		codeGroupsIcons = append(codeGroupsIcons, r.icons[element.codeGroup])
	}

	if err := tmpl.Execute(&render, struct {
		Icons            map[string]string
		Severity         severity.Severity
		CodeGroupIcons   []string
		RenderCodeGroups []string
	}{
		Icons:            r.icons,
		Severity:         s,
		CodeGroupIcons:   codeGroupsIcons,
		RenderCodeGroups: rCodeGroups,
	}); err != nil {
		return "", err
	}

	return render.String(), nil
}

func (r *render) CodeGroup(codeGroup string, nameVersions map[string]map[verdictcode.Code][]models.Verdict) (string, error) {
	var render bytes.Buffer

	tmplData, err := tmpCodegroup.ReadFile("codegroup.html")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("codegroup").Funcs(r.funcs).Parse(string(tmplData))
	if err != nil {
		return "", err
	}

	type holder struct {
		nameVersion string
		codes       map[verdictcode.Code][]models.Verdict
	}
	n := make([]holder, 0, len(nameVersions))
	for nameVersion, codes := range nameVersions {
		n = append(n, holder{nameVersion, codes})
	}
	sort.SliceStable(n, func(i, j int) bool {
		return n[i].nameVersion > n[j].nameVersion
	})

	rNameVersions := []string{}
	for _, element := range n {
		rNameVersion, err := r.Package(element.nameVersion, element.codes)
		if err != nil {
			return "", err
		}
		rNameVersions = append(rNameVersions, rNameVersion)
	}

	if err := tmpl.Execute(&render, struct {
		Icons          map[string]string
		CodeGroup      string
		RenderPackages []string
	}{
		Icons:          r.icons,
		CodeGroup:      codeGroup,
		RenderPackages: rNameVersions,
	}); err != nil {
		return "", err
	}

	return render.String(), nil
}

func (r *render) Package(nameVersion string, codes map[verdictcode.Code][]models.Verdict) (string, error) {
	var render bytes.Buffer

	tmplData, err := tmpPackage.ReadFile("package.html")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("package").Funcs(r.funcs).Parse(string(tmplData))
	if err != nil {
		return "", err
	}

	li := strings.LastIndex(nameVersion, "@")

	occurrences := 0
	type holder struct {
		code     verdictcode.Code
		verdicts []models.Verdict
	}
	c := make([]holder, 0, len(codes))
	for code, verdicts := range codes {
		c = append(c, holder{code, verdicts})
		occurrences += len(verdicts)
	}
	sort.SliceStable(c, func(i, j int) bool {
		return c[i].code < c[j].code
	})

	rCodes := []string{}
	for _, element := range c {
		rCode, err := r.Code(element.code, element.verdicts)
		if err != nil {
			return "", err
		}
		rCodes = append(rCodes, rCode)
	}

	if err := tmpl.Execute(&render, struct {
		Icons       map[string]string
		Name        string
		Version     string
		RenderCodes []string
		Occurrences int
	}{
		Icons:       r.icons,
		Name:        nameVersion[:li],
		Version:     nameVersion[li+1:],
		RenderCodes: rCodes,
		Occurrences: occurrences,
	}); err != nil {
		return "", err
	}

	return render.String(), nil
}

func (r *render) Code(code verdictcode.Code, verdicts []models.Verdict) (string, error) {
	// The verdicts provided are all guaranteed to have the same code.

	var render bytes.Buffer

	tmplData, err := tmpCode.ReadFile("code.html")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("code").Funcs(r.funcs).Parse(string(tmplData))
	if err != nil {
		return "", err
	}

	type grouped struct {
		Transitive bool
		Refs       []models.Verdict
	}
	cumulated := make(map[string]grouped)

	for _, v := range verdicts {
		nameInVerdict := strings.TrimPrefix(fmt.Sprintf("%s/%s", v.Org, v.Pkg), "/")
		name := nameInVerdict
		version := v.Version
		transitive := false

		mn, ok := v.Metadata["npm_package_name"]
		if ok {
			name = mn.(string)
		}
		transitive = name != nameInVerdict

		mv, ok := v.Metadata["npm_package_version"]
		if ok {
			version = mv.(string)
		}

		key := fmt.Sprintf("%s/%s", name, version)

		g, e := cumulated[key]
		if !e {
			g = grouped{transitive, []models.Verdict{}}
		}
		g.Refs = append(g.Refs, v)
		cumulated[key] = g
	}

	if err := tmpl.Execute(&render, struct {
		Icons             map[string]string
		Code              verdictcode.Code
		Verdicts          []models.Verdict
		CumulatedVerdicts map[string]grouped
	}{
		Icons:             r.icons,
		Code:              code,
		Verdicts:          verdicts,
		CumulatedVerdicts: cumulated,
	}); err != nil {
		return "", err
	}

	return render.String(), nil
}

func (r *render) Problems() (string, error) {
	var render bytes.Buffer

	tmplData, err := tmpProblems.ReadFile("problems.html")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("problems").Funcs(r.funcs).Parse(string(tmplData))
	if err != nil {
		return "", err
	}

	problems := make(map[string][]listen.Package)
	for _, pkg := range r.packages {
		if len(pkg.Problems) == 0 {
			continue
		}

		for _, p := range pkg.Problems {
			a, e := problems[p.Title]
			if !e {
				a = []listen.Package{}
			}

			a = append(a, pkg)
			problems[p.Title] = a
		}
	}

	if err := tmpl.Execute(&render, struct {
		Icons    map[string]string
		Problems map[string][]listen.Package
	}{
		Icons:    r.icons,
		Problems: problems,
	}); err != nil {
		return "", err
	}

	return render.String(), nil
}
