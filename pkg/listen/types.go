/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package listen

import (
	"encoding/json"

	"github.com/listendev/lstn/pkg/npm"
)

type VerdictsRequest struct {
	Name    string `url:"n"`
	Version string `url:"v,omitempty"`
	Shasum  string `url:"s,omitempty"`
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

type AnalysisRequest struct {
	PackageLockJSON npm.PackageLockJSON `json:"package-lock.json"`
	Packages        npm.Packages        `json:"packages"`
	Context         string              `json:"context"` // TODO > define
}

// MarshalJSON is a custom marshaler that encodes the
// content of the package lock in the receiving AnalysisRequest.
func (req *AnalysisRequest) MarshalJSON() ([]byte, error) {
	type AnalysisRequestAlias AnalysisRequest

	return json.Marshal(&struct {
		PackageLockJSON string `json:"package-lock.json"`
		*AnalysisRequestAlias
	}{
		PackageLockJSON:      req.PackageLockJSON.Encode(),
		AnalysisRequestAlias: (*AnalysisRequestAlias)(req),
	})
}

type Verdict struct {
	Message         string            `json:"message"`
	Priority        string            `json:"priority"`
	PackageMetadata map[string]string `json:"package_metadata,omitempty"`
	ProcessMetadata map[string]string `json:"process_metadata,omitempty"`
}

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type Package struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Shasum   string    `json:"shasum"`
	Verdicts []Verdict `json:"verdicts,omitempty"`
	Problems []Problem `json:"problems,omitempty"`
}

type Response []Package

type responseError struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
