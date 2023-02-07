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
package ua

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgentWithoutOSInfo(t *testing.T) {
	res1 := Generate(false, "ciao")

	assert.True(t, strings.HasPrefix(res1, "lstn"))
	assert.True(t, strings.Contains(res1, "TestUserAgentWithoutOSInfo"))
	assert.True(t, strings.Contains(res1, "ciao)"))
	assert.True(t, strings.HasSuffix(res1, ")"))

	res2 := Generate(false)

	assert.True(t, strings.HasPrefix(res2, "lstn"))
	assert.True(t, strings.Contains(res2, "TestUserAgentWithoutOSInfo"))
	assert.True(t, strings.HasSuffix(res2, ")"))
}

func TestUserAgentWithOSInfo(t *testing.T) {
	res1 := Generate(true, "ciao")

	assert.True(t, strings.HasPrefix(res1, "lstn"))
	assert.True(t, strings.Contains(res1, "ciao)"))
	assert.True(t, strings.Contains(res1, "TestUserAgentWithOSInfo"))
	assert.False(t, strings.HasSuffix(res1, ")"))

	res2 := Generate(true)

	assert.True(t, strings.HasPrefix(res2, "lstn"))
	assert.True(t, strings.Contains(res2, "TestUserAgentWithOSInfo"))
	assert.False(t, strings.HasSuffix(res2, ")"))
}

func TestUserAgentMoreComments(t *testing.T) {
	res1 := Generate(true, "ciao", "hello")

	assert.True(t, strings.HasPrefix(res1, "lstn"))
	assert.True(t, strings.Contains(res1, "ciao;"))
	assert.True(t, strings.Contains(res1, "hello)"))
	assert.True(t, strings.Contains(res1, "TestUserAgentMoreComments"))
	assert.False(t, strings.HasSuffix(res1, ")"))
}
