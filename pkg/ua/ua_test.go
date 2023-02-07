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
