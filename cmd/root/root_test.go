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
	"testing"

	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type commandName string
type commandsMap map[commandName]*cobra.Command
type expectedOutsMap map[commandName]string

type CmdAdditionalHelpSuite struct {
	suite.Suite
	rootC        *Command
	commands     commandsMap
	expectedOuts expectedOutsMap
	envCloser    func()
}

func (suite *CmdAdditionalHelpSuite) TearDownSuite() {
	suite.T().Cleanup(suite.envCloser)
}

func (suite *CmdAdditionalHelpSuite) SetupSuite() {
	suite.envCloser = internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "",
		"GITHUB_EVENT_PATH": "",
	})

	rootCmd, err := New(context.Background())
	if err != nil {
		suite.Fail("couldn't instantiate the root command")
	}

	suite.commands = make(commandsMap)
	suite.rootC = rootCmd

	for _, command := range rootCmd.cmd.Commands() {
		suite.commands[commandName(command.Name())] = command
	}

	suite.expectedOuts = make(expectedOutsMap)
	suite.expectedOuts[Config] = "# lstn configuration file\n\nThe `lstn` CLI looks for a configuration file .lstn.yaml in your `$HOME` directory when it starts.\n\nIn this file you can set the values for the global `lstn` configurations.\nAnyways, notice that environment variables, and flags (if any) override the values in your configuration file.\n\nHere's an example of a configuration file (with the default values):\n\n```yaml\nendpoint: \"https://npm.listen.dev\"\nloglevel: \"info\"\nregistry: \n  npm: \"https://registry.npmjs.org\"\nreporter: \n  github: \n    owner: \"...\"\n    pull: \n      id: 0\n    repo: \"...\"\n  types: \n    - \"...\"\n    - \"...\"\ntimeout: 60\ntoken: \n  github: \"...\"\n```\n"

	suite.expectedOuts[Environment] = "# lstn environment variables\n\nThe environment variables override any corresponding configuration setting.\n\nBut flags override them.\n\n`LSTN_ENDPOINT`: the listen.dev endpoint emitting the verdicts\n\n`LSTN_GH_OWNER`: set the GitHub owner name (org|user)\n\n`LSTN_GH_PULL_ID`: set the GitHub pull request ID\n\n`LSTN_GH_REPO`: set the GitHub repository name\n\n`LSTN_GH_TOKEN`: set the GitHub token\n\n`LSTN_LOGLEVEL`: set the logging level\n\n`LSTN_NPM_REGISTRY`: set a custom NPM registry\n\n`LSTN_REPORTER`: set one or more reporters to use\n\n`LSTN_TIMEOUT`: set the timeout, in seconds\n\n"

	suite.expectedOuts[Manual] = "# lstn cheatsheet\n\n## Global Flags\n\nEvery child command inherits the following flags:\n\n```\n--config string   config file (default is $HOME/.lstn.yaml)\n```\n\n## `lstn completion <bash|fish|powershell|zsh>`\n\nGenerate the autocompletion script for the specified shell.\n\n### `lstn completion bash`\n\nGenerate the autocompletion script for bash.\n\n#### Flags\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n### `lstn completion fish [flags]`\n\nGenerate the autocompletion script for fish.\n\n#### Flags\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n### `lstn completion powershell [flags]`\n\nGenerate the autocompletion script for powershell.\n\n#### Flags\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n### `lstn completion zsh [flags]`\n\nGenerate the autocompletion script for zsh.\n\n#### Flags\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n## `lstn config`\n\nDetails about the ~/.lstn.yaml config file.\n\n## `lstn environment`\n\nWhich environment variables you can use with lstn.\n\n## `lstn exit`\n\nDetails about the lstn exit codes.\n\n## `lstn help [command]`\n\nHelp about any command.\n\n## `lstn in <path>`\n\nInspect the verdicts for your dependencies tree.\n\n### Flags\n\n```\n-q, --jq string   filter the output using a jq expression\n    --json        output the verdicts (if any) in JSON form\n```\n\n### Config Flags\n\n```\n--endpoint string   the listen.dev endpoint emitting the verdicts (default \"https://npm.listen.dev\")\n--loglevel string   set the logging level (default \"info\")\n--timeout int       set the timeout, in seconds (default 60)\n```\n\n### Debug Flags\n\n```\n--debug-options   output the options, then exit\n```\n\n### Registry Flags\n\n```\n--npm-registry string   set a custom NPM registry (default \"https://registry.npmjs.org\")\n```\n\n### Reporting Flags\n\n```\n    --gh-owner string                                           set the GitHub owner name (org|user)\n    --gh-pull-id int                                            set the GitHub pull request ID\n    --gh-repo string                                            set the GitHub repository name\n-r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review)   set one or more reporters to use (default [])\n```\n\n### Token Flags\n\n```\n--gh-token string   set the GitHub token\n```\n\nFor example:\n\n```bash\nlstn in\nlstn in .\nlstn in /we/snitch\nlstn in sub/dir\n```\n\n## `lstn manual`\n\nA comprehensive reference of all the lstn commands.\n\n## `lstn reporters`\n\nA comprehensive guide to the `lstn` reporting mechanisms.\n\n## `lstn scan <path>`\n\nInspect the verdicts for your direct dependencies.\n\n### Flags\n\n```\n-e, --exclude (dep,dev,optional,peer)   sets of dependencies to exclude (in addition to the default) (default [bundle])\n-q, --jq string                         filter the output using a jq expression\n    --json                              output the verdicts (if any) in JSON form\n```\n\n### Config Flags\n\n```\n--endpoint string   the listen.dev endpoint emitting the verdicts (default \"https://npm.listen.dev\")\n--loglevel string   set the logging level (default \"info\")\n--timeout int       set the timeout, in seconds (default 60)\n```\n\n### Debug Flags\n\n```\n--debug-options   output the options, then exit\n```\n\n### Registry Flags\n\n```\n--npm-registry string   set a custom NPM registry (default \"https://registry.npmjs.org\")\n```\n\n### Reporting Flags\n\n```\n    --gh-owner string                                           set the GitHub owner name (org|user)\n    --gh-pull-id int                                            set the GitHub pull request ID\n    --gh-repo string                                            set the GitHub repository name\n-r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review)   set one or more reporters to use (default [])\n```\n\n### Token Flags\n\n```\n--gh-token string   set the GitHub token\n```\n\nFor example:\n\n```bash\nlstn scan\nlstn scan .\nlstn scan /we/snitch\nlstn scan /we/snitch -e peer\nlstn scan /we/snitch -e dev,peer\nlstn scan /we/snitch -e dev -e peer\nlstn scan sub/dir\n```\n\n## `lstn to <name> [[version] [shasum] | [version constraint]]`\n\nGet the verdicts of a package.\n\n### Flags\n\n```\n-q, --jq string   filter the output using a jq expression\n    --json        output the verdicts (if any) in JSON form\n```\n\n### Config Flags\n\n```\n--endpoint string   the listen.dev endpoint emitting the verdicts (default \"https://npm.listen.dev\")\n--loglevel string   set the logging level (default \"info\")\n--timeout int       set the timeout, in seconds (default 60)\n```\n\n### Debug Flags\n\n```\n--debug-options   output the options, then exit\n```\n\n### Registry Flags\n\n```\n--npm-registry string   set a custom NPM registry (default \"https://registry.npmjs.org\")\n```\n\n### Reporting Flags\n\n```\n    --gh-owner string                                           set the GitHub owner name (org|user)\n    --gh-pull-id int                                            set the GitHub pull request ID\n    --gh-repo string                                            set the GitHub repository name\n-r, --reporter (gh-pull-check,gh-pull-comment,gh-pull-review)   set one or more reporters to use (default [])\n```\n\n### Token Flags\n\n```\n--gh-token string   set the GitHub token\n```\n\nFor example:\n\n```bash\n# Get the verdicts for all the chalk versions that listen.dev owns\nlstn to chalk\nlstn to debug 4.3.4\nlstn to react 18.0.0 b468736d1f4a5891f38585ba8e8fb29f91c3cb96\n\n# Get the verdicts for all the existing chalk versions\nlstn to chalk \"*\"\n# Get the verdicts for nock versions >= 13.2.0 and < 13.3.0\nlstn to nock \"~13.2.x\"\n# Get the verdicts for tap versions >= 16.3.0 and < 16.4.0\nlstn to tap \"^16.3.0\"\n# Get the verdicts for prettier versions >= 2.7.0 <= 3.0.0\nlstn to prettier \">=2.7.0 <=3.0.0\"\n```\n\n## `lstn version`\n\nPrint out version information.\n\n### Flags\n\n```\n-v, -- count      increment the verbosity level\n    --changelog   output the relase notes URL\n```\n\n### Debug Flags\n\n```\n--debug-options   output the options, then exit\n```\n\n"

	suite.expectedOuts[Exit] = "The lstn CLI follows the usual conventions regarding exit codes.\n\nMeaning:\n\n* when a command completes successfully, the exit code will be 0\n\n* when a command fails for any reason, the exit code will be 1\n\n* when a command is running but gets cancelled, the exit code will be 2\n\n* when a command meets an authentication issue, the exit code will be 4\n\nNotice that it's possible that a particular command may have more exit codes,\nso it's a good practice to check the docs for the specific command\nin case you're relying on the exit codes to control some behaviour.\n"
}

func TestCmdSuites(t *testing.T) {
	suite.Run(t, new(CmdAdditionalHelpSuite))
}

func (suite *CmdAdditionalHelpSuite) TestTopics() {
	topics := []commandName{Manual, Config, Exit, Environment}
	for _, topic := range topics {
		out := execute(suite.T(), suite.rootC, topic)
		require.Equal(suite.T(), suite.expectedOuts[topic], out)
	}
}

const (
	Root        commandName = "lstn"
	Config      commandName = "config"
	Environment commandName = "environment"
	Exit        commandName = "exit"
	Manual      commandName = "manual"
)

func (m commandsMap) String() string {
	res := ""

	for name, cmd := range m {
		res += fmt.Sprintf("%-12s: %p\n", name, cmd)
	}

	return res
}

func execute(t *testing.T, c *Command, sub commandName) string {
	t.Helper()

	stdOut, _, err := internaltesting.ExecuteCommand(c.cmd, string(sub))
	if err != nil {
		t.Fatal(err)
	}

	return stdOut
}
