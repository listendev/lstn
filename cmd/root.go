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
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/listendev/lstn/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const defaultCommand = "in"

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lstn",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(c *cobra.Command, args []string) error {
		cfgOpts, ok := c.Context().Value(pkgcontext.ConfigKey).(*flags.ConfigOptions)
		if !ok {
			return fmt.Errorf("couldn't obtain configuration options")
		}

		// Obtain the mapping flag name -> struct field name
		configFlagsNames := flags.GetConfigFlagsNames()
		// Obtain the mapping flag name -> default value
		configFlagsDefaults := flags.GetConfigFlagsDefaults()

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

		if errors := cfgOpts.Validate(); errors != nil {
			ret := "invalid configuration options/flags"
			for _, e := range errors {
				ret += "\n       --"
				ret += e.Error()
			}
			return fmt.Errorf(ret)
		}

		return nil
	},
	// Uncomment the following line if your application has an action associated with it
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Fallback to the default subcommand when the user doesn't specify one explicitly.
	c, _, err := rootCmd.Find(os.Args[1:])
	if err == nil && c.Use == rootCmd.Use && c.Flags().Parse(os.Args[1:]) != pflag.ErrHelp {
		args := append([]string{defaultCommand}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cfgOpts := flags.NewConfigOptions()

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here, will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file (default is $HOME/.lstn.yaml)")
	rootCmd.MarkPersistentFlagFilename("config", "yaml")

	// Cobra also supports local flags, which will only run when this action is called directly.
	// TODO > Make these persistent?
	rootCmd.PersistentFlags().StringVar(&cfgOpts.LogLevel, "loglevel", cfgOpts.LogLevel, "log level")
	rootCmd.PersistentFlags().IntVar(&cfgOpts.Timeout, "timeout", cfgOpts.Timeout, "timeout in seconds")
	rootCmd.PersistentFlags().StringVar(&cfgOpts.Endpoint, "endpoint", cfgOpts.Endpoint, "the listen.dev endpoint emitting the verdicts")

	// Tell viper to populate variables from the configuration file
	viper.BindPFlags(rootCmd.Flags())

	// Pass the configuration options through the context
	ctx := context.WithValue(context.Background(), pkgcontext.ConfigKey, cfgOpts)
	rootCmd.SetContext(ctx)
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
	viper.SetEnvPrefix("lstn")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
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
