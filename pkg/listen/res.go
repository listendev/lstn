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
	"github.com/listendev/pkg/models"
)

type Verdict = models.Verdict
type Problem = models.Problem

type Package struct {
	// Name name of the package
	Name string `json:"name"`

	// Problems A list of problems
	Problems []Problem `json:"problems,omitempty"`

	// Shasum shasum of the package
	Digest *string `json:"digest,omitempty"`

	// Verdicts A list of verdicts
	Verdicts []Verdict `json:"verdicts"`

	// Version version of the package
	Version *string `json:"version,omitempty"`
}

type Response []Package

func (r Response) Verdicts() models.Verdicts {
	res := models.Verdicts{}
	for _, p := range r {
		for _, v := range p.Verdicts {
			res = append(res, v)
		}
	}

	return res
}

type responseError struct {
	Message string `json:"message"`
}
