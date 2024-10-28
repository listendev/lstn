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
package enable

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/pkg/apispec"
	"github.com/spf13/cobra"
)

var (
	_, filename, _, _ = runtime.Caller(0)
)

func New(ctx context.Context) (*cobra.Command, error) {
	var c = &cobra.Command{
		Use:                   "enable",
		DisableFlagsInUseLine: true,
		Short:                 "Enable the CI eavesdropping",
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, _ []string) error {
			ctx = c.Context()
			// Obtain the local options from the context
			optsFromContext, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.CiEnableKey)
			if err != nil {
				return err
			}
			opts, ok := optsFromContext.(*options.CiEnable)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			if opts.DebugOptions {
				c.Println(opts.AsJSON())

				return nil
			}

			// FIXME: exit if not on linux

			info, infoErr := ci.NewInfo()
			if infoErr != nil {
				return fmt.Errorf("not running in a CI environment")
			}

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			cs := io.ColorScheme()

			// Block when running on fork pull requests
			if info.HasReadOnlyGitHubToken() {
				c.PrintErrln(cs.WarningIcon(), "lstn ci doesn not run on fork pull requests at the moment")

				return nil
			}

			// Fetch settings from Core API
			io.StartProgressIndicator()
			coreClient, coreClientErr := apispec.NewClientWithResponses(
				opts.Endpoint.Core,
				apispec.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
					if req == nil {
						io.StopProgressIndicator()
						c.PrintErrln(cs.WarningIcon(), "empty settings request")

						return fmt.Errorf("couldn't prepare the settings request for the Core API")
					}
					req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", opts.JWT))

					return nil
				}),
			)
			if coreClientErr != nil {
				io.StopProgressIndicator()

				return coreClientErr
			}
			coreResponse, coreResponseErr := coreClient.GetApiV1SettingsWithResponse(ctx)
			if coreResponseErr != nil {
				io.StopProgressIndicator()

				return coreResponseErr
			}
			if coreResponse.StatusCode() != http.StatusOK {
				io.StopProgressIndicator()
				c.PrintErrln(cs.WarningIcon(), "settings request to Core API didn't work out", cs.Redf("(%d)", coreResponse.StatusCode()))

				return fmt.Errorf("status code is not %d", http.StatusOK)
			}
			if coreResponse.JSON200 == nil {
				io.StopProgressIndicator()
				c.PrintErrln(cs.WarningIcon(), "empty settings response")

				return fmt.Errorf("got empty settings from the Core API")
			}
			coreSettings := *coreResponse.JSON200
			io.StopProgressIndicator()
			c.Println(cs.SuccessIcon(), "Fetch settings")

			// Create configuration for runtime eavesdropping tool
			io.StartProgressIndicator()
			envConfig := fmt.Sprintf("%s\n%s\n%s=%s\n%s=%s\n", strings.Join(coreSettings.TokensSlice(), "\n"), info.Dump(), "LISTENDEV_TOKEN", opts.Token.JWT, "GITHUB_TOKEN", opts.Token.GitHub)
			envDirErr := os.MkdirAll("/var/run/jibril", 0750)
			if envDirErr != nil {
				io.StopProgressIndicator()

				return envDirErr
			}

			envConfigFilename := "/var/run/jibril/default"
			if err := os.WriteFile(envConfigFilename, []byte(envConfig), 0640); err != nil {
				io.StopProgressIndicator()

				return err
			}
			io.StopProgressIndicator()
			c.Println(cs.SuccessIcon(), "Wrote config", cs.Magenta(envConfigFilename))

			io.StartProgressIndicator()
			var jibril *exec.Cmd
			if len(opts.Directory) > 0 {
				file := filepath.Join(opts.Directory, "jibril")
				info, err := os.Stat(file)
				if os.IsNotExist(err) {
					io.StopProgressIndicator()

					return fmt.Errorf("couldn't find the jibril binary in %s", opts.Directory)
				}
				if info.IsDir() {
					io.StopProgressIndicator()

					return fmt.Errorf("expecting %s to be an executable file", file)
				}
				jibril = exec.CommandContext(ctx, file, "-s", "enable-now")
			} else {
				exe, err := exec.LookPath("jibril")
				if err != nil {
					io.StopProgressIndicator()

					return fmt.Errorf("couldn't find the jibril executable in the PATH")
				}
				jibril = exec.CommandContext(ctx, exe, "-s", "enable-now")
			}
			io.StopProgressIndicator()
			c.Println(cs.Blue("•"), "Install and enable", cs.Magenta(jibril.String()))

			io.StartProgressIndicator()
			jibrilOut, jibrilErr := jibril.CombinedOutput()
			if jibrilErr != nil {
				io.StopProgressIndicator()

				return fmt.Errorf("couldn't install and enable jibril: %w", jibrilErr)
			}
			io.StopProgressIndicator()
			c.Println(string(bytes.Trim(jibrilOut, "\n")))

			return nil
		},
	}

	// Create the local options
	enableOpts, err := options.NewCiEnable()
	if err != nil {
		return nil, err
	}
	// Local flags will only run when this command is called directly
	enableOpts.Attach(c, []string{"--ignore-packages", "--ignore-deptypes", "--select", "lockfiles", "npm-endpoint", "pypi-endpoint", "reporter", "npm-registry", "gh-owner", "gh-pull-id", "gh-repo"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.CiEnableKey, enableOpts)
	c.SetContext(ctx)

	return c, nil
}
