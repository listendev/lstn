package cmd

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/spf13/cobra"
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
	commands commandsMap
}

func (suite *CmdHelpSuite) SetupSuite() {
	Boot(&BootOptions{run: false})

	suite.commands = make(map[commandName]*cobra.Command)
	suite.commands[Root] = rootCmd

	for _, command := range rootCmd.Commands() {
		suite.commands[commandName(command.Name())] = command
	}

	fmt.Println(suite.commands)
}

func TestCmdSuites(t *testing.T) {
	suite.Run(t, new(CmdHelpSuite))
}

func (suite *CmdHelpSuite) TestTopics() {
}

// Utils

type commandsMap map[commandName]*cobra.Command

func (m commandsMap) String() string {
	res := ""

	for name, cmd := range m {
		res += fmt.Sprintf("%-12s: %p\n", name, cmd)
	}

	return res
}

func execute(t *testing.T, c *cobra.Command, args ...string) (string, string) {
	t.Helper()

	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	c.SetOut(stdout)
	c.SetErr(stderr)
	c.SetArgs(args)
	c.Execute()
	o, err := io.ReadAll(stdout)
	if err != nil {
		panic(err)
	}
	e, err := io.ReadAll(stderr)
	if err != nil {
		panic(err)
	}
	return string(o), string(e)
}
