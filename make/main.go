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

// Compile the lstn CLI et al.
//
// Usage: go run make/main.go [<envvars>...] [<lstn|man|clean|tag>...]
//
// Examples:
// - go run make/main.go GOOS=windows lstn

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/listendev/lstn/pkg/validate"
	"golang.org/x/exp/maps"
)

const outputPath = "lstn"

var (
	commands = map[string]func([]string) error{
		outputPath: func(args []string) error {
			exe := args[0]
			fmt.Fprintf(os.Stdout, "executing `%s` to build the %s executable...\n", outputPath, exe)

			info, err := os.Stat(exe)
			if err == nil && !sourceFilesLaterThan(info.ModTime()) {
				fmt.Fprintf(os.Stderr, "nothing to do because `%s` is up to date.\n", exe)

				return nil
			}

			ldflags := os.Getenv("GO_LDFLAGS")
			// Default to `-s -w` ldflags
			if len(ldflags) == 0 {
				ldflags = "-w -s"
			}
			gitTag := getGitTag()
			if gitTag != "" {
				if len(ldflags) > 0 {
					ldflags += " "
				}
				ldflags += fmt.Sprintf("-X github.com/listendev/lstn/pkg/version.VersionPrefix=%s", gitTag)
			}

			return run("go", "build", "-trimpath", "-ldflags", ldflags, "-o", exe, "./cmd/lstn")
		},
		"man": func(args []string) error {
			fmt.Fprintf(os.Stdout, "executing `%s` ...\n", args[0])

			destDir := filepath.Join("share", "man", "man1")
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return fmt.Errorf("couldn't create the destination directory")
			}

			return run("go", "run", "./make/docs", "manpages", "--dest", destDir)
		},
		"clean": func(args []string) error {
			fmt.Fprintf(os.Stdout, "executing `%s` ...\n", args[0])
			// TODO
			return nil
		},
		"tag": func(args []string) error {
			// Check we have a version argument
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "missing the version argument ...\n")
				os.Exit(1)
			}
			// Check the tag argument is semver
			v := strings.TrimPrefix(args[1], "v")
			if err := validate.Singleton.Var(v, "semver"); err != nil {
				fmt.Fprintf(os.Stderr, "%s is not a valid semantic version.\n", v)
				os.Exit(1)
			}
			if !strings.HasPrefix(v, "v") {
				v = fmt.Sprintf("v%s", v)
			}

			fmt.Fprintf(os.Stdout, "executing `%s` with version `%s` ...\n", args[0], v)

			return run("git", "tag", "-a", v, "-m", fmt.Sprintf("Release %s", v), "main")
		},
	}
)

func main() {
	// Check this tool is run from the root directory
	src, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	cwd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	if cwd == filepath.Dir(src) {
		fmt.Fprintln(os.Stderr, "this tool must be run from another directory")
		os.Exit(1)
	}

	// Construct the arguments and the environment variables
	args := os.Args[1:]
	for i, arg := range os.Args[1:] {
		if idx := strings.IndexRune(arg, '='); idx >= 0 {
			// It's an environment variable
			os.Setenv(arg[:idx], arg[idx+1:])
			args = append(args[:i], args[i+1:]...)
		}
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "specify one command in %s.\n", maps.Keys(commands))
		os.Exit(1)
	}

	if len(args) == 1 {
		if isTargetingWindows() {
			args[0] = fmt.Sprintf("%s.exe", args[0])
		}
	}

	norm := filepath.ToSlash(strings.TrimSuffix(args[0], ".exe"))
	c := commands[norm]
	if c == nil {
		fmt.Fprintf(os.Stderr, "unknown command `%s`.\n", norm)
		os.Exit(1)
	}

	if err := c(args); err != nil {
		fmt.Fprintf(os.Stderr, "failure while executing `%s`.\n", norm)
		os.Exit(1)
	}
}

func getGitTag() string {
	if des, err := getCommandOutput("git", "describe", "--tags"); err == nil {
		return des
	}

	return ""
}

func isTargetingWindows() bool {
	if os.Getenv("GOOS") == "windows" {
		return true
	}
	if runtime.GOOS == "windows" {
		return true
	}

	return false
}

func run(args ...string) error {
	exe, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	setX(args...)
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func getCommandOutput(args ...string) (string, error) {
	exe, err := exec.LookPath(args[0])
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stderr = io.Discard
	out, err := cmd.Output()

	return strings.TrimSuffix(string(out), "\n"), err
}

func setX(args ...string) {
	// Escape special chars
	fmtArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " \t'\"") {
			fmtArgs[i] = fmt.Sprintf("%q", arg)
		} else {
			fmtArgs[i] = arg
		}
	}

	fmt.Fprintf(os.Stderr, "+ %s\n", strings.Join(fmtArgs, " "))
}

func isAccessDenied(err error) bool {
	var pe *os.PathError

	return errors.As(err, &pe) && strings.Contains(pe.Err.Error(), "Access is denied")
}

func sourceFilesLaterThan(t time.Time) bool {
	foundLater := false
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore errors that occur when the project contains a symlink to a filesystem or volume that it doesn't have access to
			if path != "." && isAccessDenied(err) {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)

				return nil
			}

			return err
		}
		if foundLater {
			return filepath.SkipDir
		}
		if len(path) > 1 && (path[0] == '.' || path[0] == '_') {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}
		if info.IsDir() {
			if name := filepath.Base(path); name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}

			return nil
		}
		if path == "go.mod" || path == "go.sum" || (strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")) {
			if info.ModTime().After(t) {
				foundLater = true
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return foundLater
}
