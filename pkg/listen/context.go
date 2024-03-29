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
package listen

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/listendev/lstn/pkg/git"
	"github.com/listendev/lstn/pkg/npm"
	lstnos "github.com/listendev/lstn/pkg/os"
	"github.com/listendev/lstn/pkg/version"
)

type Context struct {
	ID      uuid.UUID         `json:"id"`
	Version version.Version   `json:"version"`
	Git     *git.Context      `json:"git,omitempty"`
	OS      *lstnos.Info      `json:"os,omitempty"`
	PMs     map[string]string `json:"packagemanagers,omitempty"`
}

func NewContext(funcs ...git.GetDirFunc) *Context {
	ret := &Context{
		ID: uuid.New(),
	}

	ret.Version = version.Get()
	ret.OS, _ = lstnos.NewInfo()

	for _, f := range funcs {
		var err error
		ret.Git, err = git.NewContextFromFunc(f)
		if err == nil {
			break
		}
	}
	if ret.Git == nil {
		ret.Git, _ = git.NewContextFromFunc(os.Getwd)
	}

	npmVersion, err := npm.Version(context.TODO())
	if err == nil {
		if ret.PMs == nil {
			ret.PMs = make(map[string]string)
		}
		ret.PMs["npm"] = npmVersion
	}

	return ret
}
