// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2025 The listen.dev team <engineering@garnet.ai>
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
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/listendev/pkg/apispec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var _, filename, _, _ = runtime.Caller(0)

func New(ctx context.Context) (*cobra.Command, error) {
	c := &cobra.Command{
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

			// Token options are mandatory in this case
			errs := []error{}
			if err := validate.Singleton.Var(opts.Token.GitHub, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(opts, "ConfigFlags.Token.GitHub")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}
			if err := validate.Singleton.Var(opts.JWT, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(opts, "ConfigFlags.Token.JWT")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}

			if len(errs) > 0 {
				ret := "invalid configuration options/flags"
				for _, e := range errs {
					ret += "\n       "
					ret += e.Error()
				}

				return fmt.Errorf("%s", ret)
			}

			if opts.DebugOptions {
				c.Println(opts.AsJSON())

				return nil
			}

			isLocal := opts.Endpoint.IsLocalCore()

			info, infoErr := ci.NewInfo()
			if !isLocal && infoErr != nil {
				return fmt.Errorf("not running in a CI environment")
			}

			io := c.Context().Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)
			cs := io.ColorScheme()

			// Block when running on fork pull requests
			if !isLocal && info.HasReadOnlyGitHubToken() {
				c.PrintErrln(cs.WarningIcon(), "lstn ci doesn not run on fork pull requests at the moment")

				return nil
			}

			// Locate the jibril binary
			io.StartProgressIndicator()
			jibrilFile, err := jibrilFile(opts.Directory)
			if err != nil {
				io.StopProgressIndicator()

				return err
			}

			io.StopProgressIndicator()
			c.Println(cs.SuccessIcon(), "Found jibril binary", cs.Magenta(jibrilFile))

			// Install jibril
			{
				io.StartProgressIndicator()
				jibril := exec.CommandContext(ctx, jibrilFile, "--systemd", "install")
				jibrilOut, err := jibril.CombinedOutput()
				if err != nil {
					io.StopProgressIndicator()

					return fmt.Errorf("couldn't install and enable jibril: %w", err)
				}

				c.Println(cs.SuccessIcon(), "Installed jibril", cs.Magenta(string(bytes.Trim(jibrilOut, "\n"))))
				io.StopProgressIndicator()
			}

			// create a client for the Core API
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

			// get settings from the Core API
			coreSettings, err := settings(ctx, coreClient, c, cs)
			if err != nil {
				io.StopProgressIndicator()

				return err
			}
			io.StopProgressIndicator()
			c.Println(cs.SuccessIcon(), "Fetch settings")

			githubRepository := os.Getenv("GITHUB_REPOSITORY")
			if githubRepository == "" {
				c.PrintErrln(cs.WarningIcon(), "GITHUB_REPOSITORY is empty, if you are in a dev machine please set something")

				return fmt.Errorf("missing GITHUB_REPOSITORY var")
			}

			githubRepositoryID := os.Getenv("GITHUB_REPOSITORY_ID")
			if githubRepositoryID == "" {
				c.PrintErrln(cs.WarningIcon(), "GITHUB_REPOSITORY_ID is empty, if you are in a dev machine please set something")

				return fmt.Errorf("missing GITHUB_REPOSITORY_ID var")
			}

			{ // get yaml config for jibril
				io.StartProgressIndicator()
				config, err := coreClient.GetConfigWithResponse(ctx)
				if err != nil || config.StatusCode() != http.StatusOK {
					io.StopProgressIndicator()

					c.PrintErrln(cs.WarningIcon(), "couldn't fetch the configuration for jibril", cs.Redf("(%d)", config.StatusCode()))

					return err
				}

				jibrilConfig := *config.JSON200

				if err := writeConfigToYaml(jibrilConfig); err != nil {
					c.PrintErrln(cs.WarningIcon(), "couldn't write the configuration YAML for jibril")

					return err
				}

				io.StopProgressIndicator()
				c.Println(cs.SuccessIcon(), "Wrote jibril config", cs.Magenta("/etc/jibril/config.yaml"))
			}

			{ // fetch network policy for this specific workflow
				io.StartProgressIndicator()

				policy, err := coreClient.GetNetPolicyWithResponse(ctx, &apispec.GetNetPolicyParams{
					GithubRepository:   githubRepository,
					GithubRepositoryId: githubRepositoryID,
				})

				if err != nil || policy.StatusCode() != http.StatusOK {
					io.StopProgressIndicator()

					c.PrintErrln(cs.WarningIcon(), "couldn't fetch the network policy for jibril", cs.Redf("(%d)", policy.StatusCode()))

					return err
				}

				netPolicy := *policy.JSON200

				if err := writeNetworkPolicyToYaml(netPolicy); err != nil {
					c.PrintErrln(cs.WarningIcon(), "couldn't write the network policy YAML for jibril")

					return err
				}

				io.StopProgressIndicator()
				c.Println(cs.SuccessIcon(), "Wrote jibril network policy", cs.Magenta("/etc/jibril/netpolicy.yaml"))
			}

			{ // Create env file for runtime eavesdropping tool as a systemd service
				io.StartProgressIndicator()
				jibrilEnvConfigFilename, err := createEnvForJibrill(isLocal, coreSettings, info, opts.JWT, opts.Token.GitHub)
				if err != nil {
					io.StopProgressIndicator()

					return err
				}

				io.StopProgressIndicator()
				c.Println(cs.SuccessIcon(), "Wrote jibril env config", cs.Magenta(jibrilEnvConfigFilename))
			}

			{ // run jibril
				// jibril := exec.CommandContext(ctx, jibrilFile, "-s", "enable-now", "||", "journalctl", "-xeu", "jibril")
				jibril := exec.CommandContext(ctx, jibrilFile, "-s", "enable-now")

				io.StopProgressIndicator()
				c.Println(cs.Blue("•"), "Enable jibril", cs.Magenta(jibril.String()))
				io.StartProgressIndicator()

				jibrilOut, jibrilErr := jibril.CombinedOutput()
				if jibrilErr != nil {
					io.StopProgressIndicator()

					return fmt.Errorf("couldn't enable jibril: %w", jibrilErr)
				}
				io.StopProgressIndicator()
				c.Println(string(jibrilOut))
			}

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

func settings(ctx context.Context, client *apispec.ClientWithResponses, c *cobra.Command, cs *iostreams.ColorScheme) (*apispec.Settings, error) {
	coreResponse, coreResponseErr := client.GetApiV1SettingsWithResponse(ctx)
	if coreResponseErr != nil {
		return nil, coreResponseErr
	}

	if coreResponse.StatusCode() != http.StatusOK {
		c.PrintErrln(cs.WarningIcon(), "settings request to Core API didn't work out", cs.Redf("(%d)", coreResponse.StatusCode()))

		return nil, fmt.Errorf("status code is not %d", http.StatusOK)
	}

	if coreResponse.JSON200 == nil {
		c.PrintErrln(cs.WarningIcon(), "empty settings response")

		return nil, fmt.Errorf("got empty settings from the Core API")
	}

	return coreResponse.JSON200, nil
}

func jibrilFile(directory string) (string, error) {
	if len(directory) > 0 {
		file := filepath.Join(directory, "jibril")
		info, err := os.Stat(file)
		if os.IsNotExist(err) {
			return "", fmt.Errorf("couldn't find the jibril binary in %s", directory)
		}

		if info.IsDir() {
			return "", fmt.Errorf("expecting %s to be an executable file", file)
		}

		return file, nil
	}

	exe, err := exec.LookPath("jibril")
	if err != nil {
		return "", fmt.Errorf("couldn't find the jibril executable in the PATH")
	}

	return exe, nil
}

// createEnvForJibrill creates the files that jibril needs to run in the CI environment.
// It returns the path to the file that contains config for jibril.
func createEnvForJibrill(isLocal bool, settings *apispec.Settings, info *ci.Info, jwt string, ghToken string) (string, error) {
	var dump string
	if info != nil {
		dump = info.Dump()
	}

	envConfig := fmt.Sprintf("%s\n%s\n%s=%s\n%s=%s\n", strings.Join(settings.TokensSlice(), "\n"), dump, "LISTENDEV_TOKEN", jwt, "GITHUB_TOKEN", ghToken)

	dirPath := "/var/run/jibril"
	if isLocal {
		dirPath = "./jibril"
	}

	envDirErr := os.MkdirAll(dirPath, 0o750)
	if envDirErr != nil {
		return "", envDirErr
	}

	envConfigFilename := "/var/run/jibril/default"
	if isLocal {
		envConfigFilename = "./jibril/default"
	}

	if err := os.WriteFile(envConfigFilename, []byte(envConfig), 0o640); err != nil {
		return "", err
	}

	return envConfigFilename, nil
}

// writeConfigToYaml writes the configuration to a YAML file.
// The path is hardcoded to /etc/jibril/config.yaml because is the where jibril expects it.
func writeConfigToYaml(config apispec.JibrilConfig) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll("/etc/jibril", 0o750); err != nil {
		return err
	}

	yamlConfig := "/etc/jibril/config.yaml"

	// Create the YAML file
	f, err := os.Create(yamlConfig)
	if err != nil {
		return err
	}
	defer f.Close()

	// Encode the config to YAML
	yamlEncoder := yaml.NewEncoder(f)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(&config); err != nil {
		return err
	}

	return nil
}

// writeNetworkPolicyToYaml writes the network policy to a YAML file.
// The path is hardcoded to /etc/jibril/netpolicy.yaml because is the where jibril expects it, but you can change it
// putting a different path in the jibril configuration.
func writeNetworkPolicyToYaml(policy apispec.NetworkPolicy) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll("/etc/jibril", 0o750); err != nil {
		return err
	}

	yamlConfig := "/etc/jibril/netpolicy.yaml"
	// Create the YAML file
	f, err := os.Create(yamlConfig)
	if err != nil {
		return err
	}
	defer f.Close()

	type WrapPolicy struct {
		NetworkPolicy *apispec.NetworkPolicy `yaml:"network_policy"`
	}

	np := WrapPolicy{NetworkPolicy: &policy}

	// Encode the policy to YAML
	yamlEncoder := yaml.NewEncoder(f)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(&np); err != nil {
		return err
	}

	return nil
}
