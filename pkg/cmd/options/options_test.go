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
package options

import (
	"testing"

	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestOptionsGotConfigDefaultsFromEnv(t *testing.T) {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_PATH": "../../ci/testdata/github_event_pull_request.json",
	})
	t.Cleanup(closer)

	rootOpts, err := NewRoot()
	assert.Nil(t, err)
	assert.Equal(t, 285, rootOpts.ID)
	assert.Equal(t, "reviewdog", rootOpts.Owner)
	assert.Equal(t, "reviewdog", rootOpts.Repo)
	assert.Equal(t, "https://registry.npmjs.org", rootOpts.NPM)
	assert.Equal(t, "info", rootOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", rootOpts.Endpoint.Npm)
	assert.Equal(t, "https://pypi.listen.dev", rootOpts.Endpoint.PyPi)
	assert.Equal(t, 60, rootOpts.Timeout)

	inOpts, err := NewIn()
	assert.Nil(t, err)
	assert.Equal(t, 285, inOpts.ID)
	assert.Equal(t, "reviewdog", inOpts.Owner)
	assert.Equal(t, "reviewdog", inOpts.Repo)
	assert.Equal(t, "https://registry.npmjs.org", inOpts.NPM)
	assert.Equal(t, "info", inOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", inOpts.Endpoint.Npm)
	assert.Equal(t, "https://pypi.listen.dev", inOpts.Endpoint.PyPi)
	assert.Equal(t, 60, inOpts.Timeout)

	scanOpts, err := NewScan()
	assert.Nil(t, err)
	assert.Equal(t, 285, scanOpts.ID)
	assert.Equal(t, "reviewdog", scanOpts.Owner)
	assert.Equal(t, "reviewdog", scanOpts.Repo)
	assert.Equal(t, "https://registry.npmjs.org", scanOpts.NPM)
	assert.Equal(t, "info", scanOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", scanOpts.Endpoint.Npm)
	assert.Equal(t, "https://pypi.listen.dev", scanOpts.Endpoint.PyPi)
	assert.Equal(t, 60, scanOpts.Timeout)

	toOpts, err := NewTo()
	assert.Nil(t, err)
	assert.Equal(t, 285, toOpts.ID)
	assert.Equal(t, "reviewdog", toOpts.Owner)
	assert.Equal(t, "reviewdog", toOpts.Repo)
	assert.Equal(t, "https://registry.npmjs.org", toOpts.NPM)
	assert.Equal(t, "info", toOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", toOpts.Endpoint.Npm)
	assert.Equal(t, "https://pypi.listen.dev", toOpts.Endpoint.PyPi)
	assert.Equal(t, 60, toOpts.Timeout)
}
