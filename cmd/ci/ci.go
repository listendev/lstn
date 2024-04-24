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
package ci

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/groups"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var ciCmd = &cobra.Command{
		Use:                   "ci",
		GroupID:               groups.Core.ID,
		DisableFlagsInUseLine: true,
		Short:                 "Listen in on what your CI does",
		Long: `Eavesdrop everything happening under the hoods into your CI.

Using this command, you can spy network and file activities occurring in your CI, whether it's your dependencies doing something shady or you.
This command requires a listen.dev pro account.`,
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, _ []string) error {
			ctx = c.Context()
			// Obtain the local options from the context
			opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.CiKey)
			if err != nil {
				return err
			}
			ciOpts, ok := opts.(*options.Ci)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			errs := []error{}
			if err := validate.Singleton.Var(ciOpts.Token.GitHub, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(ciOpts, "ConfigFlags.Token.GitHub")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}
			if err := validate.Singleton.Var(ciOpts.Token.JWT, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(ciOpts, "ConfigFlags.Token.JWT")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}
			if len(errs) > 0 {
				ret := "invalid configuration options/flags"
				for _, e := range errs {
					ret += "\n       "
					ret += e.Error()
				}

				return fmt.Errorf(ret)
			}

			if ciOpts.DebugOptions {
				c.Println(ciOpts.AsJSON())

				return nil
			}

			// FIXME: exit if not on linux

			info, infoErr := ci.NewInfo()
			if infoErr != nil {
				return fmt.Errorf("not running in a CI environment")
			}

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			cs := io.ColorScheme()

			if !info.IsGitHubPullRequest() {
				c.PrintErrln(cs.WarningIcon(), "lstn ci only runs on GitHub pull requests at the moment")

				return nil
			}
			// Block when running on fork pull requests
			if info.HasReadOnlyGitHubToken() {
				c.PrintErrln(cs.WarningIcon(), "lstn ci doesn not run on fork pull requests at the moment")

				return nil
			}

			io.StartProgressIndicator()
			envConfig := fmt.Sprintf("%s\n%s=%s\n%s=%s\n", info.Dump(), "LISTENDEV_TOKEN", ciOpts.Token.JWT, "GITHUB_TOKEN", ciOpts.Token.GitHub)
			envDirErr := os.MkdirAll("/var/run/argus", 0750)
			if envDirErr != nil {
				io.StopProgressIndicator()

				return envDirErr
			}

			envConfigFilename := "/var/run/argus/default"
			if err := os.WriteFile(envConfigFilename, []byte(envConfig), 0640); err != nil {
				io.StopProgressIndicator()

				return err
			}
			io.StopProgressIndicator()
			c.Println(cs.SuccessIcon(), "Wrote config", cs.Magenta(envConfigFilename))

			io.StartProgressIndicator()
			var argus *exec.Cmd
			if len(ciOpts.Directory) > 0 {
				file := filepath.Join(ciOpts.Directory, "argus")
				info, err := os.Stat(file)
				if os.IsNotExist(err) {
					io.StopProgressIndicator()

					return fmt.Errorf("couldn't find the argus binary in %s", ciOpts.Directory)
				}
				if info.IsDir() {
					io.StopProgressIndicator()

					return fmt.Errorf("expecting %s to be an executable file", file)
				}
				argus = exec.CommandContext(ctx, file, "-s", "enable-now")
			} else {
				exe, err := exec.LookPath("argus")
				if err != nil {
					io.StopProgressIndicator()

					return fmt.Errorf("couldn't find the argus executable in the PATH")
				}
				argus = exec.CommandContext(ctx, exe, "-s", "enable-now")
			}
			io.StopProgressIndicator()
			c.Println(cs.Blue("•"), "Install and enable", cs.Magenta(argus.String()))

			io.StartProgressIndicator()
			argusOut, argusErr := argus.CombinedOutput()
			if argusErr != nil {
				io.StopProgressIndicator()

				return fmt.Errorf("couldn't install and enable argus: %w", argusErr)
			}
			io.StopProgressIndicator()
			c.Println(string(bytes.Trim(argusOut, "\n")))

			return nil
		},
	}

	// Obtain the local options
	ciOpts, err := options.NewCi()
	if err != nil {
		return nil, err
	}

	// Persistent flags here will work for this command and all subcommands
	// ciCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags will only run when this command is called directly
	ciOpts.Attach(ciCmd, []string{"--ignore-packages", "--ignore-deptypes", "--select", "lockfiles", "npm-endpoint", "pypi-endpoint", "reporter", "npm-registry", "gh-owner", "gh-pull-id", "gh-repo"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.CiKey, ciOpts)
	ciCmd.SetContext(ctx)

	return ciCmd, nil
}
