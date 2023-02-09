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
	"context"
	"testing"
	"time"

	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/stretchr/testify/assert"
)

var (
	localEndpoint    = "http://127.0.0.1:3000"
	nonLocalEndpoint = "https://smtg.listen.dev"
)

type mockContextLocalEndpoint struct{}

func (ctx mockContextLocalEndpoint) Deadline() (deadline time.Time, ok bool) {
	return deadline, ok
}

func (ctx mockContextLocalEndpoint) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

func (ctx mockContextLocalEndpoint) Err() error {
	return context.DeadlineExceeded
}

func (ctx mockContextLocalEndpoint) Value(key interface{}) interface{} {
	// pkgcontext.ConfigKey
	if key == pkgcontext.ConfigKey {
		c, _ := flags.NewConfigOptions()
		c.Endpoint = localEndpoint

		return c
	}

	return nil
}

func TestLocalEndpoint(t *testing.T) {
	endpoint, err := getBaseURLFromContext(mockContextLocalEndpoint{})
	assert.Nil(t, err)
	assert.Equal(t, localEndpoint, endpoint)
	assert.Equal(t, "/api/npm", getAPIPrefix(endpoint))
}

type mockContextNonLocalEndpoint struct{}

func (ctx mockContextNonLocalEndpoint) Deadline() (deadline time.Time, ok bool) {
	return deadline, ok
}

func (ctx mockContextNonLocalEndpoint) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

func (ctx mockContextNonLocalEndpoint) Err() error {
	return context.DeadlineExceeded
}

func (ctx mockContextNonLocalEndpoint) Value(key interface{}) interface{} {
	// pkgcontext.ConfigKey
	if key == pkgcontext.ConfigKey {
		c, _ := flags.NewConfigOptions()
		c.Endpoint = nonLocalEndpoint

		return c
	}

	return nil
}

func TestNonLocalEndpoint(t *testing.T) {
	endpoint, err := getBaseURLFromContext(mockContextNonLocalEndpoint{})
	assert.Nil(t, err)
	assert.Equal(t, nonLocalEndpoint, endpoint)
	assert.Equal(t, "/api", getAPIPrefix(endpoint))
}
