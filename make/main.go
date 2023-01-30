/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Compile the lstn CLI et al.
//
// Usage: go run make/main.go [<envvars>...] [<lstn|man|clean>...]
//
// Examples:
// - go run make/main.go GOOS=windows lstn

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const outputPath = "lstn"

var (
	commands = map[string]func(string) error{
		outputPath: func(exe string) error {
			fmt.Fprintf(os.Stdout, "executing `%s` to build the %s executable...\n", outputPath, exe)

			info, err := os.Stat(exe)
			if err == nil && !sourceFilesLaterThan(info.ModTime()) {
				fmt.Fprintf(os.Stderr, "nothing to do because `%s` is up to date.\n", exe)
				return nil
			}

			// TODO > Set LDFLAGS

			return run("go", "build", "-trimpath", "-o", exe, "./cmd/lstn")
		},
		"man": func(_ string) error {
			fmt.Fprintf(os.Stdout, "executing `%s` ...\n", "man")
			// TODO
			return nil
		},
		"clean": func(x string) error {
			fmt.Fprintf(os.Stdout, "executing `%s` ...\n", "clean")
			// TODO
			return nil
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

	if len(args) == 1 {
		if isTargetingWindows() {
			args[0] = fmt.Sprintf("%s.exe", args[0])
		}
	}

	for _, arg := range args {
		norm := filepath.ToSlash(strings.TrimSuffix(arg, ".exe"))
		c := commands[norm]
		if c == nil {
			fmt.Fprintf(os.Stderr, "unknown command `%s`.\n", norm)
			os.Exit(1)
		}

		err := c(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintf(os.Stderr, "failure while executing `%s`.\n", norm)
			os.Exit(1)
		}
	}
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
	set_x(args...)
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func set_x(args ...string) {
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
			} else {
				return nil
			}
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
