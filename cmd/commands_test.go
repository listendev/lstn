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
			stderr:  "Running without a configuration file\nError: directory _CWD_ does not contain a package.json file\n",
			errstr:  "directory _CWD_ does not contain a package.json file",
		},
		// lstn to
		{
			name:    "lstn to",
			cmdline: []string{"to"},
			stdout:  "",
			stderr:  "Error: requires at least 1 arg (package name)\n",
			errstr:  "requires at least 1 arg (package name)",
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
	"endpoint": "https://npm.listen.dev",
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
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
	"endpoint": "https://npm.listen.dev",
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://some.io",
	"reporter": null,
	"timeout": 2222
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn in --help
		{
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			name:    "lstn in --help",
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

Filtering Flags:
      --ignore-packages strings   list of packages to not process

Registry Flags:
      --npm-registry string   set a custom NPM registry (default "https://registry.npmjs.org")

Reporting Flags:
      --gh-owner string                                           set the GitHub owner name (org|user)
      --gh-pull-id int                                            set the GitHub pull request ID
      --gh-repo string                                            set the GitHub repository name
  -r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review)   set one or more reporters to use (default [])

Token Flags:
      --gh-token string   set the GitHub token

Global Flags:
      --config string   config file (default is $HOME/.lstn.yaml)
`),
			stderr: "",
			errstr: "",
		},
		// lstn in --debug-options
		{
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			name:    "lstn in --debug-options",
			cmdline: []string{"in", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
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
	"endpoint": "https://npm.listen.dev",
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
	"timeout": 8888
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_ENDPOINT=https://npm-staging.listen.dev lstn in --debug-options
		{
			name: "LSTN_ENDPOINT=https://npm-staging.listen.dev lstn in --debug-options",
			envvar: map[string]string{
				"LSTN_ENDPOINT": "https://npm-staging.listen.dev/",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"in", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm-staging.listen.dev",
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan -e dev,peer --debug-options
		{
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			name:    "lstn scan -e dev,peer --debug-options",
			cmdline: []string{"scan", "-e", "dev,peer", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110,
		66,
		88
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		44
	],
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "leodido",
	"gh-pull-id": 111,
	"gh-repo": "go-urn",
	"gh-token": "xxx",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		33,
		22
	],
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_REPORTER=gh-pull-check lstn scan --debug-options --gh-token xxx -r gh-pull-review --gh-owner leodido --gh-repo go-urn --gh-pull-id 111
		// FIXME: for coherence flags MUST always override so -r here should override not merge // "reporter": [33]
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "leodido",
	"gh-pull-id": 111,
	"gh-repo": "go-urn",
	"gh-token": "xxx",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": [
		33,
		44
	],
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "fntlnz",
	"gh-pull-id": 654,
	"gh-repo": "",
	"gh-token": "yyy",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.com",
	"reporter": null,
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan --debug-options --config testdata/config_with_reporters.yaml
		{
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			name:    "lstn scan --debug-options --config testdata/config_with_reporters.yaml",
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_with_reporters.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_with_reporters.yaml
{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "leodido",
	"gh-pull-id": 78999,
	"gh-repo": "go-urn",
	"gh-token": "zzz",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://some.io",
	"reporter": [
		33
	],
	"timeout": 2222
}
`),
			stderr: "",
			errstr: "",
		},
		// FIXME: Add LSTN_REPORTER=gh-pull-check and see if it merges with the one in the configuration file
		// LSTN_GH_PULL_ID=887755 LSTN_GH_REPO=go-conventionalcommits LSTN_ENDPOINT=https://npm-stage.listen.dev LSTN_TIMEOUT=33331 lstn scan --debug-options --config testdata/config_with_reporters.yaml
		{
			name: "LSTN_GH_PULL_ID=887755 LSTN_GH_REPO=go-conventionalcommits LSTN_ENDPOINT=https://npm-stage.listen.dev LSTN_TIMEOUT=33331 lstn scan --debug-options --config testdata/config_with_reporters.yaml",
			envvar: map[string]string{
				"LSTN_GH_PULL_ID": "887755",
				"LSTN_GH_REPO":    "go-conventionalcommits",
				"LSTN_ENDPOINT":   "https://npm-stage.listen.dev",
				"LSTN_TIMEOUT":    "33331",
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_with_reporters.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_with_reporters.yaml
{
	"debug-options": true,
	"endpoint": "https://npm-stage.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "leodido",
	"gh-pull-id": 887755,
	"gh-repo": "go-conventionalcommits",
	"gh-token": "zzz",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://some.io",
	"reporter": [
		33
	],
	"timeout": 33331
}
`),
			stderr: "",
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
		// GITHUB_ACTIONS=true GITHUB_EVENT_PATH=../pkg/.../github_event_pul_request.json lstn scan --debug-options
		{
			name: "GITHUB_ACTIONS=true GITHUB_EVENT_PATH=../pkg/ci/testdata/github_event_pul_request.json lstn scan --debug-options",
			envvar: map[string]string{
				"GITHUB_ACTIONS":    "true",
				"GITHUB_EVENT_PATH": path.Join(cwd, "../pkg/ci/testdata/github_event_pull_request.json"),
			},
			cmdline: []string{"scan", "--debug-options"},
			stdout: heredoc.Doc(`{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "reviewdog",
	"gh-pull-id": 285,
	"gh-repo": "reviewdog",
	"gh-token": "",
	"ignore-packages": null,
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=overriddenbyflag lstn scan --ignore-packages @vue/devtools,anotherpackages --debug-options
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": [
		"@vue/devtools",
		"anotherpackage"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": [
		"@vue/devtools",
		"anotherpackage"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": [
		"@vue/devtools"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=@vue/devtools --debug-options
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": [
		"@vue/devtools"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
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
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "",
	"ignore-packages": [
		"@vue/devtools",
		"anotherpackage"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://registry.npmjs.org",
	"reporter": null,
	"timeout": 60
}
`),
			stderr: "Running without a configuration file\n",
			errstr: "",
		},
		// lstn scan --debug-options --ignore-packages aaaaa --config testdata/config_with_ignores.yaml
		{
			name: "lstn scan --debug-options --ignore-packages aaaaa --config testdata/config_with_ignores.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--ignore-packages", "aaaaa", "--config", path.Join(cwd, "testdata", "config_with_ignores.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_with_ignores.yaml
{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-packages": [
		"aaaaa"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": null,
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// lstn scan --debug-options --config testdata/config_with_ignores.yaml
		{
			name: "lstn scan --debug-options --config testdata/config_with_ignores.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS": "",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_with_ignores.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_with_ignores.yaml
{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-packages": [
		"donotprocessme",
		"metoo"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": null,
	"timeout": 1111
}
`),
			stderr: "",
			errstr: "",
		},
		// LSTN_IGNORE_PACKAGES=vvv lstn scan --debug-options --config testdata/config_with_ignores.yaml
		{
			name: "lstn scan --debug-options --config testdata/config_with_ignores.yaml",
			envvar: map[string]string{
				// Temporarily pretend not to be in a GitHub Action (to make test work in a GitHub Action workflow)
				"GITHUB_ACTIONS":       "",
				"LSTN_IGNORE_PACKAGES": "overrideThoseFromConfig",
			},
			cmdline: []string{"scan", "--debug-options", "--config", path.Join(cwd, "testdata", "config_with_ignores.yaml")},
			stdout: heredoc.Doc(`Using config file: _CWD_/testdata/config_with_ignores.yaml
{
	"debug-options": true,
	"endpoint": "https://npm.listen.dev",
	"exclude": [
		110
	],
	"gh-owner": "",
	"gh-pull-id": 0,
	"gh-repo": "",
	"gh-token": "zzz",
	"ignore-packages": [
		"overrideThoseFromConfig"
	],
	"jq": "",
	"json": false,
	"loglevel": "info",
	"npm-registry": "https://smtg.io",
	"reporter": null,
	"timeout": 1111
}
`),
			stderr: "",
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

// FIXME: remove me
// func TestChildCommands2(t *testing.T) {
// 	cwd, _ := os.Getwd()

// 	cases := []testCase{}

// 	for _, tc := range cases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			closer := internaltesting.EnvSetter(tc.envvar)
// 			t.Cleanup(closer)

// 			rootC, e := root.New(context.Background())
// 			assert.Nil(t, e)
// 			c := rootC.Command()

// 			stdOut, stdErr, err := internaltesting.ExecuteCommand(c, tc.cmdline...)

// 			tc.stdout = strings.ReplaceAll(tc.stdout, "_CWD_", cwd)
// 			tc.stderr = strings.ReplaceAll(tc.stderr, "_CWD_", cwd)

// 			if tc.errstr != "" {
// 				if assert.Error(t, err) {
// 					tc.errstr = strings.ReplaceAll(tc.errstr, "_CWD_", cwd)
// 					assert.Equal(t, tc.errstr, err.Error())
// 				}
// 			}
// 			assert.Equal(t, tc.stdout, stdOut)
// 			assert.Equal(t, tc.stderr, stdErr)
// 		})
// 	}
// }
