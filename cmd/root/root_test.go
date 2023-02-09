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
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CmdAdditionalHelpSuite struct {
	suite.Suite
	commands     commandsMap
	expectedOuts expectedOutsMap
}

func (suite *CmdAdditionalHelpSuite) SetupSuite() {
	rootCmd, err := New(context.Background())
	if err != nil {
		suite.Fail("couldn't instantiate the root command")
	}

	suite.commands = make(commandsMap)
	suite.commands[Root] = rootCmd.cmd

	for _, command := range rootCmd.cmd.Commands() {
		suite.commands[commandName(command.Name())] = command
	}

	suite.expectedOuts = make(expectedOutsMap)
	b := "# lstn configuration file\n\n"
	b += "The `lstn` CLI looks for a configuration file .lstn.yaml in your `$HOME` directory when it starts.\n\n"
	b += "In this file you can set the values for the global `lstn` configurations.\n"
	b += "Anyways, notice that environment variables, and flags (if any) override the values in your configuration file.\n\n"
	b += "Here's an example of a configuration file (with the default values):\n\n"
	b += "```yaml\nendpoint: http://127.0.0.1:3000\nloglevel: info\ntimeout: 60\n```\n"
	suite.expectedOuts[Config] = b

	b = "# lstn environment variables\n\n"
	b += "The environment variables override any corresponding configuration setting.\n\n"
	b += "But flags override them.\n\n"
	b += "`LSTN_ENDPOINT`: the listen.dev endpoint emitting the verdicts\n\n"
	b += "`LSTN_LOGLEVEL`: log level\n\n"
	b += "`LSTN_TIMEOUT`: timeout in seconds\n\n"
	suite.expectedOuts[Environment] = b

	suite.expectedOuts[Manual] = "# lstn cheatsheet\n\n## Global Flags\n\nEvery child command inherits the following flags:\n\n```\n--config string     config file (default is $HOME/.lstn.yaml)\n--endpoint string   the listen.dev endpoint emitting the verdicts (default \"http://127.0.0.1:3000\")\n--loglevel string   log level (default \"info\")\n--timeout int       timeout in seconds (default 60)\n```\n\n## `lstn completion`\n\nGenerate the autocompletion script for the specified shell.\n\n### `lstn completion bash`\n\nGenerate the autocompletion script for bash.\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n### `lstn completion fish [flags]`\n\nGenerate the autocompletion script for fish.\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n### `lstn completion powershell [flags]`\n\nGenerate the autocompletion script for powershell.\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n### `lstn completion zsh [flags]`\n\nGenerate the autocompletion script for zsh.\n\n```\n--no-descriptions   disable completion descriptions\n```\n\n## `lstn config`\n\nDetails about the ~/.lstn.yaml config file.\n\n```\n-h, --help   help for config\n```\n\n## `lstn environment`\n\nWhich environment variables you can use with lstn.\n\n```\n-h, --help   help for environment\n```\n\n## `lstn exit`\n\nDetails about the lstn exit codes.\n\n## `lstn help [command]`\n\nHelp about any command.\n\n## `lstn in <path>`\n\nInspect the verdicts of your dependencies.\n\n```\n-q, --jq string   filter the output using a jq expression\n    --json        output the verdicts (if any) in JSON form\n```\n\nFor example:\n\n```bash\nlstn in\nlstn in .\nlstn in /we/snitch\nlstn in sub/dir\n```\n\n## `lstn manual`\n\nA comprehensive reference of all the lstn commands.\n\n```\n-h, --help   help for manual\n```\n\n## `lstn to <name> [version] [shasum]`\n\nGet the verdicts of a package.\n\n```\n-q, --jq string   filter the output using a jq expression\n    --json        output the verdicts (if any) in JSON form\n```\n\nFor example:\n\n```bash\nlstn to chalk\nlstn to debug 4.3.4\n```\n\n## `lstn version`\n\nPrint out version information.\n\n"

	b = "The lstn CLI follows the usual conventions regarding exit codes.\n\n"
	b += "Meaning:\n\n"
	b += "* when a command completes successfully, the exit code will be 0\n\n"
	b += "* when a command fails for any reason, the exit code will be 1\n\n"
	b += "* when a command is running but gets cancelled, the exit code will be 2\n\n"
	b += "* when a command meets an authentication issue, the exit code will be 4\n\n"
	b += "Notice that it's possible that a particular command may have more exit codes,\n"
	b += "so it's a good practice to check the docs for the specific command\n"
	b += "in case you're relying on the exit codes to control some behaviour.\n"
	suite.expectedOuts[Exit] = b
}

func TestCmdSuites(t *testing.T) {
	suite.Run(t, new(CmdAdditionalHelpSuite))
}

func (suite *CmdAdditionalHelpSuite) TestTopics() {
	topics := []commandName{Config, Environment, Manual, Exit}
	for _, topic := range topics {
		out := execute(suite.T(), suite.commands[Root], topic)
		require.Equal(suite.T(), suite.expectedOuts[topic], out)
	}
}

type commandName string

const (
	Root        commandName = "lstn"
	Config      commandName = "config"
	Environment             = "environment"
	Exit                    = "exit"
	In          commandName = "in"
	Manual                  = "manual"
	To                      = "to"
	Version                 = "version"
)

type commandsMap map[commandName]*cobra.Command
type expectedOutsMap map[commandName]string

func (m commandsMap) String() string {
	res := ""

	for name, cmd := range m {
		res += fmt.Sprintf("%-12s: %p\n", name, cmd)
	}

	return res
}

func execute(t *testing.T, c *cobra.Command, sub commandName) string {
	t.Helper()

	b := bytes.NewBufferString("")
	c.SetOut(b)
	c.SetArgs([]string{string(sub)})
	err := c.Execute()
	if err != nil {
		t.Fatalf("Error during execution of %s %s", c.Name(), sub)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("Error during reading of %s %s", c.Name(), sub)
	}

	return string(out)
}
