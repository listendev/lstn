package packagesprinter

import (
	"fmt"
	"io"
	"strconv"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/utils"
	"github.com/listendev/lstn/pkg/listen"
)

type VerdictPriority string

const (
	VerdictPriorityLow    VerdictPriority = "low"
	VerdictPriorityMedium VerdictPriority = "medium"
	VerdictPriorityHigh   VerdictPriority = "high"
)

func verdictPriorityToColorFunc(colorScheme *iostreams.ColorScheme, p VerdictPriority) func(string) string {
	var fn func(string) string
	switch p {
	case VerdictPriorityHigh:
		fn = colorScheme.Red
	case VerdictPriorityMedium:
		fn = colorScheme.Yellow
	case VerdictPriorityLow:
		fn = colorScheme.Cyan
	default:
		fn = func(s string) string {
			return s
		}
	}

	return fn
}

func printVerdictMetadata(out io.Writer, cs *iostreams.ColorScheme, metadata map[string]interface{}) {
	for mdkey, md := range metadata {
		if mdkey == "npm_package_name" || mdkey == "npm_package_version" {
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
		fmt.Fprintf(out, "    %s: %s\n", mdkey, cs.Gray(mdStr))
	}
}

func printVerdict(out io.Writer, cs *iostreams.ColorScheme, p listen.Package, verdict listen.Verdict) {
	prioColor := verdictPriorityToColorFunc(cs, VerdictPriority(verdict.Priority))
	fmt.Fprintf(out, "  %s %s", prioColor(fmt.Sprintf("[%s]", verdict.Priority)), verdict.Message)
	metadataPackageName := ""
	metadataPackageVersion := ""
	if packageName, ok := verdict.Metadata["npm_package_name"]; ok {
		metadataPackageName = packageName.(string)
	}
	if packageVersion, ok := verdict.Metadata["npm_package_version"]; ok {
		metadataPackageVersion = packageVersion.(string)
	}

	if metadataPackageName != p.Name && metadataPackageVersion != p.Version {
		fmt.Fprintf(out, cs.Bold(" (from transitive dependency %s@%s)"), cs.CyanBold(metadataPackageName), cs.CyanBold(metadataPackageVersion))
	}
	fmt.Fprintln(out, "")
	printVerdictMetadata(out, cs, verdict.Metadata)
}

func printProblem(out io.Writer, cs *iostreams.ColorScheme, p listen.Package, problem listen.Problem) {
	fmt.Fprintf(out, "  %s: %s", cs.Yellow(fmt.Sprintf("- %s", problem.Title)), cs.Gray(problem.Type))
	fmt.Fprintln(out, "")
}

func printPackages(out io.Writer, cs *iostreams.ColorScheme, packages *listen.Response) {
	for _, p := range *packages {
		if len(p.Verdicts) == 0 && len(p.Problems) == 0 {
			continue
		}
		fmt.Fprintln(out, "")
		thereIsAre := "are"
		verdictsWord := "verdicts"
		if len(p.Verdicts) == 1 {
			verdictsWord = "verdict"
			thereIsAre = "is"
		}
		problemsWord := "problems"
		if len(p.Problems) == 1 {
			problemsWord = "problem"
		}
		fmt.Fprintf(out, "There %s %s %s and %s %s for %s@%s\n", thereIsAre, cs.Bold(strconv.Itoa(len(p.Verdicts))), verdictsWord, cs.Bold(strconv.Itoa(len(p.Problems))), problemsWord, cs.CyanBold(p.Name), cs.CyanBold(p.Version))
		fmt.Fprintln(out, "")
		for _, verdict := range p.Verdicts {
			printVerdict(out, cs, p, verdict)
		}

		for _, problem := range p.Problems {
			printProblem(out, cs, p, problem)
		}
		fmt.Fprintln(out, "")
	}

}

func PrintTable(io *iostreams.IOStreams, packages *listen.Response) error {
	tab := utils.NewTablePrinter(io)

	cs := io.ColorScheme()

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
		tab.AddField(p.Version, nil, nil)

		tab.AddField(fmt.Sprintf("%s %d verdicts", verdictsIcon, len(p.Verdicts)), nil, verdictsColor)
		tab.AddField(fmt.Sprintf("%s %d problems", problemsIcon, len(p.Problems)), nil, problemsColor)

		tab.EndRow()
	}

	err := tab.Render()
	if err != nil {
		return err
	}

	printPackages(io.Out, cs, packages)

	return nil
}
