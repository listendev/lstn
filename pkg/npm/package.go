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
package npm

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func readPackageJSON(dir string) (io.Reader, error) {
	name := filepath.Join(dir, "package.json")

	f, err := activeFS.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory %s does not contain a package.json file", dir)
		}

		return nil, fmt.Errorf("couldn't read the package.json file")
	}

	return f, nil
}
