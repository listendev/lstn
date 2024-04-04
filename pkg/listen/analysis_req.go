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

	listentype "github.com/listendev/lstn/pkg/listen/type"
	"github.com/listendev/lstn/pkg/validate"
)

// AnalysisRequest represents the payload for the analysis listen.dev API endpoint.
type AnalysisRequest struct {
	Manifest listentype.AnalysisRequester `json:"manifest" name:"manifest" validate:"mandatory"`
	Context  *Context                     `json:"context,omitempty"`
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
		Manifest string `json:"manifest"`
		*AnalysisRequestAlias
	}{
		Manifest:             req.Manifest.Encode(),
		AnalysisRequestAlias: (*AnalysisRequestAlias)(&req),
	})
}

type RequestOption func(*AnalysisRequest)

func WithRequestContext() RequestOption {
	return func(req *AnalysisRequest) {
		req.Context = NewContext()
	}
}

func NewAnalysisRequest(lockfile listentype.AnalysisRequester, opts ...RequestOption) (*AnalysisRequest, error) {
	if lockfile == nil {
		return nil, fmt.Errorf("couldn't create the analysis request")
	}
	if !lockfile.Ok() {
		return nil, fmt.Errorf("couldn't create the analysis request because of invalid lockfile")
	}

	req := &AnalysisRequest{lockfile, nil}

	for _, opt := range opts {
		opt(req)
	}

	return req, nil
}

func (req AnalysisRequest) IsRequest() bool {
	return true
}
