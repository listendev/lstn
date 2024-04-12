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
package help

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func configHelpTopicFunc() TopicFunc {
	return func(c *cobra.Command, _ []string) {
		b := bytes.NewBufferString("# lstn configuration file\n\n")
		fmt.Fprintf(b, "%s\n\n", "The `lstn` CLI looks for a configuration file `.lstn.yaml` in your `$HOME` or into the current working directory from which `lstn` is getting called.")
		fmt.Fprintf(b, "%s\n\n", "When invoking `lstn in <dir>` it also looks for `.lstn.yaml` into `<dir>`.")

		fmt.Fprintf(b, "%s\n", "In this file you can set the values for the global `lstn` configurations.")
		fmt.Fprintf(b, "%s\n\n", "Anyways, notice that environment variables, and flags (if any) override the values in your configuration file.")
		fmt.Fprintf(b, "%s\n\n", "Here's an example of a configuration file (with the default values):")

		// NOTE > Assuming c.Parent() is the root one
		p := c.Root()
		if p.HasFlags() {
			cfgFlags := &flags.ConfigFlags{}
			configFlagsNames := flags.GetNames(cfgFlags)
			configFlagsDefaults := flags.GetDefaults(cfgFlags)
			configFlagsTypes := flags.GetTypes(cfgFlags)

			fileContent := ""

			dots := []string{}
			p.Flags().VisitAll(func(f *pflag.Flag) {
				flagName := f.Name
				target, ok := configFlagsNames[flagName]
				if ok {
					dots = append(dots, fmt.Sprintf("%s:%s", strings.ToLower(target), flagName))
				}
			})
			sort.Strings(dots)

			keys := map[string]bool{}
			for _, dot := range dots {
				components := strings.Split(dot, ":")
				flagName := components[1]
				parts := strings.Split(components[0], ".")
				num := len(parts)
				for i, part := range parts {
					path := part
					if i > 0 {
						path = strings.Join(parts[:i], ".") + "." + part
					}
					if _, found := keys[path]; found {
						continue
					}

					def := configFlagsDefaults[flagName]
					if def == "" && num == (i+1) {
						def = "..."
					}
					if def != "" && num > (i+1) {
						def = ""
					}

					switch configFlagsTypes[flagName].String() {
					case "string":
						if def != "" {
							def = fmt.Sprintf("%q", def)
						}
					case "int":
						if def == "..." {
							def = "0"
						}
					}

					if strings.HasPrefix(configFlagsTypes[flagName].String(), "[]") {
						def = ""
					}

					fileContent += fmt.Sprintf("%s%s: %s\n", strings.Repeat(" ", i*2), part, def)
					if strings.HasPrefix(configFlagsTypes[flagName].String(), "[]") && num == (i+1) {
						sp := strings.Repeat(" ", (i+1)*2)
						fileContent += fmt.Sprintf("%s- %q\n%s- %q\n", sp, "...", sp, "...")
					}

					keys[path] = true
				}
			}

			if fileContent != "" {
				fmt.Fprintf(b, "```yaml\n%s```\n", fileContent)
			}
		}

		c.Printf("%s", b.String())
	}
}
