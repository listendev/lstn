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
	"io"
	"text/template"

	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/pkg/models"
	"github.com/listendev/pkg/verdictcode"
)

//go:embed container.html
var tmplContainer embed.FS

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
	Icons          containerDataIcons
	LowSeverity    severityData
	MediumSeverity severityData
	HighSeverity   severityData
	Problems       problemsData
}

type containerDataIcons struct {
	HighSeverity, MediumSeverity, LowSeverity string
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

func RenderContainer(
	w io.Writer,
	packages []listen.Package,
) error {
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
		Icons: containerDataIcons{
			HighSeverity:   "🚨",
			MediumSeverity: "⚠️",
			LowSeverity:    "🔷",
		},
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
