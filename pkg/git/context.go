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
	"encoding/json"
	"fmt"

	"github.com/go-git/go-git/v5"
)

type User struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

func (u User) jsonValue() interface{} {
	var zero User
	if u == zero {
		return nil
	}

	return u
}

type Author struct {
	Name  string `json:"email,omitempty"`
	Email string `json:"name,omitempty"`
}

func (a Author) jsonValue() interface{} {
	var zero Author
	if a == zero {
		return nil
	}

	return a
}

type FetchURL struct {
	URL string `json:"url"`
}

type PushURL struct {
	URL string `json:"url"`
}

type Remote struct {
	FetchURL `json:"fetch"`
	PushURL  `json:"push"`
}

type Context struct {
	User    `json:"user,omitempty"`
	Author  `json:"author,omitempty"`
	Remotes map[string]*Remote `json:"remotes,omitempty"`
}

func (c Context) MarshalJSON() ([]byte, error) {
	// Avoid recursion
	type AliasContext Context

	type AliasWithI struct {
		AliasContext
		User   interface{} `json:"user,omitempty"`
		Author interface{} `json:"author,omitempty"`
	}

	return json.Marshal(AliasWithI{
		AliasContext: AliasContext(c),
		User:         c.User.jsonValue(),
		Author:       c.Author.jsonValue(),
	})
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

	conf, err := GetFinalConfig(repo)
	if err != nil {
		return nil, err
	}

	remotes := make(map[string]*Remote, len(conf.Remotes))
	for n, c := range conf.Remotes {
		r := &Remote{}
		if len(c.URLs) > 0 {
			r.FetchURL.URL = c.URLs[0]
			r.PushURL.URL = c.URLs[0]
		}
		remotes[n] = r
	}

	c := &Context{
		Remotes: remotes,
		Author:  Author(conf.Author),
		User:    User(conf.User),
	}

	return c, nil
}
