package help

import (
	"fmt"
	"os"

	"github.com/listendev/lstn/pkg/text"
	"github.com/spf13/cobra"
)

var Topics = map[string]map[string]string{
	"config": {
		"short": "Details about the .lstn.yml file",
		"long":  "TODO config",
	},
	"environment": {
		"short": "Which environment variables you can use with lstn",
		"long":  "TODO environment",
	},
	"exit": {
		"short": "How lstn exits",
		"long":  "TODO exit",
	},
	"manual": {
		"short": "A comprehensive reference of all the lstn commands",
	},
}

type HelpTopicFunc func(*cobra.Command, []string)

var topicsHelpFuncs = map[string]func() HelpTopicFunc{
	"manual": manualHelpTopicFunc,
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
