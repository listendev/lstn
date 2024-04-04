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
package pypi

import (
	"encoding/base64"

	"github.com/listendev/lstn/pkg/validate"
)

type poetryLock struct {
	bytes []byte
}

// Encode encodes the receiving PoetryLock instance in a base64 string.
func (p *poetryLock) Encode() string {
	return base64.StdEncoding.EncodeToString(p.bytes)
}

func (p *poetryLock) Ok() bool {
	err := validate.Singleton.Struct(p)

	return err == nil
}
