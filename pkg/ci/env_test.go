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
package ci

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestNewInfo_GitHubActionsWithoutEventPath(t *testing.T) {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS": "true",
	})
	t.Cleanup(closer)

	assert.True(t, IsRunningInGitHubAction())

	res, err := NewInfo()
	if assert.Error(t, err) {
		assert.Equal(t, "couldn't find the GITHUB_EVENT_PATH environment variable", err.Error())
	}
	assert.Nil(t, res)
}

func TestNewInfo_GitHubActionsPushEvent(t *testing.T) {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_PATH": "testdata/github_event_push.json",
	})
	t.Cleanup(closer)

	assert.True(t, IsRunningInGitHubAction())

	got, err := NewInfo()
	assert.Nil(t, err)

	exp := &Info{
		Owner: "reviewdog",
		Repo:  "reviewdog",
		SHA:   "febdd4bf26c6e8856c792303cfc66fa5e7bc975b",
	}
	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("info mismatch (-want +got):\n%s", diff)
	}

	assert.False(t, got.IsPullRequest())
}

func TestNewInfo_GitHubActionsPullRequestEvent(t *testing.T) {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_PATH": "testdata/github_event_pull_request.json",
	})
	t.Cleanup(closer)

	assert.True(t, IsRunningInGitHubAction())

	got, err := NewInfo()
	assert.Nil(t, err)

	exp := &Info{
		Owner:  "reviewdog",
		Repo:   "reviewdog",
		SHA:    "cb23119096646023c05e14ea708b7f20cee906d5",
		Num:    285,
		Branch: "go1.13",
	}
	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("info mismatch (-want +got):\n%s", diff)
	}

	assert.True(t, got.IsPullRequest())
}

func TestNewInfo_GitHubActionsReRunEvent(t *testing.T) {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_PATH": "testdata/github_event_rerun.json",
	})
	t.Cleanup(closer)

	assert.True(t, IsRunningInGitHubAction())

	got, err := NewInfo()
	assert.Nil(t, err)

	exp := &Info{
		Owner:  "reviewdog",
		Repo:   "reviewdog",
		SHA:    "ba8f36cd3eb401e9de9ee5718e11d390fdbe4afa",
		Num:    286,
		Branch: "github-actions-env",
	}
	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("info mismatch (-want +got):\n%s", diff)
	}

	assert.True(t, got.IsPullRequest())
}
