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

package packagesprinter

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/utils"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/pkg/models"
	"github.com/listendev/pkg/verdictcode"
)

const (
	verdictSeverityLow    string = "low"
	verdictSeverityMedium string = "medium"
	verdictSeverityHigh   string = "high"
)

func verdictSeverityToColorFunc(colorScheme *iostreams.ColorScheme, sev string) func(string) string {
	var fn func(string) string
	switch sev {
	case verdictSeverityHigh:
		fn = colorScheme.Red
	case verdictSeverityMedium:
		fn = colorScheme.Yellow
	case verdictSeverityLow:
		fn = colorScheme.Cyan
	default:
		fn = func(s string) string {
			return s
		}
	}

	return fn
}

type TablePrinter struct {
	streams *iostreams.IOStreams
}

func NewTablePrinter(streams *iostreams.IOStreams) *TablePrinter {
	return &TablePrinter{
		streams: streams,
	}
}

func (t *TablePrinter) RenderPackages(pkgs *listen.Response) error {
	err := t.printTable(pkgs)
	if err != nil {
		return err
	}
	t.printPackages(pkgs)

	return nil
}

func (t *TablePrinter) printVerdictMetadata(metadata map[string]interface{}) {
	keys := make([]string, 0, len(metadata))
	for k := range metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	cs := t.streams.ColorScheme()
	for _, mdkey := range keys {
		md := metadata[mdkey]
		if mdkey == "npm_package_name" || mdkey == "npm_package_version" || mdkey == "file_content" || mdkey == "lines" {
			continue
		}
		if md == nil {
			continue
		}
		mdStr := ""
		if md, ok := md.(string); ok {
			mdStr = md
		}
		if md, ok := md.(int); ok {
			mdStr = strconv.Itoa(md)
		}
		if len(mdStr) == 0 {
			continue
		}
		fmt.Fprintf(t.streams.Out, "    %s: %s\n", mdkey, cs.Gray(mdStr))
	}
}

func (t *TablePrinter) printVerdict(p *listen.Package, verdict listen.Verdict) {
	cs := t.streams.ColorScheme()
	out := t.streams.Out
	prioColor := verdictSeverityToColorFunc(cs, verdict.Severity.String())
	fmt.Fprintf(out, "  %s %s", prioColor(fmt.Sprintf("[%s]", verdict.Severity)), verdict.Message)
	metadataPackageName := ""
	metadataPackageVersion := ""
	if packageName, ok := verdict.Metadata["npm_package_name"]; ok {
		metadataPackageName = packageName.(string)
	}
	if packageVersion, ok := verdict.Metadata["npm_package_version"]; ok {
		metadataPackageVersion = packageVersion.(string)
	}

	if metadataPackageName != "" &&
		metadataPackageVersion != "" &&
		metadataPackageName != p.Name &&
		p.Version != nil &&
		metadataPackageVersion != *p.Version {
		fmt.Fprintf(out, cs.Bold(" (from transitive dependency %s@%s)"), cs.CyanBold(metadataPackageName), cs.CyanBold(metadataPackageVersion))
	}
	fmt.Fprintln(out, "")
	t.printVerdictMetadata(verdict.Metadata)
}

func (t *TablePrinter) printProblem(problem listen.Problem) {
	cs := t.streams.ColorScheme()
	out := t.streams.Out
	fmt.Fprintf(out, "  %s: %s", cs.Yellow(fmt.Sprintf("- %s", problem.Title)), cs.Gray(problem.Type))
	fmt.Fprintln(out, "")
}

func (t *TablePrinter) printPackage(p *listen.Package) {
	verdicts := []models.Verdict{}
	for _, v := range p.Verdicts {
		if v.Code == verdictcode.UNK {
			continue
		}
		verdicts = append(verdicts, v)
	}

	cs := t.streams.ColorScheme()
	out := t.streams.Out
	thereIsAre := "are"
	verdictsWord := "verdicts"
	if len(verdicts) == 1 {
		verdictsWord = "verdict"
		thereIsAre = "is"
	}
	problemsWord := "problems"
	if len(p.Problems) == 1 {
		problemsWord = "problem"
	}
	versionStr := ""
	if p.Version != nil {
		versionStr = fmt.Sprintf("@%s", cs.CyanBold(*p.Version))
	}

	fmt.Fprintf(out, "There %s %s %s and %s %s for %s%s\n", thereIsAre, cs.Bold(strconv.Itoa(len(verdicts))), verdictsWord, cs.Bold(strconv.Itoa(len(p.Problems))), problemsWord, cs.CyanBold(p.Name), versionStr)
	fmt.Fprintln(out, "")
	for _, verdict := range verdicts {
		t.printVerdict(p, verdict)
	}

	for _, problem := range p.Problems {
		t.printProblem(problem)
	}
	fmt.Fprintln(out, "")
}

func (t *TablePrinter) printPackages(packages *listen.Response) {
	out := t.streams.Out
	for _, p := range *packages {
		if len(p.Verdicts) == 0 && len(p.Problems) == 0 {
			continue
		}
		fmt.Fprintln(out, "")
		t.printPackage(&p)
	}
}

func (t *TablePrinter) printTable(packages *listen.Response) error {
	tab := utils.NewTablePrinter(t.streams)

	cs := t.streams.ColorScheme()

	for _, p := range *packages {
		verdictsColor := cs.ColorFromString("green")
		verdictsIcon := cs.SuccessIcon()
		if len(p.Verdicts) > 0 {
			verdictsColor = cs.ColorFromString("red")
			verdictsIcon = cs.FailureIcon()
		}

		problemsColor := cs.ColorFromString("green")
		problemsIcon := cs.SuccessIcon()
		if len(p.Problems) > 0 {
			problemsColor = cs.ColorFromString("yellow")
			problemsIcon = cs.WarningIcon()
		}

		tab.AddField(p.Name, nil, cs.Bold)

		version := ""
		if p.Version != nil {
			version = *p.Version
		}
		tab.AddField(version, nil, nil)

		tab.AddField(fmt.Sprintf("%s %d verdicts", verdictsIcon, len(p.Verdicts)), nil, verdictsColor)
		tab.AddField(fmt.Sprintf("%s %d problems", problemsIcon, len(p.Problems)), nil, problemsColor)

		tab.EndRow()
	}

	return tab.Render()
}
