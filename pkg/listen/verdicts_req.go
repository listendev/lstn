// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2024 The listen.dev team <engineering@garnet.ai>
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
	"github.com/listendev/lstn/pkg/jsonpath"
	"github.com/listendev/lstn/pkg/validate"
)

type Request interface {
	IsRequest() bool
	Ok() (bool, error)
}

// VerdictsRequest represents the payload for the verdicts listen.dev API endpoint.
type VerdictsRequest struct {
	Name    string   `json:"name"              name:"name"                 validate:"mandatory"`
	Version string   `json:"version,omitempty" validate:"omitempty,semver"`
	Digest  string   `json:"digest,omitempty"  validate:"omitempty,digest"`
	Select  string   `json:"select,omitempty"`
	Context *Context `json:"context,omitempty"`
}

func fillVerdictsRequest(r *VerdictsRequest, args []string) (*VerdictsRequest, error) {
	switch len(args) {
	case 3:
		r.Digest = args[2]

		fallthrough
	case 2:
		r.Version = args[1]

		fallthrough
	case 1:
		if len(args[0]) == 0 {
			return nil, fmt.Errorf("a verdicts request requires at least one argument (package name)")
		}
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

func NewBulkVerdictsRequestsFromMap(deps map[string]*semver.Version, selection string) ([]*VerdictsRequest, error) {
	if len(deps) == 0 {
		return nil, fmt.Errorf("couldn't create a request set from empty dependencies map")
	}

	c := NewContext()

	i := 0
	reqs := make([]*VerdictsRequest, len(deps))
	for name, vers := range deps {
		inputs := []string{name}
		if vers != nil {
			inputs = append(inputs, vers.String())
		}
		var reqErr error
		reqs[i], reqErr = NewVerdictsRequestWithContext(inputs, c)
		if reqErr != nil {
			return nil, reqErr
		}
		reqs[i].Select = jsonpath.Make(selection)

		i++
	}

	return reqs, nil
}

func NewBulkVerdictsRequests(names []string, versions semver.Collection, selection string) ([]*VerdictsRequest, error) {
	if len(names) != len(versions) {
		return nil, fmt.Errorf("couldn't create a request set because of mismatching lengths")
	}

	c := NewContext()

	reqs := make([]*VerdictsRequest, len(versions))
	for i, v := range versions {
		inputs := []string{names[i], v.String()}
		var reqErr error
		reqs[i], reqErr = NewVerdictsRequestWithContext(inputs, c)
		if reqErr != nil {
			return nil, reqErr
		}
		reqs[i].Select = jsonpath.Make(selection)
	}

	return reqs, nil
}

func (req VerdictsRequest) IsRequest() bool {
	return true
}

func (req *VerdictsRequest) Ok() (bool, error) {
	err := validate.Singleton.Struct(req)
	if err != nil {
		if all, isValidationErrors := err.(validate.ValidationError); isValidationErrors {
			return false, fmt.Errorf("%s", all[0].Translate(validate.Translator)) // Only the first one
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
