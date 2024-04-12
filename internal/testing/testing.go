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
package testing

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func MockHTTPServer(assert *assert.Assertions, path string, resp []byte, status int, wantMethod string) *httptest.Server {
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != wantMethod {
			assert.Failf("expected a %s request, got %s", wantMethod, r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, path) {
			assert.Failf("expected to request .../analysis or .../verdicts, got %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			assert.Failf("expected content-type: application/json header, got: %s", ct)
		}

		w.WriteHeader(status)
		_, err := w.Write(resp)
		assert.Nil(err)
	}))
}

type NPM struct {
	Version      string
	WithNVM      bool
	WithNVMNoUse bool
}

// StubNPM creates a fake npm for testing reasons.
func StubNpm(npm NPM) error {
	args := os.Args[1:]
	if len(args) < 1 {
		return fmt.Errorf("fake npm without arguments")
	}

	if npm.WithNVM {
		if len(args) >= 2 {
			if args[0] == "-c" {
				if strings.HasPrefix(args[1], "source") && strings.Contains(args[1], "nvm.sh") && strings.Contains(args[1], "npm") && strings.HasSuffix(args[1], "--version") {
					if npm.WithNVMNoUse && !strings.Contains(args[1], "--no-use") {
						return fmt.Errorf("missing --no-use nvm flag")
					}
					fmt.Println(npm.Version)

					return nil
				}
			}
		}
	} else {
		switch args[0] {
		case "--version":
			fmt.Println(npm.Version)

			return nil
		case "i":
			fallthrough
		case "install":
			// if len(args) > 1 {
			// 	// TODO:: --package-lock-only --audit
			// }
			fmt.Println("installing...")

			return nil
		}
	}

	return fmt.Errorf("couldn't fake npm correctly")
}

// WriteFileContent writes content to a path inside a billy.Filesystem.
// The containing directories (and any parent) are created as needed using fs.MkdirAll().
func WriteFileContent(fs billy.Filesystem, path string, fileContent string, executable bool) error {
	// Ensure the parent folder exists
	pathDir := filepath.Dir(path)
	if err := fs.MkdirAll(pathDir, os.ModePerm); err != nil {
		return err
	}

	// Set file permissions
	perms := os.FileMode(0666)
	if executable {
		perms = os.FileMode(0777)
	}

	// Create the file
	f, err := fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perms)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the content
	_, err = f.Write([]byte(fileContent))
	if err != nil {
		return err
	}

	return f.Close()
}

func CopyFile(src, dst string) error {
	i, err := os.Open(src)
	if err != nil {
		return err
	}
	defer i.Close()

	o, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer o.Close()

	if _, err = io.Copy(o, i); err != nil {
		return err
	}

	return nil
}

func CopyExecutable(src, dst string) error {
	err := CopyFile(src, dst)
	if err != nil {
		return err
	}

	return os.Chmod(dst, 0o755)
}

func ExecuteCommand(root *cobra.Command, cmdline ...string) (string, string, error) {
	bOut := bytes.NewBufferString("")
	root.SetOut(bOut)

	bErr := bytes.NewBufferString("")
	root.SetErr(bErr)

	root.SetArgs(cmdline)
	execErr := root.Execute()
	outOut, err := io.ReadAll(bOut)
	if err != nil {
		return "", "", fmt.Errorf("error while reading the stdout: %s %s", root.Name(), strings.Join(cmdline, " "))
	}

	outErr, err := io.ReadAll(bErr)
	if err != nil {
		return "", "", fmt.Errorf("error while reading the stderr: %s %s", root.Name(), strings.Join(cmdline, " "))
	}

	return string(outOut), string(outErr), execErr
}

func EnvSetter(envs map[string]string) (closer func()) {
	originalEnvs := map[string]string{}

	for name, value := range envs {
		if originalValue, ok := os.LookupEnv(name); ok {
			originalEnvs[name] = originalValue
		}
		_ = os.Setenv(name, value)
	}

	return func() {
		for name := range envs {
			origValue, has := originalEnvs[name]
			if has {
				_ = os.Setenv(name, origValue)
			} else {
				_ = os.Unsetenv(name)
			}
		}
	}
}

func StringsReplaceAll(slice []string, with string) []string {
	ret := []string{}
	for _, s := range slice {
		ret = append(ret, strings.ReplaceAll(s, "_CWD_", with))
	}

	return ret
}
