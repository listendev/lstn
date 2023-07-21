package templates

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/pkg/models"
	"github.com/listendev/pkg/models/severity"
	"github.com/listendev/pkg/verdictcode"
)

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

	codeGroupsIcons := []string{}
	rCodeGroups := []string{}
	for codeGroup, nameVersions := range codeGroups {
		rCodeGroup, err := r.CodeGroup(codeGroup, nameVersions)
		if err != nil {
			return "", err
		}

		rCodeGroups = append(rCodeGroups, rCodeGroup)
		codeGroupsIcons = append(codeGroupsIcons, r.icons[codeGroup])
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

	occurrences := 0
	rCodes := []string{}
	for code, verdicts := range codes {
		rCode, err := r.Code(code, verdicts)
		if err != nil {
			return "", err
		}
		rCodes = append(rCodes, rCode)
		occurrences += len(verdicts)
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
		var name string
		var version string

		if mn, ok := v.Metadata["npm_package_name"]; !ok {
			return "", fmt.Errorf("'npm_package_name' of %s %s is not of type string", v.Pkg, v.Version)
		} else {
			name = mn.(string)
		}

		if mv, ok := v.Metadata["npm_package_version"]; !ok {
			return "", fmt.Errorf("'npm_package_version' of %s %s is not of type string", v.Pkg, v.Version)
		} else {
			version = mv.(string)
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
