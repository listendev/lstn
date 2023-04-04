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
	assert.Equal(t, 285, rootOpts.ConfigFlags.Reporting.GitHub.Pull.ID)
	assert.Equal(t, "reviewdog", rootOpts.ConfigFlags.Reporting.GitHub.Owner)
	assert.Equal(t, "reviewdog", rootOpts.ConfigFlags.Reporting.GitHub.Repo)
	assert.Equal(t, "https://registry.npmjs.org", rootOpts.Registry.NPM)
	assert.Equal(t, "info", rootOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", rootOpts.Endpoint)
	assert.Equal(t, 60, rootOpts.Timeout)

	inOpts, err := NewIn()
	assert.Nil(t, err)
	assert.Equal(t, 285, inOpts.ConfigFlags.Reporting.GitHub.Pull.ID)
	assert.Equal(t, "reviewdog", inOpts.ConfigFlags.Reporting.GitHub.Owner)
	assert.Equal(t, "reviewdog", inOpts.ConfigFlags.Reporting.GitHub.Repo)
	assert.Equal(t, "https://registry.npmjs.org", inOpts.Registry.NPM)
	assert.Equal(t, "info", inOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", inOpts.Endpoint)
	assert.Equal(t, 60, inOpts.Timeout)

	scanOpts, err := NewScan()
	assert.Nil(t, err)
	assert.Equal(t, 285, scanOpts.ConfigFlags.Reporting.GitHub.Pull.ID)
	assert.Equal(t, "reviewdog", scanOpts.ConfigFlags.Reporting.GitHub.Owner)
	assert.Equal(t, "reviewdog", scanOpts.ConfigFlags.Reporting.GitHub.Repo)
	assert.Equal(t, "https://registry.npmjs.org", scanOpts.Registry.NPM)
	assert.Equal(t, "info", scanOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", scanOpts.Endpoint)
	assert.Equal(t, 60, scanOpts.Timeout)

	toOpts, err := NewTo()
	assert.Nil(t, err)
	assert.Equal(t, 285, toOpts.ConfigFlags.Reporting.GitHub.Pull.ID)
	assert.Equal(t, "reviewdog", toOpts.ConfigFlags.Reporting.GitHub.Owner)
	assert.Equal(t, "reviewdog", toOpts.ConfigFlags.Reporting.GitHub.Repo)
	assert.Equal(t, "https://registry.npmjs.org", toOpts.Registry.NPM)
	assert.Equal(t, "info", toOpts.LogLevel)
	assert.Equal(t, "https://npm.listen.dev", toOpts.Endpoint)
	assert.Equal(t, 60, toOpts.Timeout)
}
