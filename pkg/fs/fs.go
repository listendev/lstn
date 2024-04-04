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
package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

// Current provides a configurable entry point to the billy.Filesystem in active use
// for reading global and system level configuration files.
//
// Override this in tests to mock the filesystem
// (then reset to restore default behavior).
var Current = defaultFS()

// defaultFS provides a billy.Filesystem abstraction over the
// OS filesystem (via osfs.OS) scoped to the root directory
// in order to enable access to global and system configuration files
// via absolute paths.
func defaultFS() billy.Filesystem {
	return osfs.New("/")
}

func Read(dir, filename string) (io.Reader, error) {
	name := filepath.Join(dir, filename)

	f, err := Current.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory %s does not contain the %s file", dir, filename)
		} else if os.IsPermission(err) {
			return nil, fmt.Errorf("insufficient permission to open %s", name)
		}

		return nil, fmt.Errorf("couldn't read the %s file", filename)
	}

	return f, nil
}
