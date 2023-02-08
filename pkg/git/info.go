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
package git

import (
	"fmt"

	"github.com/go-git/go-git/v5" // FIXME: switch to other impl/dep/fork?
)

type Context struct {
	Remotes []*git.Remote `json:"remotes,omitempty"` // FIXME: map these to internal type?
}

type GetDirFunc func() (string, error)

func NewContextFromFunc(f GetDirFunc) (*Context, error) {
	dir, err := f()
	if err != nil {
		return nil, fmt.Errorf("couldn't get the directory where to look for a git repository")
	}

	return NewContextFromPath(dir)
}

func NewContextFromPath(path string) (*Context, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open the git repository at %s", path)
	}

	remotes, _ := repo.Remotes()
	c := &Context{
		Remotes: remotes,
	}

	// FIXME
	// The current implementation does not respect the Git config semantics:
	// repo.Config() doesn't merge the user/system Git configuration.
	// This may result in a few issues (eg., querying remote references).
	// See https://github.com/go-git/go-git/issues/395

	return c, nil
}
