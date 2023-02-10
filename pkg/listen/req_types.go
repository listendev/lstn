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
	"encoding/json"
	"fmt"
	"os"

	"github.com/listendev/lstn/pkg/git"
	"github.com/listendev/lstn/pkg/npm"
	lstnos "github.com/listendev/lstn/pkg/os"
	"github.com/listendev/lstn/pkg/version"
)

// VerdictsRequest represents the payload for the verdicts listen.dev endpoint.
type VerdictsRequest struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Shasum  string `json:"shasum,omitempty"`
}

func NewVerdictsRequest(args []string) *VerdictsRequest {
	ret := &VerdictsRequest{}

	switch len(args) {
	case 3:
		ret.Shasum = args[2]

		fallthrough
	case 2:
		ret.Version = args[1]

		fallthrough
	case 1:
		ret.Name = args[0]
	}

	return ret
}

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

type AnalysisRequest struct {
	PackageLockJSON npm.PackageLockJSON `json:"package-lock"`
	Packages        npm.Packages        `json:"packages"`
	Context         *AnalysisContext    `json:"context"`
}

// MarshalJSON is a custom marshaler that encodes the
// content of the package lock in the receiving AnalysisRequest.
func (req *AnalysisRequest) MarshalJSON() ([]byte, error) {
	type AnalysisRequestAlias AnalysisRequest

	return json.Marshal(&struct {
		PackageLockJSON string `json:"package-lock"`
		*AnalysisRequestAlias
	}{
		PackageLockJSON:      req.PackageLockJSON.Encode(),
		AnalysisRequestAlias: (*AnalysisRequestAlias)(req),
	})
}

func NewAnalysisRequest(packageLockJSON npm.PackageLockJSON, packages npm.Packages) (*AnalysisRequest, error) {
	if packageLockJSON == nil {
		return nil, fmt.Errorf("couldn't create the analysis request")
	}
	if !packageLockJSON.Ok() || !packages.Ok() {
		return nil, fmt.Errorf("couldn't create the analysis request")
	}

	return &AnalysisRequest{
		packageLockJSON,
		packages,
		NewAnalysisContext(),
	}, nil
}