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
	"os"

	"github.com/listendev/lstn/pkg/git"
	lstnos "github.com/listendev/lstn/pkg/os"
	"github.com/listendev/lstn/pkg/version"
)

type AnalysisContext struct {
	Version version.Version `json:"version"`
	Git     *git.Context    `json:"git,omitempty"`
	OS      *lstnos.Info    `json:"os,omitempty"`
}

func NewAnalysisContext(funcs ...git.GetDirFunc) *AnalysisContext {
	ret := &AnalysisContext{}

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

	return ret
}
