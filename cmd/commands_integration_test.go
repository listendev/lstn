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
	"os"
	"path"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/listendev/lstn/cmd/root"
	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name    string
	envvar  map[string]string
	cmdline []string
	stdout  string
	stderr  string
	errstr  string
}

func TestChildCommands(t *testing.T) {
	cwd, _ := os.Getwd()

	cases := []testCase{
		// lstn in
		{
			name:    "lstn in",
			cmdline: []string{"in"},
			stdout:  "",
			stderr:  "Running without a configuration file\nError: directory _CWD_ does not contain the package-lock.json file\n",
			errstr:  "directory _CWD_ does not contain the package-lock.json file",
		},
		// lstn to
		{
			name: "lstn to",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"to"},
			stdout:  "",
			stderr:  "Error: requires at least 1 arg (package name)\n",
			errstr:  "requires at least 1 arg (package name)",
		},
		// lstn to --help
		{
			name:    "lstn to --help",
			cmdline: []string{"to", "--help"},
			stdout: heredoc.Doc(`Query listen.dev for the verdicts of a package.

Using this command, you can audit a single package version or all the versions of a package and obtain their verdicts.

Specifying the package name is mandatory.

It lists out the verdicts of all the versions of the input package name.

Usage:
  lstn to <name> [[version] [shasum] | [version constraint]]

Examples:
  # Get the verdicts for all the chalk versions that listen.dev owns
  lstn to chalk
  lstn to debug 4.3.4
  lstn to react 18.0.0 b468736d1f4a5891f38585ba8e8fb29f91c3cb96

  # Get the verdicts for all the existing chalk versions
  lstn to chalk "*"
  # Get the verdicts for nock versions >= 13.2.0 and < 13.3.0
  lstn to nock "~13.2.x"
  # Get the verdicts for tap versions >= 16.3.0 and < 16.4.0
  lstn to tap "^16.3.0"
  # Get the verdicts for prettier versions >= 2.7.0 <= 3.0.0
  lstn to prettier ">=2.7.0 <=3.0.0"

Flags:
      --json   output the verdicts (if any) in JSON form

Config Flags:
      --loglevel string        set the logging level (default "info")
      --npm-endpoint string    the listen.dev endpoint emitting the NPM verdicts (default "https://npm.listen.dev")
      --pypi-endpoint string   the listen.dev endpoint emitting the PyPi verdicts (default "https://pypi.listen.dev")
      --timeout int            set the timeout, in seconds (default 60)

Debug Flags:
      --debug-options   output the options, then exit

Filtering Flags:
  -q, --jq string       filter the output verdicts using a jq expression (requires --json)
  -s, --select string   filter the output verdicts using a jsonpath script expression (server-side)

Registry Flags:
      --npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")

Global Flags:
      --config string   config file (default is $HOME/.lstn.yaml)
`),
			stderr: "",
			errstr: "",
		},
		// lstn scan --help
		{
			name: "lstn scan --help",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--help"},
			stdout: heredoc.Doc(`Query listen.dev for the verdicts of the dependencies in your project.

Using this command, you can audit the first-level dependencies configured for a project and obtain their verdicts.
This requires a package.json file to fetch the package name and version of the project dependencies.

The verdicts it returns are listed by the name of each package and its specified version.

Usage:
  lstn scan [path]

Examples:
  lstn scan
  lstn scan .
  lstn scan sub/dir
  lstn scan /we/snitch
  lstn scan /we/snitch --ignore-deptypes peer
  lstn scan /we/snitch --ignore-deptypes dev,peer
  lstn scan /we/snitch --ignore-deptypes dev --ignore-deptypes peer
  lstn scan /we/snitch --ignore-packages react,glob --ignore-deptypes peer
  lstn scan /we/snitch --ignore-packages react --ignore-packages glob,@vue/devtools

Flags:
      --json   output the verdicts (if any) in JSON form

Config Flags:
      --loglevel string        set the logging level (default "info")
      --npm-endpoint string    the listen.dev endpoint emitting the NPM verdicts (default "https://npm.listen.dev")
      --pypi-endpoint string   the listen.dev endpoint emitting the PyPi verdicts (default "https://pypi.listen.dev")
      --timeout int            set the timeout, in seconds (default 60)

Debug Flags:
      --debug-options   output the options, then exit

Filtering Flags:
      --ignore-deptypes (dep,dev,optional,peer)   the list of dependencies types to not process (default [bundle])
      --ignore-packages strings                   the list of packages to not process
  -q, --jq string                                 filter the output verdicts using a jq expression (requires --json)
  -s, --select string                             filter the output verdicts using a jsonpath script expression (server-side)

Registry Flags:
      --npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")

Reporting Flags:
      --gh-owner string                                               set the GitHub owner name (org|user)
      --gh-pull-id int                                                set the GitHub pull request ID
      --gh-repo string                                                set the GitHub repository name
  -r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review,pro)   set one or more reporters to use (default [])

Token Flags:
      --gh-token string   set the GitHub token

Global Flags:
      --config string   config file (default is $HOME/.lstn.yaml)
`),
			stderr: "",
			errstr: "",
		},
		// lstn to --debug-options
		{
			name: "lstn to --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"to", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn to --debug-options --npm-registry https://some.io --timeout 2222
		{
			name: "lstn to --debug-options --npm-registry https://some.io --timeout 2222",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"to", "--debug-options", "--npm-registry", "https://some.io", "--timeout", "2222"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://some.io",
	"reporter": [],
	"select": "",
	"timeout": 2222
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn in --help
		{
			name: "lstn in --help",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"in", "--help"},
			stdout: heredoc.Doc(`Query listen.dev for the verdicts of all the dependencies in your project.

Using this command, you can audit all the dependencies configured for a project and obtain their verdicts.
This requires a package.json file to fetch the package name and version of the project dependencies.

The verdicts it returns are listed by the name of each package and its specified version.

Usage:
  lstn in [path]

Examples:
  lstn in
  lstn in .
  lstn in /we/snitch
  lstn in sub/dir

Flags:
      --genlock             whether to generate the lock file on the fly or not
      --json                output the verdicts (if any) in JSON form
  -l, --lockfiles strings   set one or more lock file paths (relative to the working dir) to lookup for (default [package-lock.json,poetry.lock])

Config Flags:
      --loglevel string        set the logging level (default "info")
      --npm-endpoint string    the listen.dev endpoint emitting the NPM verdicts (default "https://npm.listen.dev")
      --pypi-endpoint string   the listen.dev endpoint emitting the PyPi verdicts (default "https://pypi.listen.dev")
      --timeout int            set the timeout, in seconds (default 60)

Debug Flags:
      --debug-options   output the options, then exit

Filtering Flags:
  -q, --jq string   filter the output verdicts using a jq expression (requires --json)

Registry Flags:
      --npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")

Reporting Flags:
      --gh-owner string                                               set the GitHub owner name (org|user)
      --gh-pull-id int                                                set the GitHub pull request ID
      --gh-repo string                                                set the GitHub repository name
  -r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review,pro)   set one or more reporters to use (default [])

Token Flags:
      --gh-token string    set the GitHub token
      --jwt-token string   set the listen.dev auth token

Global Flags:
      --config string   config file (default is $HOME/.lstn.yaml)
`),
			stderr: "",
			errstr: "",
		},
		// lstn in --debug-options
		{
			name: "lstn in --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"in", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"genlock": false,
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"lockfiles": [
		"package-lock.json",
		"poetry.lock"
	],
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_TIMEOUT=9999 lstn in --debug-options --timeout 8888
		{
			name: "LSTN_TIMEOUT=9999 lstn in --debug-options --timeout 8888",
			envvar: map[string]string{
				"LSTN_TIMEOUT": "9999",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"in", "--debug-options", "--timeout", "8888"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"genlock": false,
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"lockfiles": [
		"package-lock.json",
		"poetry.lock"
	],
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 8888
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_NPM_ENDPOINT=https://npm-staging.listen.dev lstn in --debug-options
		{
			name: "LSTN_NPM_ENDPOINT=https://npm-staging.listen.dev lstn in --debug-options",
			envvar: map[string]string{
				"LSTN_NPM_ENDPOINT": "https://npm-staging.listen.dev/",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"in", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm-staging.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"genlock": false,
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"lockfiles": [
		"package-lock.json",
		"poetry.lock"
	],
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_REPORTER=gh-pull-check lstn scan --debug-options
		{
			name: "LSTN_REPORTER=gh-pull-check lstn scan --debug-options",
			envvar: map[string]string{
				"LSTN_REPORTER": "gh-pull-check",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		44
	],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan --debug-options --gh-token xxx -r gh-pull-review,gh-pull-comment --gh-owner leodido --gh-repo go-urn --gh-pull-id 111
		{
			name: "lstn scan --debug-options --gh-token xxx -r gh-pull-review,gh-pull-comment --gh-owner leodido --gh-repo go-urn --gh-pull-id 111",
			cmdline: []string{
				"scan",
				"--debug-options",
				"--gh-token",
				"xxx",
				"-r",
				"gh-pull-review,gh-pull-comment",
				"--gh-owner",
				"leodido",
				"--gh-repo",
				"go-urn",
				"--gh-pull-id",
				"111",
			},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "leodido",
	"gh-pull-id": 111,
	"gh-repo": "go-urn",
	"gh-token": "xxx",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		33,
		22
	],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_REPORTER=gh-pull-check lstn scan --debug-options --gh-token xxx -r gh-pull-review --gh-owner leodido --gh-repo go-urn --gh-pull-id 111
		// FIXME: for coherence flags MUST always override so -r here should override not merge
		// FIXME: we would like to have `"reporter": [33]` here
		// FIXME: the problem is that the enum setter always merges after the first set invocation
		{
			name: "LSTN_REPORTER=gh-pull-check lstn scan --debug-options --gh-token xxx -r gh-pull-review --gh-owner leodido --gh-repo go-urn --gh-pull-id 111",
			envvar: map[string]string{
				"LSTN_REPORTER": "gh-pull-check",
			},
			cmdline: []string{
				"scan",
				"--debug-options",
				"--gh-token",
				"xxx",
				"-r",
				"gh-pull-review",
				"--gh-owner",
				"leodido",
				"--gh-repo",
				"go-urn",
				"--gh-pull-id",
				"111",
			},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "leodido",
	"gh-pull-id": 111,
	"gh-repo": "go-urn",
	"gh-token": "xxx",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		33,
		44
	],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_GH_OWNER=fntlnz LSTN_GH_PULL_ID=654 LSTN_GH_TOKEN=yyy LSTN_NPM_REGISTRY=https://registry.npmjs.com lstn scan --debug-options
		{
			name: "LSTN_GH_OWNER=fntlnz LSTN_GH_PULL_ID=654 LSTN_GH_TOKEN=yyy LSTN_NPM_REGISTRY=https://registry.npmjs.com lstn scan --debug-options",
			envvar: map[string]string{
				"LSTN_GH_OWNER":     "fntlnz",
				"LSTN_GH_PULL_ID":   "654",
				"LSTN_GH_TOKEN":     "yyy",
				"LSTN_NPM_REGISTRY": "https://registry.npmjs.com",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "fntlnz",
	"gh-pull-id": 654,
	"gh-repo": "",
	"gh-token": "yyy",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.com",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// NOTE: the API for lstn scan doesn't support the JWT auth yet
		// 		// LSTN_JWT_TOKEN=some123jwt.aaa.zzz lstn scan --debug-options
		// 		{
		// 			name: "LSTN_JWT_TOKEN=some123jwt.aaa.zzz lstn scan --debug-options",
		// 			envvar: map[string]string{
		// 				"LSTN_JWT_TOKEN": "some123jwt.aaa.zzz",
		// 				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
		// 				"GITHUB_ACTIONS": "",
		// 			},
		// 			cmdline: []string{"scan", "--debug-options"},
		// 			stdout: heredoc.Doc(`{
		// 	"debug-options": true,
		// 	"endpoint": {
		// 		"npm": "https://npm.listen.dev",
		// 		"pypi": "https://pypi.listen.dev"
		// },
		// 	"gh-owner": "",
		// 	"gh-pull-id": 0,
		// 	"gh-repo": "",
		// 	"gh-token": "",
		// 	"ignore-deptypes": [
		// 		110
		// 	],
		// 	"ignore-packages": null,
		// 	"jq": "",
		// 	"json": false,
		// 	"jwt-token": "some123jwt.aaa.zzz",
		// 	"loglevel": "info",
		// 	"npm-registry": "https://registry.npmjs.org",
		// 	"reporter": [],
		// 	"select": "",
		// 	"timeout": 60
		// }
		// `),
		// 			stderr: "Running without a configuration file\n",
		// 			errstr: "",
		// 		},
		// LSTN_JWT_TOKEN=some123jwt.aaa.xxx lstn in --debug-options
		{
			name: "LSTN_JWT_TOKEN=some123jwt.aaa.xxx lstn in --debug-options",
			envvar: map[string]string{
				"LSTN_JWT_TOKEN": "some123jwt.aaa.xxx",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"in", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"genlock": false,
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "some123jwt.aaa.xxx",
	"lockfiles": [
		"package-lock.json",
		"poetry.lock"
	],
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan --debug-options --config testdata/config_reporting.yaml
		{
			name: "lstn scan --debug-options --config testdata/config_reporting.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_reporting.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_reporting.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "leodido",
	"gh-pull-id": 78999,
	"gh-repo": "go-urn",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://some.io",
	"reporter": [
		33
	],
	"select": "",
	"timeout": 2222
}
`),
			stderr: "",
			errstr: "",
		},
		// LSTN_GH_PULL_ID=887755 LSTN_GH_REPO=go-conventionalcommits LSTN_NPM_ENDPOINT=https://npm-stage.listen.dev LSTN_TIMEOUT=33331 lstn scan --debug-options --config testdata/config_reporting.yaml
		{
			name: "LSTN_GH_PULL_ID=887755 LSTN_GH_REPO=go-conventionalcommits LSTN_NPM_ENDPOINT=https://npm-stage.listen.dev LSTN_TIMEOUT=33331 lstn scan --debug-options --config testdata/config_reporting.yaml",
			envvar: map[string]string{
				"LSTN_GH_PULL_ID":   "887755",
				"LSTN_GH_REPO":      "go-conventionalcommits",
				"LSTN_NPM_ENDPOINT": "https://npm-stage.listen.dev",
				"LSTN_TIMEOUT":      "33331",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_reporting.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_reporting.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm-stage.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "leodido",
	"gh-pull-id": 887755,
	"gh-repo": "go-conventionalcommits",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://some.io",
	"reporter": [
		33
	],
	"select": "",
	"timeout": 33331
}
`),
			stderr: "",
			errstr: "",
		},
		// LSTN_REPORTER=gh-pull-check LSTN_GH_PULL_ID=887755 LSTN_GH_REPO=go-conventionalcommits LSTN_NPM_ENDPOINT=https://npm-stage.listen.dev LSTN_TIMEOUT=33331 lstn scan --debug-options --config testdata/config_reporting.yaml
		{
			name: "LSTN_REPORTER=gh-pull-check LSTN_GH_PULL_ID=887755 LSTN_GH_REPO=go-conventionalcommits LSTN_NPM_ENDPOINT=https://npm-stage.listen.dev LSTN_TIMEOUT=33331 lstn scan --debug-options --config testdata/config_reporting.yaml",
			envvar: map[string]string{
				"LSTN_GH_PULL_ID":   "887755",
				"LSTN_GH_REPO":      "go-conventionalcommits",
				"LSTN_NPM_ENDPOINT": "https://npm-stage.listen.dev",
				"LSTN_TIMEOUT":      "33331",
				"LSTN_REPORTER":     "gh-pull-check",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_reporting.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_reporting.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm-stage.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "leodido",
	"gh-pull-id": 887755,
	"gh-repo": "go-conventionalcommits",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://some.io",
	"reporter": [
		44
	],
	"select": "",
	"timeout": 33331
}
`),
			stderr: "",
			errstr: "",
		},
		// lstn scan --debug-options --reporter gh-pull-comment,gh-pull-comment -r gh-pull-check,gh-pull-comment
		{
			name: "lstn scan --debug-options --reporter gh-pull-comment,gh-pull-comment -r gh-pull-check,gh-pull-comment",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--reporter", "gh-pull-comment,gh-pull-comment", "-r", "gh-pull-check,gh-pull-comment"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		22,
		44
	],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan --debug-options --reporter pro
		{
			name: "lstn scan --debug-options --reporter pro",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--reporter", "pro"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		55
	],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// FIXME: COMPLETE ME
		// lstn scan --debug-options --ignore-deptypes dev,dev --ignore-deptypes optional,dev
		{
			name: "lstn scan --debug-options --ignore-deptypes dev,dev --ignore-deptypes optional,dev",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--ignore-deptypes", "dev,dev", "--ignore-deptypes", "optional,dev"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110,
		66,
		132
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn version --debug-options
		{
			name:    "lstn version --debug-options",
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
		// lstn version --debug-options -v
		{
			name:    "lstn version --debug-options -v",
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
		// lstn version --debug-options -vv
		{
			name:    "lstn version --debug-options -vv",
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
		// lstn version --debug-options -v=2
		{
			name:    "lstn version --debug-options -v=2",
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
		// GITHUB_ACTIONS=true GITHUB_EVENT_PATH=../pkg/ci/testdata/github_event_pul_request.json lstn scan --debug-options
		{
			name: "GITHUB_ACTIONS=true GITHUB_EVENT_PATH=../pkg/ci/testdata/github_event_pul_request.json lstn scan --debug-options",
			envvar: map[string]string{
				"GITHUB_ACTIONS":    "true",
				"GITHUB_EVENT_PATH": path.Join(cwd, "../pkg/ci/testdata/github_event_pull_request.json"),
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "reviewdog",
	"gh-pull-id": 285,
	"gh-repo": "reviewdog",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=overriddenbyflag lstn scan --ignore-packages @vue/devtools,anotherpackage --debug-options
		{
			name: "LSTN_IGNORE_PACKAGES=overriddenbyflag lstn scan --ignore-packages @vue/devtools,anotherpackage --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "overriddenbyflag",
			},
			cmdline: []string{"scan", "--debug-options", "--ignore-packages", "@vue/devtools,anotherpackage"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": [
		"@vue/devtools",
		"anotherpackage"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=overriddenbyflag lstn scan --ignore-packages @vue/devtools --ignore-packages anotherpackage --debug-options
		{
			name: "LSTN_IGNORE_PACKAGES=overriddenbyflag lstn scan --ignore-packages @vue/devtools --ignore-packages anotherpackage --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "overriddenbyflag",
			},
			cmdline: []string{"scan", "--debug-options", "--ignore-packages", "@vue/devtools", "--ignore-packages", "anotherpackage"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": [
		"@vue/devtools",
		"anotherpackage"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=overriddenbyflag lstn scan --ignore-packages @vue/devtools --debug-options
		{
			name: "LSTN_IGNORE_PACKAGES=overriddenbyflag lstn scan --ignore-packages @vue/devtools --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "overriddenbyflag",
			},
			cmdline: []string{"scan", "--debug-options", "--ignore-packages", "@vue/devtools"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": [
		"@vue/devtools"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=@vue/devtools lstn scan --debug-options
		{
			name: "LSTN_IGNORE_PACKAGES=@vue/devtools lstn scan --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "@vue/devtools",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": [
		"@vue/devtools"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=@vue/devtools,@vue/devtools lstn scan --debug-options
		{
			name: "LSTN_IGNORE_PACKAGES=@vue/devtools,@vue/devtools lstn scan --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "@vue/devtools,@vue/devtools",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": [
		"@vue/devtools"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=@vue/devtools,anotherpackage --debug-options
		{
			name: "LSTN_IGNORE_PACKAGES=@vue/devtools,anotherpackage lstn scan --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "@vue/devtools,anotherpackage",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": [
		"@vue/devtools",
		"anotherpackage"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan --debug-options --ignore-packages aaaaa --config testdata/config_filtering.yaml
		{
			name: "lstn scan --debug-options --ignore-packages aaaaa --config testdata/config_filtering.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--ignore-packages", "aaaaa", "--config", path.Join(cwd, "testdata", "config_filtering.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_filtering.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110,
		88
	],
	"ignore-packages": [
		"aaaaa"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": [],
	"select": "",
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// lstn scan --debug-options --config testdata/config_filtering.yaml
		{
			name: "lstn scan --debug-options --config testdata/config_filtering.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_filtering.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_filtering.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110,
		88
	],
	"ignore-packages": [
		"donotprocessme",
		"metoo"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": [],
	"select": "",
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=vvv lstn scan --debug-options --config testdata/config_filtering.yaml
		{
			name: "LSTN_IGNORE_PACKAGES=overrideThoseFromConfig lstn scan --debug-options --config testdata/config_filtering.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "overrideThoseFromConfig",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_filtering.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_filtering.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110,
		88
	],
	"ignore-packages": [
		"overrideThoseFromConfig"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": [],
	"select": "",
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// LSTN_REPORTER=wrong lstn scan --debug-options
		{
			name: "LSTN_REPORTER=wrong lstn scan --debug-options",
			envvar: map[string]string{
				"LSTN_REPORTER": "wrong",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout:  "",
			stderr:  "Running without a configuration file\nError: reporter must be 'gh-pull-check', 'gh-pull-comment', 'gh-pull-review', 'pro'; got wrong\n",
			errstr:  "reporter must be 'gh-pull-check', 'gh-pull-comment', 'gh-pull-review', 'pro'; got wrong",
		},
		// lstn scan --ignore-deptypes dev,peer,dev --debug-options
		{
			name: "lstn scan --ignore-deptypes dev,peer,dev --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--ignore-deptypes", "dev,peer,dev", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110,
		66,
		88
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_DEPTYPES=dev,peer lstn scan --debug-options
		{
			name: "LSTN_IGNORE_DEPTYPES=dev,peer lstn scan --debug-options",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_DEPTYPES": "dev,peer",
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110,
		66,
		88
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// FIXME: the flag does not override the env var
		// FIXME: same issue as for other enum flags (their setters merge in)
		// LSTN_IGNORE_DEPTYPES=dev,peer lstn scan --debug-options --ignore-deptypes optional
		{
			name: "LSTN_IGNORE_DEPTYPES=dev,peer,dev lstn scan --debug-options --ignore-deptypes optional",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_DEPTYPES": "dev,peer,dev",
			},
			cmdline: []string{"scan", "--debug-options", "--ignore-deptypes", "optional"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110,
		132,
		66,
		88
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan --debug-options --config testdata/config_filtering.yaml
		{
			name: "lstn scan --debug-options --config testdata/config_filtering.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_filtering.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_filtering.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110,
		88
	],
	"ignore-packages": [
		"donotprocessme",
		"metoo"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": [],
	"select": "",
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// LSTN_IGNORE_DEPTYPES=dev lstn scan --debug-options --config testdata/config_filtering.yaml
		{
			name: "LSTN_IGNORE_DEPTYPES=dev lstn scan --debug-options --config testdata/config_filtering.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_DEPTYPES": "dev",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_filtering.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_filtering.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110,
		66
	],
	"ignore-packages": [
		"donotprocessme",
		"metoo"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": [],
	"select": "",
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=overrideThoseFromConfig LSTN_IGNORE_DEPTYPES=dev lstn scan --debug-options --config testdata/config_filtering.yaml
		{
			name: "LSTN_IGNORE_PACKAGES=overrideThoseFromConfig LSTN_IGNORE_DEPTYPES=dev lstn scan --debug-options --config testdata/config_filtering.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_DEPTYPES": "dev",
				"LSTN_IGNORE_PACKAGES": "overrideThoseFromConfig",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_filtering.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_filtering.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110,
		66
	],
	"ignore-packages": [
		"overrideThoseFromConfig"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": [],
	"select": "",
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// FIXME: the flag does not override the config value
		// FIXME: same issue as for other enum flags (their setters merge in)
		// lstn scan --debug-options --config testdata/config_filtering.yaml --ignore-deptypes dev,optional
		{
			name: "lstn scan --debug-options --config testdata/config_filtering.yaml --ignore-deptypes dev,optional",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_filtering.yaml"), "--ignore-deptypes", "dev,optional"},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_filtering.yaml
{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-deptypes": [
		110,
		66,
		132,
		88
	],
	"ignore-packages": [
		"donotprocessme",
		"metoo"
	],
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": [],
	"select": "",
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// lstn scan --debug-options --select '@.severity == "high"'
		{
			name: `lstn scan --debug-options --select '@.severity == "high"'`,
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--select", `@.severity == "high"`},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "@.severity == \"high\"",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_SELECT='"network" in @.categories' lstn scan --debug-options
		{
			name: `LSTN_SELECT='"network" in @.categories' lstn scan --debug-options`,
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
				"LSTN_SELECT":    `"network" in @.categories`,
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "\"network\" in @.categories",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn to --debug-options -s '(@.file !~ "^advisory" && @.message != "")'
		// TODO: double-check why the debug options JSON is returning unicode chars in place of the `&` ones
		{
			name: `lstn to --debug-options -s '(@.file !~ "^advisory" && @.message != "")'`,
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"to", "--debug-options", "-s", `(@.file !~ "^advisory" && @.message != "")`},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": {
		"npm": "https://npm.listen.dev",
		"pypi": "https://pypi.listen.dev"
	},
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-deptypes": [
		110
	],
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"jwt-token": "",
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [],
	"select": "(@.file !~ \"^advisory\" \u0026\u0026 @.message != \"\")",
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			closer := internaltesting.EnvSetter(tc.envvar)
			t.Cleanup(closer)

			rootC, e := root.New(context.Background())
			assert.Nil(t, e)
			c := rootC.Command()

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
