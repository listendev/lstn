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
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ghetzel/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestNewPoetryLockFromReader(t *testing.T) {
	fixture, errFixture := os.Open("testdata/poetry.lock")
	require.Nil(t, errFixture)
	defer fixture.Close()

	var b bytes.Buffer
	r := io.TeeReader(fixture, &b)

	lockFromDir, errFromDir := NewPoetryLockFromReader(r)
	assert.Nil(t, errFromDir)
	require.IsType(t, &poetryLock{}, lockFromDir)

	assert.Equal(t, b.Bytes(), lockFromDir.(*poetryLock).bytes)
}
