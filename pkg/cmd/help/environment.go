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
	"strings"

	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func envHelpTopicFunc() TopicFunc {
	return func(c *cobra.Command, _ []string) {
		b := bytes.NewBufferString("# lstn environment variables\n\n")
		fmt.Fprintf(b, "%s\n\n", "The environment variables override any corresponding configuration setting.")
		fmt.Fprintf(b, "%s\n\n", "But flags override them.")

		// NOTE > Assuming c.Parent() is the root one
		p := c.Parent()
		if p.HasFlags() {
			configFlagsNames := flags.GetNames(&flags.ConfigFlags{})
			p.Flags().VisitAll(func(f *pflag.Flag) {
				flagName := f.Name
				_, ok := configFlagsNames[flagName]
				if ok {
					envVarName := strings.ToUpper(fmt.Sprintf("%s%s%s", flags.EnvPrefix, flags.EnvSeparator, flags.EnvReplacer.Replace(flagName)))
					fmt.Fprintf(b, "`%s`: %s\n\n", envVarName, f.Usage)
				}
			})
		}

		c.Printf("%s", b.String())
	}
}
