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
package flagusages

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

// FlagGroupAnnotation is the annotations key that marks a flag as belonging to a specific group.
var FlagGroupAnnotation = "lstn___group"

// Set generates the flag usages of the flags local to the input command.
//
// It also groups the flags by the FlagGroupAnnotation annotation.
func Set(c *cobra.Command) {
	lKey := "<local>"
	groups := map[string]*pflag.FlagSet{
		"<origin>": c.LocalFlags(),
	}
	delete(groups, "<origin>")

	addToLocal := func(f *pflag.Flag) {
		if groups[lKey] == nil {
			groups[lKey] = pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
		}
		groups[lKey].AddFlag(f)
	}

	c.LocalNonPersistentFlags().VisitAll(func(f *pflag.Flag) {
		if len(f.Annotations) == 0 {
			addToLocal(f)
		} else {
			if annotations, ok := f.Annotations[FlagGroupAnnotation]; ok {
				g := annotations[0]
				if groups[g] == nil {
					groups[g] = pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
				}
				groups[g].AddFlag(f)
			} else {
				addToLocal(f)
			}
		}
	})

	usages := ""
	if lFlags, ok := groups[lKey]; ok {
		usages += "Flags:\n"
		usages += lFlags.FlagUsages() + "\n"
		delete(groups, lKey)
	}

	for group, flags := range groups {
		usages += fmt.Sprintf("%s Flags:\n", group)
		usages += flags.FlagUsages() + "\n"
	}

	c.SetUsageTemplate(fmt.Sprintf(usageTemplate, strings.TrimSpace(usages)))
}
