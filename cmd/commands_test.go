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
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/listendev/lstn/cmd/root"
	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

type testCase struct {
	envvar  map[string]string
	cmdline []string
	stdout  string
	stderr  string
	errstr  string
}

func TestChildCommands(t *testing.T) {
	cases := []testCase{
		{
			cmdline: []string{"in"},
			stdout:  "",
			stderr:  "Running without a configuration file\nError: directory _CWD_ does not contain a package.json file\n",
			errstr:  "directory _CWD_ does not contain a package.json file",
		},
		{
			cmdline: []string{"in", "--help"},
			stdout: heredoc.Doc(`Query listen.dev for the verdicts of all the dependencies in your project.

Using this command, you can audit all the dependencies configured for a project and obtain their verdicts.
This requires a package.json file to fetch the package name and version of the project dependencies.

The verdicts it returns are listed by the name of each package and its specified version.

Usage:
  lstn in <path>

Examples:
  lstn in
  lstn in .
  lstn in /we/snitch
  lstn in sub/dir

Flags:
  -q, --jq string   filter the output using a jq expression
      --json        output the verdicts (if any) in JSON form

Config Flags:
      --endpoint string   the listen.dev endpoint emitting the verdicts (default "https://npm.listen.dev")
      --loglevel string   set the logging level (default "info")
      --timeout int       set the timeout, in seconds (default 60)

Debug Flags:
      --debug-options   output the options, then exit

Registry Flags:
      --npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")

Token Flags:
      --gh-token string   set the GitHub token

Global Flags:
      --config string   config file (default is $HOME/.lstn.yaml)
`),
			stderr: "",
			errstr: "",
		},
		{
			cmdline: []string{"in", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"gh-token": "",
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		{
			envvar: map[string]string{
				"LSTN_ENDPOINT": "https://npm-staging.listen.dev/",
			},
			cmdline: []string{"in", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm-staging.listen.dev",
	"gh-token": "",
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		{
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-token": "",
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		{
			cmdline: []string{"scan", "--debug-options", "--gh-token", "xxx"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-token": "xxx",
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		{
			envvar: map[string]string{
				"LSTN_GH_TOKEN":     "yyy",
				"LSTN_NPM_REGISTRY": "https://registry.npmjs.com",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-token": "yyy",
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.com",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		{
			cmdline: []string{"version", "--debug-options"},
			stdout: heredoc.Doc(`{
	"changelog": false,
	"debug-options": true,
	"verbosity": 0
}
`),
			stderr: "",
			errstr: "",
		},
		{
			cmdline: []string{"version", "--debug-options", "-v"},
			stdout: heredoc.Doc(`{
	"changelog": false,
	"debug-options": true,
	"verbosity": 1
}
`),
			stderr: "",
			errstr: "",
		},
		{
			cmdline: []string{"version", "--debug-options", "-vv"},
			stdout: heredoc.Doc(`{
	"changelog": false,
	"debug-options": true,
	"verbosity": 2
}
`),
			stderr: "",
			errstr: "",
		},
	}

	cwd, _ := os.Getwd()

	for _, tc := range cases {
		tc := tc

		envdesc := strings.Join(maps.Keys(tc.envvar), " ")

		desc := strings.Join(tc.cmdline, " ")
		rootC, e := root.New(context.Background())
		assert.Nil(t, e)
		c := rootC.Command()

		t.Run(fmt.Sprintf("%s %s%s", c.Name(), desc, envdesc), func(t *testing.T) {
			closer := internaltesting.EnvSetter(tc.envvar)
			t.Cleanup(closer)

			stdOut, stdErr, err := internaltesting.ExecuteCommand(c, tc.cmdline...)

			tc.stdout = strings.ReplaceAll(tc.stdout, "_CWD_", cwd)
			tc.stderr = strings.ReplaceAll(tc.stderr, "_CWD_", cwd)

			if tc.errstr != "" {
				if assert.Error(t, err) {
					tc.errstr = strings.ReplaceAll(tc.errstr, "_CWD_", cwd)
					assert.Equal(t, tc.errstr, err.Error())
				}
			}
			assert.Equal(t, tc.stdout, stdOut)
			assert.Equal(t, tc.stderr, stdErr)
		})
	}
}
