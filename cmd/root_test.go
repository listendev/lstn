package cmd

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

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

type CmdHelpSuite struct {
	suite.Suite
	commands     commandsMap
	expectedOuts expectedOutsMap
}

func (suite *CmdHelpSuite) SetupSuite() {
	Boot(&BootOptions{run: false})

	suite.commands = make(commandsMap)
	suite.commands[Root] = rootCmd

	for _, command := range rootCmd.Commands() {
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
}

func TestCmdSuites(t *testing.T) {
	suite.Run(t, new(CmdHelpSuite))
}

func (suite *CmdHelpSuite) TestTopics() {
	topics := []commandName{Config, Environment}
	for _, topic := range topics {
		out := execute(suite.T(), suite.commands[Root], topic)
		require.Equal(suite.T(), suite.expectedOuts[topic], out)
	}
}

// Utils

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
	c.Execute()
	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("Error during reading of %s %s", c.Name(), sub)
	}
	return string(out)
}
