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
package root

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/listendev/lstn/cmd/in"
	"github.com/listendev/lstn/cmd/to"
	"github.com/listendev/lstn/cmd/version"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	"github.com/listendev/lstn/pkg/cmd/groups"
	pkghelp "github.com/listendev/lstn/pkg/cmd/help"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/jq"
	lstnversion "github.com/listendev/lstn/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfgFile           string
	_, filename, _, _ = runtime.Caller(0)
)

type Command struct {
	cmd *cobra.Command
	ctx context.Context
}

//gocyclo:ignore
func New(ctx context.Context) (*Command, error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:               "lstn",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Short:             "Analyze the behavior of your dependencies using listen.dev",
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			// Do not check for the config file if the command is not available (eg., help) or not core (eg., version)
			c, _, err := c.Find(os.Args[1:])
			if err == nil && (c.IsAvailableCommand() && c.GroupID == groups.Core.ID) {
				// If a config file is found, read it in
				if err := viper.ReadInConfig(); err == nil {
					fmt.Fprintln(os.Stderr, "Using config file: ", viper.ConfigFileUsed())
				} else {
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						// Config file not found, ignore...
						fmt.Fprintln(os.Stderr, "Running without a configuration file")
					} else {
						// Config file was found but another error was produced
						fmt.Fprintf(os.Stderr, "Error running with config file: %s\n", viper.ConfigFileUsed())
					}
				}
			}

			// Obtain the configuration options
			cfgOpts, ok := c.Context().Value(pkgcontext.ConfigKey).(*flags.ConfigFlags)
			if !ok {
				return fmt.Errorf("couldn't obtain configuration options")
			}
			// Obtain the mapping flag name -> struct field name
			configFlagsNames := flags.GetNames(cfgOpts)
			// Obtain the mapping flag name -> default value
			configFlagsDefaults := flags.GetDefaults(cfgOpts)
			// Implement flag precedence over environment variables, over configuration file
			c.Flags().VisitAll(func(f *pflag.Flag) {
				flagName := f.Name
				// Only for configuration flags...
				fieldName, ok := configFlagsNames[flagName]
				if ok {
					v := flags.GetField(cfgOpts, fieldName)
					if v.IsValid() {
						switch v.Interface().(type) {
						case int:
							// Store the flag value (it equals to the default when no flag)
							flagValue, _ := c.Flags().GetInt(flagName)
							// Set the value coming from environment variable or config file (viper)
							value := viper.GetInt64(flagName)
							if value != 0 && fmt.Sprintf("%d", value) != configFlagsDefaults[flagName] {
								v.SetInt(value)
							}
							// Flag value takes precedence nevertheless
							// Re-set the field when the flag value was not equal to the default value
							if fmt.Sprintf("%d", flagValue) != configFlagsDefaults[flagName] {
								v.SetInt(int64(flagValue))
							}
						case string:
							// Store the flag value (it equals to the default when no flag)
							flagValue, _ := c.Flags().GetString(flagName)
							// Set the value coming from environment variable or config file (viper)
							value := viper.GetString(flagName)
							if value != "" && value != configFlagsDefaults[flagName] {
								v.SetString(value)
							}
							// Flag value takes precedence nevertheless
							// Re-set the field when the flag value was not equal to the default value
							if flagValue != configFlagsDefaults[flagName] {
								v.SetString(flagValue)
							}
						default:
						}
					}
				}
			})

			// Validate the config options
			// NOTE > It must happen after the precedence mechanism
			if errors := cfgOpts.Validate(); errors != nil {
				ret := "invalid configuration options/flags"
				for _, e := range errors {
					ret += "\n       "
					ret += e.Error()
				}

				return fmt.Errorf(ret)
			}

			// Transform the config options values
			if err := cfgOpts.Transform(c.Context()); err != nil {
				return err
			}

			// Set the context with the actual configuration values
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(c.Context(), time.Second*time.Duration(cfgOpts.Timeout))
			ctx = context.WithValue(ctx, pkgcontext.ContextCancelFuncKey, cancel)
			c.SetContext(ctx)

			return nil
		},
		PersistentPostRunE: func(c *cobra.Command, args []string) error {
			contextCancel, ok := c.Context().Value(pkgcontext.ContextCancelFuncKey).(context.CancelFunc)
			if !ok {
				return fmt.Errorf("couldn't obtain configuration options")
			}

			defer contextCancel()

			return nil
		},
		// Uncomment the following line if your application has an action associated with it
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	// Cobra supports persistent flags, which, if defined here, will be global to the whole application
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file (default is $HOME/.lstn.yaml)")
	err := rootCmd.MarkPersistentFlagFilename("config", "yaml")
	if err != nil {
		return nil, err
	}

	// Cobra also supports local flags, which will only run when this action is called directly
	rootOpts, err := options.NewRoot()
	if err != nil {
		return nil, err
	}
	rootOpts.Attach(rootCmd)

	// Tell viper to populate variables from the configuration file
	err = viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		return nil, err
	}

	// Pass the configuration options through the context
	ctx = context.WithValue(ctx, pkgcontext.ConfigKey, &rootOpts.ConfigFlags)

	// Store the version in the context
	vers := lstnversion.Get()
	ctx = context.WithValue(ctx, pkgcontext.VersionTagKey, vers.Tag)
	ctx = context.WithValue(ctx, pkgcontext.VersionShortKey, vers.Short)
	ctx = context.WithValue(ctx, pkgcontext.VersionLongKey, vers.Long)

	// Setup the core group
	rootCmd.AddGroup(&groups.Core)

	// Setup the `in` subcommand
	inCmd, err := in.New(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(inCmd)

	// Setup the `to` subcommand
	toCmd, err := to.New(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(toCmd)

	// Setup the `version` subcommand
	versionCmd, err := version.New(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(versionCmd)

	// Setup the help topics subcommands
	for t := range pkghelp.Topics {
		rootCmd.AddCommand(pkghelp.NewTopic(t))
	}

	// Setup help and completion subcommands
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	for _, c := range rootCmd.Commands() {
		switch c.Name() {
		case "help":
			c.Use = "help [command]"
			c.DisableFlagsInUseLine = true
			c.DisableAutoGenTag = true
			flagusages.Set(c)
		case "completion":
			completions := []string{}
			for _, sub := range c.Commands() {
				completions = append(completions, sub.Name())
				sub.DisableAutoGenTag = true
				sub.Annotations = map[string]string{
					"source": project.GetSourceURL(filename),
				}
				flagusages.Set(sub)
			}
			c.Use = fmt.Sprintf("completion <%s>", strings.Join(completions, "|"))
			c.DisableFlagsInUseLine = true
			c.DisableAutoGenTag = true
			c.Annotations = map[string]string{
				"source": project.GetSourceURL(filename),
			}
			flagusages.Set(c)
		}
	}

	// Fallback to the default subcommand when the user doesn't specify one explicitly.
	c, _, err := rootCmd.Find(os.Args[1:])
	if err == nil && c.Use == rootCmd.Use && c.Flags().Parse(os.Args[1:]) != pflag.ErrHelp {
		args := append([]string{inCmd.Name()}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	return &Command{rootCmd, ctx}, nil
}

type ExitCode int

const (
	exitOK     ExitCode = 0
	exitError  ExitCode = 1
	exitCancel ExitCode = 2
	exitAuth   ExitCode = 4
)

// Boot adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (c *Command) Go() ExitCode {
	err := c.cmd.ExecuteContext(c.ctx)
	if err != nil {
		if ctxErr := pkgcontext.Error(c.ctx, err); ctxErr != nil {
			return exitCancel
		}

		// Proxy jq halt errors as they are
		// NOTE > The default halt_error exit code is 5 but user can specify other exit codes
		if err, ok := err.(*jq.HaltError); ok {
			return ExitCode(err.ExitCode())
		}

		return exitError
	}

	return exitOK
}

func (c *Command) Command() *cobra.Command {
	return c.cmd
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".lstn" (without extension)
		viper.AddConfigPath(home)
		// TODO > Add current working directory as config path too?
		viper.SetConfigType("yaml")
		viper.SetConfigName(".lstn")
	}

	viper.AutomaticEnv() // Read in environment variables that match
	viper.SetEnvPrefix(flags.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", flags.EnvSeparator))
}
