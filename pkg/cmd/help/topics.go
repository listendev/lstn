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
	"fmt"
	"os"

	"github.com/listendev/lstn/pkg/text"
	"github.com/spf13/cobra"
)

var Topics = map[string]map[string]string{
	"config": {
		"short": "Details about the .lstn.yml config file",
		"long":  "TODO config",
	},
	"environment": {
		"alias": "env",
		"short": "Which environment variables you can use with lstn",
	},
	"exit": {
		"short": "Details about the lstn exit codes",
		"long":  "TODO exit",
	},
	"manual": {
		"short": "A comprehensive reference of all the lstn commands",
	},
}

type HelpTopicFunc func(*cobra.Command, []string)

var topicsHelpFuncs = map[string]func() HelpTopicFunc{
	"manual":      manualHelpTopicFunc,
	"environment": envHelpTopicFunc,
}

// TODO > print out markdown

func NewTopic(topic string) *cobra.Command {
	c := &cobra.Command{
		Use:                   topic,
		DisableFlagsInUseLine: true,
		Short:                 Topics[topic]["short"],
		Long:                  Topics[topic]["long"],
		Example:               Topics[topic]["example"],
		// TODO > remove these if unused
		Annotations: map[string]string{
			"markdown:generate": "true",
			"markdown:basename": "lstn_help_" + topic,
		},
	}

	if Topics[topic]["alias"] != "" {
		c.Aliases = []string{Topics[topic]["alias"]}
	}

	c.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c.Long != "" {
			fmt.Fprintf(os.Stdout, c.Long)
			if c.Example != "" {
				fmt.Fprintf(os.Stdout, "\n\nExamples:\n")
				fmt.Fprintf(os.Stdout, "%s", text.Indent(c.Example, "  "))
			}
		} else if topicsHelpFuncs[c.Use] != nil {
			topicsHelpFuncs[c.Use]()(c, args)
		}
	})

	c.SetUsageFunc(func(c *cobra.Command) error {
		fmt.Fprintf(os.Stdout, c.Use)
		return nil
	})

	return c
}
