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
package root

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/listendev/lstn/cmd/in"
	"github.com/listendev/lstn/cmd/to"
	"github.com/listendev/lstn/cmd/version"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/help"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/jq"
	lstnversion "github.com/listendev/lstn/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

type command struct {
	cmd *cobra.Command
	ctx context.Context
}

func New(ctx context.Context) (*command, error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "lstn",
		SilenceUsage: true,
		Short:        "Analyze the behavior of your dependencies using listen.dev",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			// Do not check for the config file if the command is not available (eg., help) or not core (eg., version)
			c, _, err := c.Find(os.Args[1:])
			if err == nil && (c.IsAvailableCommand() || c.GroupID == groups.Core.ID) {
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

			// TODO ? ignore global/config flags for non core commands

			cfgOpts, ok := c.Context().Value(pkgcontext.ConfigKey).(*flags.ConfigOptions)
			if !ok {
				return fmt.Errorf("couldn't obtain configuration options")
			}

			// Obtain the mapping flag name -> struct field name
			configFlagsNames := flags.GetConfigFlagsNames()
			// Obtain the mapping flag name -> default value
			configFlagsDefaults := flags.GetConfigFlagsDefaults()
			// Implement flag precedence over environment variables, over configuration file
			c.Flags().VisitAll(func(f *pflag.Flag) {
				flagName := f.Name
				// Only for configuration flags...
				fieldName, ok := configFlagsNames[flagName]
				if ok {
					v := cfgOpts.GetField(fieldName)
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
			ctx, cancel := context.WithTimeout(c.Context(), time.Second*time.Duration(cfgOpts.Timeout))
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

	// Obtain the configuration options
	cfgOpts, err := flags.NewConfigOptions()
	if err != nil {
		return nil, err
	}

	// Cobra supports persistent flags, which, if defined here, will be global to the whole application
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file (default is $HOME/.lstn.yaml)")
	err = rootCmd.MarkPersistentFlagFilename("config", "yaml")
	if err != nil {
		return nil, err
	}

	// Cobra also supports local flags, which will only run when this action is called directly
	rootCmd.PersistentFlags().StringVar(&cfgOpts.LogLevel, "loglevel", cfgOpts.LogLevel, "log level")
	rootCmd.PersistentFlags().IntVar(&cfgOpts.Timeout, "timeout", cfgOpts.Timeout, "timeout in seconds")
	rootCmd.PersistentFlags().StringVar(&cfgOpts.Endpoint, "endpoint", cfgOpts.Endpoint, "the listen.dev endpoint emitting the verdicts")

	// Tell viper to populate variables from the configuration file
	err = viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		return nil, err
	}

	// Pass the configuration options through the context
	ctx = context.WithValue(ctx, pkgcontext.ConfigKey, cfgOpts)

	// Store the version in the context
	shortVersion, longVersion := lstnversion.Version()
	ctx = context.WithValue(ctx, pkgcontext.ShortVersionKey, shortVersion)
	ctx = context.WithValue(ctx, pkgcontext.LongVersionKey, longVersion)

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
	for t := range help.Topics {
		rootCmd.AddCommand(help.NewTopic(t))
	}

	// Fallback to the default subcommand when the user doesn't specify one explicitly.
	c, _, err := rootCmd.Find(os.Args[1:])
	if err == nil && c.Use == rootCmd.Use && c.Flags().Parse(os.Args[1:]) != pflag.ErrHelp {
		args := append([]string{inCmd.Name()}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	return &command{rootCmd, ctx}, nil
}

type exitCode int

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
	exitAuth   exitCode = 4
)

// Boot adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func (c *command) Go() exitCode {
	err := c.cmd.ExecuteContext(c.ctx)
	if err != nil {
		if ctxErr := pkgcontext.ContextError(c.ctx, err); ctxErr != nil {
			return exitCancel
		}

		// Proxy jq halt errors as they are
		// NOTE > The default halt_error exit code is 5 but user can specify other exit codes
		if err, ok := err.(*jq.HaltError); ok {
			return exitCode(err.ExitCode())
		}

		return exitError
	}

	return exitOK
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set
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