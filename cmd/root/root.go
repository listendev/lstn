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
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/XANi/goneric"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/cmd/in"
	"github.com/listendev/lstn/cmd/scan"
	"github.com/listendev/lstn/cmd/to"
	"github.com/listendev/lstn/cmd/version"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/flagusages"
	"github.com/listendev/lstn/pkg/cmd/groups"
	pkghelp "github.com/listendev/lstn/pkg/cmd/help"
	"github.com/listendev/lstn/pkg/cmd/options"
	lstnviper "github.com/listendev/lstn/pkg/cmd/viper"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/jq"
	lstnversion "github.com/listendev/lstn/pkg/version"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
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
	rootOpts, rootOptsErr := options.NewRoot()
	if rootOptsErr != nil {
		return nil, rootOptsErr
	}

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
			withConfigFile := false
			c, _, err := c.Find(os.Args[1:])
			if err == nil && (c.IsAvailableCommand() && c.GroupID == groups.Core.ID) {
				// If a config file is found, read it in
				if err := viper.ReadInConfig(); err == nil {
					c.Printf("Using config file: %s\n", viper.ConfigFileUsed())
					withConfigFile = true
				} else {
					if _, ok := err.(viper.ConfigFileNotFoundError); ok {
						// Config file not found, ignore...
						c.PrintErrln("Running without a configuration file")
					} else {
						// Config file was found but another error was produced
						c.PrintErrf("Error running with config file: %s\n", viper.ConfigFileUsed())
					}
				}
			}

			// Obtain the configuration options
			cfgOpts, ok := c.Context().Value(pkgcontext.ConfigKey).(*flags.ConfigFlags)
			if !ok {
				return fmt.Errorf("couldn't obtain configuration options")
			}

			// Update them with the values in the config file (if any)
			if withConfigFile {
				viperOpts := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
					mapstructure.StringToTimeDurationHookFunc(),
					mapstructure.StringToSliceHookFunc(","),
					lstnviper.StringToReportType(),
				))
				if err := viper.Unmarshal(&cfgOpts, viperOpts); err != nil {
					return err
				}
			}

			// Update them with environment variable values
			// Obtain the mapping flag name -> struct field name
			configFlagsNames := flags.GetNames(cfgOpts)
			// Obtain the mapping flag name -> default value
			configFlagsDefaults := flags.GetDefaults(cfgOpts)
			// Implement flag precedence over environment variables, over configuration file
			var flagErr error
			c.Flags().VisitAll(func(f *pflag.Flag) {
				flagName := f.Name
				// Only for configuration flags...
				fieldName, ok := configFlagsNames[flagName]
				if ok {
					v := flags.GetField(cfgOpts, fieldName)
					defaultVal, hasDefault := configFlagsDefaults[flagName]
					if v.IsValid() {
						switch v.Interface().(type) {
						case int:
							// Store the flag value (it equals to the default when no flag)
							flagValue, _ := c.Flags().GetInt(flagName)
							// Set the value coming from environment variable or config file (viper)
							value := viper.GetInt64(flagName)
							if value != 0 || (hasDefault && fmt.Sprintf("%d", value) != defaultVal) {
								v.SetInt(value)
							}
							// Flag value takes precedence nevertheless
							// Re-set the field when the flag value was not equal to the default value or to the zero value
							if (!hasDefault && flagValue != 0) || (hasDefault && fmt.Sprintf("%d", flagValue) != defaultVal) {
								v.SetInt(int64(flagValue))
							}
						case string:
							// Store the flag value (it equals to the default when no flag)
							flagValue, _ := c.Flags().GetString(flagName)
							// Set the value coming from environment variable or config file (viper)
							value := viper.GetString(flagName)
							if value != "" && value != defaultVal {
								v.SetString(value)
							}
							// Flag value takes precedence nevertheless
							// Re-set the field when the flag value was not equal to the default value
							if flagValue != defaultVal {
								v.SetString(flagValue)
							}
						case []string:
							// Store the flag value (it equals to the default when no flag)
							flagValue, _ := c.Flags().GetStringSlice(flagName)
							// Set the value (string slice) coming from environment variable or config file (viper)
							// This fallbacks to the default value when no environment variable or config file
							value := viper.GetStringSlice(flagName)
							if len(value) > 0 {
								res := []string{}
								for _, elem := range value {
									res = append(res, strings.Split(elem, ",")...)
								}
								v.Set(reflect.ValueOf(goneric.SliceDedupe(res)))
							}
							// Grab the actual default value
							actualDefaultVal := []string{}
							if defaultVal != "" && defaultVal != "[]" {
								if err := json.Unmarshal([]byte(defaultVal), &actualDefaultVal); err != nil {
									flagErr = err

									return
								}
							}
							// Use the flag slice value when it's not empty and it's different (order doesn't matter) than the default
							if len(flagValue) > 0 && !goneric.CompareSliceSet(flagValue, actualDefaultVal) {
								v.Set(reflect.ValueOf(flagValue))
							}
						case []cmd.ReportType:
							// Store the flag value (it equals to the default when no flag)
							enumFlag, _ := c.Flags().Lookup(flagName).Value.(*enumflag.EnumFlagValue[cmd.ReportType])
							flagValue := enumFlag.Get().([]cmd.ReportType)
							// Set the value coming from environment variable or config file (viper)
							value := viper.GetString(flagName)
							if value != "[]" && value != defaultVal {
								reportTypeErr := enumFlag.Set(value)
								if reportTypeErr != nil {
									flagErr = fmt.Errorf("%s %s; got %s", flagName, reportTypeErr.Error(), value)

									return
								}
								// Substitute the slice
								v.Set(reflect.ValueOf(enumFlag.Get()))
							}
							// Use the flag slice value when it's not empty
							if len(flagValue) > 0 {
								reportTypeErr := enumFlag.Set(strings.Join(goneric.Map(func(t cmd.ReportType) string {
									return t.String()
								}, flagValue...), ","))
								if reportTypeErr != nil {
									flagErr = fmt.Errorf("%s %s; got %s", flagName, reportTypeErr.Error(), flagValue)

									return
								}
								// Substitute the slice
								v.Set(reflect.ValueOf(enumFlag.Get()))
							}
						default:
						}
					}
				}
			})
			// Some custom flags (eg. enum flags) may have their own validation mechanism
			if flagErr != nil {
				return flagErr
			}

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

			io := iostreams.System()
			ctx = context.WithValue(ctx, pkgcontext.IOStreamsKey, io)
			c.SetContext(ctx)

			if rootOpts.DebugOptions {
				c.Println(rootOpts.AsJSON())

				return nil
			}

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
	persistentFlagFilenameErr := rootCmd.MarkPersistentFlagFilename("config", "yaml")
	if persistentFlagFilenameErr != nil {
		return nil, persistentFlagFilenameErr
	}

	// Cobra also supports local flags, which will only run when this action is called directly
	rootOpts.Attach(rootCmd, []string{})

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

	// Setup the `scan` subcommand
	scanCmd, err := scan.New(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(scanCmd)

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
		args := append([]string{scanCmd.Name()}, os.Args[1:]...)
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
	cobra.OnFinalize(cleanConfig)
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
}

func cleanConfig() {
	cfgFile = ""
	viper.Reset()
}
