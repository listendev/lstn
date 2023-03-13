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
package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDepsByType(t *testing.T) {
	e1 := map[string]string{
		"react": "18.0.0",
	}

	p1 := &packageJSON{
		Dependencies: e1,
	}

	assert.Equal(t, e1, p1.getDepsByType(Dependencies))
	assert.Equal(t, map[string]string{}, p1.getDepsByType(DevDependencies))

	e2 := map[string]string{
		"react":      "18.0.0",
		"node-fetch": "2.6.9",
	}

	p2 := &packageJSON{
		DevDependencies: e2,
	}

	assert.Equal(t, map[string]string{}, p2.getDepsByType(Dependencies))
	assert.Equal(t, e2, p2.getDepsByType(DevDependencies))
}
