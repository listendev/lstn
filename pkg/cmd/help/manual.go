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
	"io"
	"os"
	"strings"

	"github.com/listendev/lstn/pkg/text"
	"github.com/spf13/cobra"
)

func cReference(w io.Writer, c *cobra.Command, depth int) {
	// Name
	fmt.Fprintf(w, "%s `%s`\n\n", strings.Repeat("#", depth), c.UseLine())

	// Short description
	descr := c.Short
	if !strings.HasSuffix(descr, ".") {
		descr += "."
	}
	fmt.Fprintf(w, "%s\n\n", descr)

	// Local flags
	if c.HasAvailableLocalFlags() {
		localFlagsUsage := c.LocalFlags().FlagUsages()
		if localFlagsUsage != "" {
			fmt.Fprintf(w, "```\n%s```\n\n", text.Dedent(localFlagsUsage))
		}
	}

	// Examples
	if c.HasExample() {
		fmt.Fprintf(w, "For example:\n\n```bash\n%s\n```\n\n", text.Dedent(c.Example))
	}

	// Subcommands
	for _, c := range c.Commands() {
		if c.Hidden {
			continue
		}
		cReference(w, c, depth+1)
	}
}

func manualHelpTopicFunc() HelpTopicFunc {
	return func(c *cobra.Command, args []string) {
		b := bytes.NewBufferString("# lstn cheatsheet\n\n")

		// NOTE > Assuming c.Parent() is the root one
		p := c.Parent()
		if p.HasPersistentFlags() {
			fmt.Fprintf(b, "## Global Flags\n\n")
			fmt.Fprintf(b, "Every child command inherits the following flags:\n\n")
			persFlagsUsage := p.PersistentFlags().FlagUsages()
			if persFlagsUsage != "" {
				fmt.Fprintf(b, "```\n%s```\n\n", text.Dedent(persFlagsUsage))
			}
		}

		for _, c := range p.Commands() {
			if c.Hidden {
				continue
			}
			cReference(b, c, 2)
		}

		fmt.Fprintf(os.Stdout, "%s", b.String())
	}
}
