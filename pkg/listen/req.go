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

	"github.com/Masterminds/semver/v3"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/listendev/lstn/pkg/validate"
)

type Request interface {
	IsRequest() bool
	Ok() (bool, error)
}

// VerdictsRequest represents the payload for the verdicts listen.dev endpoint.
type VerdictsRequest struct {
	Name    string   `json:"name" name:"name" validate:"mandatory"`
	Version string   `json:"version,omitempty" validate:"omitempty,semver"`
	Shasum  string   `json:"shasum,omitempty" validate:"omitempty,shasum"`
	Context *Context `json:"context,omitempty"`
}

func fillVerdictsRequest(r *VerdictsRequest, args []string) (*VerdictsRequest, error) {
	switch len(args) {
	case 3:
		r.Shasum = args[2]

		fallthrough
	case 2:
		// TODO: resolve version here?
		r.Version = args[1]

		fallthrough
	case 1:
		r.Name = args[0]

	default:
		return nil, fmt.Errorf("a verdicts request requires at least one argument (package name)")
	}

	return r, nil
}

func NewVerdictsRequestWithContext(args []string, c *Context) (*VerdictsRequest, error) {
	ret := &VerdictsRequest{
		Context: c,
	}

	return fillVerdictsRequest(ret, args)
}

func NewVerdictsRequest(args []string) (*VerdictsRequest, error) {
	ret := &VerdictsRequest{
		Context: NewContext(),
	}

	return fillVerdictsRequest(ret, args)
}

func NewVerdictsRequestsFromVersionCollection(args []string, versions semver.Collection) ([]*VerdictsRequest, error) {
	c := NewContext()

	reqs := make([]*VerdictsRequest, len(versions))
	for i, v := range versions {
		inputs := []string{args[0], v.String()}
		if len(args) > 2 {
			inputs = append(inputs, args[2])
		}
		var reqErr error
		reqs[i], reqErr = NewVerdictsRequestWithContext(inputs, c)
		if reqErr != nil {
			return nil, reqErr
		}
	}

	return reqs, nil
}

func (req VerdictsRequest) IsRequest() bool {
	return true
}

func (req *VerdictsRequest) Ok() (bool, error) {
	err := validate.Singleton.Struct(req)
	if err != nil {
		if all, isValidationErrors := err.(validate.ValidationErrors); isValidationErrors {
			return false, fmt.Errorf(all[0].Translate(validate.Translator))
		}

		return false, err
	}

	return true, nil
}

func (req VerdictsRequest) MarshalJSON() ([]byte, error) {
	if isOk, err := req.Ok(); !isOk {
		return nil, err
	}

	type VerdictsRequestAlias VerdictsRequest

	return json.Marshal(&struct {
		*VerdictsRequestAlias
	}{
		VerdictsRequestAlias: (*VerdictsRequestAlias)(&req),
	})
}

type AnalysisRequest struct {
	PackageLockJSON npm.PackageLockJSON `json:"package-lock" name:"package lock" validate:"mandatory"`
	Context         *Context            `json:"context,omitempty"`
}

func (req *AnalysisRequest) Ok() (bool, error) {
	err := validate.Singleton.Struct(req)
	if err != nil {
		if all, isValidationErrors := err.(validate.ValidationErrors); isValidationErrors {
			return false, fmt.Errorf(all[0].Translate(validate.Translator))
		}

		return false, err
	}

	return true, nil
}

// MarshalJSON is a custom marshaler that encodes the
// content of the package lock in the receiving AnalysisRequest.
func (req AnalysisRequest) MarshalJSON() ([]byte, error) {
	if isOk, err := req.Ok(); !isOk {
		return nil, err
	}

	type AnalysisRequestAlias AnalysisRequest

	return json.Marshal(&struct {
		PackageLockJSON string `json:"package-lock"`
		*AnalysisRequestAlias
	}{
		PackageLockJSON:      req.PackageLockJSON.Encode(),
		AnalysisRequestAlias: (*AnalysisRequestAlias)(&req),
	})
}

func NewAnalysisRequest(packageLockJSON npm.PackageLockJSON) (*AnalysisRequest, error) {
	if packageLockJSON == nil {
		return nil, fmt.Errorf("couldn't create the analysis request")
	}
	if !packageLockJSON.Ok() {
		return nil, fmt.Errorf("couldn't create the analysis request because of invalid package-lock.json")
	}

	return &AnalysisRequest{
		packageLockJSON,
		NewContext(),
	}, nil
}

func (req AnalysisRequest) IsRequest() bool {
	return true
}
