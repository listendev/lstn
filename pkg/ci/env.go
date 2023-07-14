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
package ci

import (
	"fmt"
)

type Info struct {
	Owner  string
	Repo   string
	SHA    string
	Num    int // Pull (merge) request number
	Branch string
}

func (i *Info) IsPullRequest() bool {
	return i.Num != 0
}

// NewInfo creates a Info from environment variables.
func NewInfo() (*Info, error) {
	if IsRunningInGitHubAction() {
		return NewInfoFromGitHubEvent()
	}

	// TODO: implement logic for other CI systems

	return nil, fmt.Errorf("CI systems other than GitHub Actions are not supported yet")
}
