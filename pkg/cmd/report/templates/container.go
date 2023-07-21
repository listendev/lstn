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
type severityData struct {
	Packages      []listen.Package
	TotalVerdicts int
	DetailsRender string
}
type problemsData struct {
	Packages      []listen.Package
	TotalProblems int
	DetailsRender string
}

type containerData struct {
	Icons          map[string]string
	GroupedRender  string
	Amounts        amounts
	High           string
	Medium         string
	Low            string
	LowSeverity    severityData
	MediumSeverity severityData
	HighSeverity   severityData
	Problems       problemsData
}

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

func countVerdicts(packages []listen.Package) int {
	var count int
	for _, p := range packages {
		verdicts := []models.Verdict{}
		for _, v := range p.Verdicts {
			if v.Code == verdictcode.UNK {
				continue
			}
			verdicts = append(verdicts, v)
		}
		count += len(verdicts)
	}

	return count
}

func countProblems(packages []listen.Package) int {
	var count int
	for _, p := range packages {
		count += len(p.Problems)
	}

	return count
}

func filterPackagesByVerdictSeverity(packages []listen.Package, sev string) []listen.Package {
	filteredPackages := []listen.Package{}
	for _, p := range packages {
		if len(p.Verdicts) == 0 {
			continue
		}
		currentPackage := p
		currentPackage.Verdicts = []listen.Verdict{}
		for _, v := range p.Verdicts {
			if v.Severity.String() == sev {
				currentPackage.Verdicts = append(currentPackage.Verdicts, v)

				break
			}
		}
		if len(currentPackage.Verdicts) > 0 {
			filteredPackages = append(filteredPackages, currentPackage)
		}
	}

	return filteredPackages
}

func renderGrouped(codesMap groupedByCodesSeverity, icons map[string]string) (string, error) {
	var render bytes.Buffer

	for code, severitiesMap := range codesMap {
		if len(severitiesMap) == 0 {
			continue
		}

		// TODO make renderer a struct and add methods on it. This way you can set icons (and any other thing) in the struct and make it available in all renders.
		err := RenderCodeGroup(&render, code, severitiesMap, icons)
		if err != nil {
			return "", err
		}
	}

	return render.String(), nil
}

func renderDetails(packages []listen.Package) (string, error) {
	var detailsRender bytes.Buffer
	for _, p := range packages {
		if len(p.Verdicts) == 0 {
			continue
		}
		err := RenderSingleVerdictsPackage(&detailsRender, p)
		if err != nil {
			return "", err
		}
	}

	return detailsRender.String(), nil
}

func renderProblems(packages []listen.Package) (string, error) {
	var detailsRender bytes.Buffer
	for _, p := range packages {
		if len(p.Problems) == 0 {
			continue
		}
		err := RenderSingleProblemsPackage(&detailsRender, p)
		if err != nil {
			return "", err
		}
	}

	return detailsRender.String(), nil
}

type groupedByCodesSeverity map[string]map[severity.Severity][]models.Verdict

func nestGroupCodeSeverity(packages []listen.Package) groupedByCodesSeverity {
	// Ignore UNK
	codesMap := groupedByCodesSeverity{
		"DDN": make(map[severity.Severity][]models.Verdict),
		"FNI": make(map[severity.Severity][]models.Verdict),
		"MDN": make(map[severity.Severity][]models.Verdict),
		"STN": make(map[severity.Severity][]models.Verdict),
		"TSN": make(map[severity.Severity][]models.Verdict),
	}

	for _, pkg := range packages {
		for _, verdict := range pkg.Verdicts {

		code:
			for codePrefix, severitiesMap := range codesMap {
				if strings.HasPrefix(verdict.Code.String(), codePrefix) {
					severitiesMap[verdict.Severity] = append(severitiesMap[verdict.Severity], verdict)

					break code
				}
			}

		}
	}

	return codesMap
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
			for codeGroup := range codeDataMap {
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

type render struct {
	data  nestedSeverityCodeGroupCode
	icons map[string]string
	funcs template.FuncMap
}

func NewFromPackages(packages []listen.Package, icons map[string]string, funcs template.FuncMap) *render {
	data := nestSeverityCodeGroupCode(packages)
	return &render{data, icons, funcs}
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

	rCodeGroups := []string{}
	for codeGroup, nameVersions := range codeGroups {
		rCodeGroup, err := r.CodeGroup(codeGroup, nameVersions)
		if err != nil {
			return "", err
		}

		rCodeGroups = append(rCodeGroups, rCodeGroup)
	}

	if err := tmpl.Execute(&render, struct {
		Severity         severity.Severity
		Icons            map[string]string
		RenderCodeGroups []string
	}{
		Icons:            r.icons,
		Severity:         s,
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

	rNameVersions := []string{}
	for nameVersion, codes := range nameVersions {
		rNameVersion, err := r.Package(nameVersion, codes)
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

	li := strings.LastIndex(nameVersion, "/")

	rCodes := []string{}
	for code, verdicts := range codes {
		rCode, err := r.Code(code, verdicts)
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
	}{
		Icons:       r.icons,
		Name:        nameVersion[:li],
		Version:     nameVersion[li+1:],
		RenderCodes: rCodes,
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
		var ok bool
		var name string
		var version string

		name, ok = v.Metadata["npm_package_name"].(string)
		if !ok {
			return "", fmt.Errorf("'npm_package_name' of %s %s is not of type string", v.Pkg, v.Version)
		}
		version, ok = v.Metadata["npm_package_version"].(string)
		if !ok {
			return "", fmt.Errorf("'npm_package_version' of %s %s is not of type string", v.Pkg, v.Version)
		}

		transitive := name != v.Pkg && version != v.Version

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

var icons = map[string]string{
	"high":    "🚨",
	"medium":  "⚠️",
	"low":     "🔷",
	"package": "📦",
}

type codeData struct {
	Label string
	Icon  string
}

var codeDataMap = map[string]codeData{
	// Ignore UNK
	"FNI": {"Dynamic instrumentation", "📡"},
	"TSN": {"Typosquatting", "🔀"},
	"MDN": {"Metadata", "📑"},
	"STN": {"Static analysis", "🔎"},
	"DDN": {"Advisories", "🛡️"},
}

type nameVersion struct {
	Name    string
	Version string
}

var funcs = template.FuncMap{
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
	"codeGroupData": func(codeGroup string) codeData {
		return codeDataMap[codeGroup]
	},
	"getCodeMessage": func(verdict []models.Verdict) string {
		return verdict[0].Message
	},
	"extractMessage": func(grouped TransitivesAndNon) string {
		if len(grouped.Transitive) > 0 {
			return grouped.Transitive[0].Message
		}
		return grouped.NonTransitive[0].Message
	},
	"getNameVersion": func(nameSlashVersion string) nameVersion {
		i := nameSlashVersion

		li := strings.LastIndex(i, "/")
		return nameVersion{i[:li], i[li+1:]}
	},
	"transitiveAndNonAmount": func(grouped TransitivesAndNon) int {
		return len(grouped.NonTransitive) + len(grouped.Transitive)
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

	nestedGroups := nestGroupCodeSeverity(packages)
	groupedRender, err := renderGrouped(nestedGroups, icons)
	if err != nil {
		return err
	}

	highAlertPackages := filterPackagesByVerdictSeverity(packages, "high")
	mediumAlertPackages := filterPackagesByVerdictSeverity(packages, "medium")
	lowAlertPacakges := filterPackagesByVerdictSeverity(packages, "low")

	lowDetails, err := renderDetails(lowAlertPacakges)
	if err != nil {
		return err
	}
	lowSeverityData := severityData{
		Packages:      lowAlertPacakges,
		TotalVerdicts: countVerdicts(lowAlertPacakges),
		DetailsRender: lowDetails,
	}

	mediumDetails, err := renderDetails(mediumAlertPackages)
	if err != nil {
		return err
	}
	mediumSeverityData := severityData{
		Packages:      mediumAlertPackages,
		TotalVerdicts: countVerdicts(mediumAlertPackages),
		DetailsRender: mediumDetails,
	}

	highDetails, err := renderDetails(highAlertPackages)
	if err != nil {
		return err
	}
	highSeverityData := severityData{
		Packages:      highAlertPackages,
		TotalVerdicts: countVerdicts(highAlertPackages),
		DetailsRender: highDetails,
	}

	problems, err := renderProblems(packages)
	if err != nil {
		return err
	}
	pdata := problemsData{
		Packages:      packages,
		TotalProblems: countProblems(packages),
		DetailsRender: problems,
	}

	cdata := containerData{
		Icons:          icons,
		GroupedRender:  groupedRender,
		High:           rHigh,
		Medium:         rMedium,
		Low:            rLow,
		Amounts:        newAmounts(packages),
		LowSeverity:    lowSeverityData,
		MediumSeverity: mediumSeverityData,
		HighSeverity:   highSeverityData,
		Problems:       pdata,
	}

	tmplData, err := tmplContainer.ReadFile("container.html")
	if err != nil {
		return err
	}

	tmpl, err := template.New("container").Parse(string(tmplData))
	if err != nil {
		return err
	}

	return tmpl.Execute(w, cdata)
}
