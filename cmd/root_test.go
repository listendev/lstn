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
	b += fmt.Sprintf("%s\n\n", "The `lstn` CLI looks for a configuration file .lstn.yaml in your `$HOME` directory when it starts.")
	b += fmt.Sprintf("%s\n", "In this file you can set the values for the global `lstn` configurations.")
	b += fmt.Sprintf("%s\n\n", "Anyways, notice that environment variables, and flags (if any) override the values in your configuration file.")
	b += fmt.Sprintf("%s\n\n", "Here's an example of a configuration file (with the default values):")
	b += "```yaml\nendpoint: http://127.0.0.1:3000\nloglevel: info\ntimeout: 60\n```\n"
	suite.expectedOuts[Config] = b
}

func TestCmdSuites(t *testing.T) {
	suite.Run(t, new(CmdHelpSuite))
}

func (suite *CmdHelpSuite) TestTopics() {
	topics := []commandName{Config}
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
