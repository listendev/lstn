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
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
)

type NPM struct {
	Version string
}

// StubNPM creates a fake npm for testing reasons.
func StubNpm(npm NPM) error {
	args := os.Args[1:]
	if len(args) < 1 {
		return fmt.Errorf("fake npm without arguments")
	}

	switch args[0] {
	case "--version":
		fmt.Println(npm.Version)

		return nil
	case "i":
		fallthrough
	case "install":
		if len(args) > 1 {
			// TODO:: --package-lock-only --audit
		}
		fmt.Println("installing...")

		return nil
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
