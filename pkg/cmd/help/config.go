/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package help

import (
	"bytes"
	"fmt"

	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func configHelpTopicFunc() HelpTopicFunc {
	return func(c *cobra.Command, args []string) {
		b := bytes.NewBufferString("# lstn configuration file\n\n")
		fmt.Fprintf(b, "%s\n\n", "The `lstn` CLI looks for a configuration file .lstn.yaml in your `$HOME` directory when it starts.")
		fmt.Fprintf(b, "%s\n", "In this file you can set the values for the global `lstn` configurations.")
		fmt.Fprintf(b, "%s\n\n", "Anyways, notice that environment variables, and flags (if any) override the values in your configuration file.")
		fmt.Fprintf(b, "%s\n\n", "Here's an example of a configuration file (with the default values):")

		// NOTE > Assuming c.Parent() is the root one
		p := c.Parent()
		if p.HasPersistentFlags() {
			configFlagsNames := flags.GetConfigFlagsNames()
			configFlagsDefaults := flags.GetConfigFlagsDefaults()
			fileContent := ""

			p.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				flagName := f.Name
				_, ok := configFlagsNames[flagName]
				if ok {
					fileContent += fmt.Sprintf("%s: %s\n", flagName, configFlagsDefaults[flagName])
				}
			})

			if fileContent != "" {
				fmt.Fprintf(b, "```yaml\n%s```\n", fileContent)
			}
		}

		c.Printf("%s", b.String())
	}
}
