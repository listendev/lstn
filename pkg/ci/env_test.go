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
	"github.com/google/go-cmp/cmp/cmpopts"
	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestNewInfo_GitHubActionsWithoutEventPath(t *testing.T) {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_PATH": "",
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
	if diff := cmp.Diff(exp, got, cmpopts.IgnoreFields(Info{}, "EventName", "Action", "ActionPath", "ActionRepository", "Actor", "ActorID", "Job", "Ref", "RefName", "RefProtected", "RefType", "RepoFullName", "RepoID", "RepoOwner", "RepoOwnerID", "RunAttempt", "RunID", "RunNumber", "RunnerArch", "RunnerOs", "SeverURL", "TriggeringActor", "Workflow", "WorkflowRef", "WorkflowSha", "Workspace")); diff != "" {
		t.Errorf("info mismatch (-want +got):\n%s", diff)
	}

	assert.False(t, got.IsGitHubPullRequest())

	gotDump := got.Dump()
	assert.Contains(t, gotDump, "OWNER=reviewdog")
	assert.Contains(t, gotDump, "REPO=reviewdog")
	assert.Contains(t, gotDump, "SHA=febdd4bf26c6e8856c792303cfc66fa5e7bc975b")
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
	if diff := cmp.Diff(exp, got, cmpopts.IgnoreFields(Info{}, "EventName", "Action", "ActionPath", "ActionRepository", "Actor", "ActorID", "Job", "Ref", "RefName", "RefProtected", "RefType", "RepoFullName", "RepoID", "RepoOwner", "RepoOwnerID", "RunAttempt", "RunID", "RunNumber", "RunnerArch", "RunnerOs", "SeverURL", "TriggeringActor", "Workflow", "WorkflowRef", "WorkflowSha", "Workspace")); diff != "" {
		t.Errorf("info mismatch (-want +got):\n%s", diff)
	}

	assert.True(t, got.IsGitHubPullRequest())

	gotDump := got.Dump()
	assert.Contains(t, gotDump, "BRANCH=go1.13")
	assert.Contains(t, gotDump, "NUM=285")
	assert.Contains(t, gotDump, "OWNER=reviewdog")
	assert.Contains(t, gotDump, "REPO=reviewdog")
	assert.Contains(t, gotDump, "SHA=cb23119096646023c05e14ea708b7f20cee906d5")
}

func TestNewInfo_GitHubActionsPullRequestForkEvent(t *testing.T) {
	closer := internaltesting.EnvSetter(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_PATH": "testdata/github_event_pull_request_from_fork.json",
	})
	t.Cleanup(closer)

	assert.True(t, IsRunningInGitHubAction())

	got, err := NewInfo()
	assert.Nil(t, err)

	exp := &Info{
		Owner:  "listendev",
		Repo:   "lstn",
		SHA:    "5864e328f129b98726813940b5bfa44707963bdc",
		Num:    217,
		Branch: "build/pkg",
		Fork:   true,
	}
	if diff := cmp.Diff(exp, got, cmpopts.IgnoreFields(Info{}, "EventName", "Action", "ActionPath", "ActionRepository", "Actor", "ActorID", "Job", "Ref", "RefName", "RefProtected", "RefType", "RepoFullName", "RepoID", "RepoOwner", "RepoOwnerID", "RunAttempt", "RunID", "RunNumber", "RunnerArch", "RunnerOs", "SeverURL", "TriggeringActor", "Workflow", "WorkflowRef", "WorkflowSha", "Workspace")); diff != "" {
		t.Errorf("info mismatch (-want +got):\n%s", diff)
	}

	assert.True(t, got.IsGitHubPullRequest())

	gotDump := got.Dump()
	assert.Contains(t, gotDump, "BRANCH=build/pkg")
	assert.Contains(t, gotDump, "FORK=true")
	assert.Contains(t, gotDump, "NUM=217")
	assert.Contains(t, gotDump, "OWNER=listendev")
	assert.Contains(t, gotDump, "REPO=lstn")
	assert.Contains(t, gotDump, "SHA=5864e328f129b98726813940b5bfa44707963bdc")
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
	if diff := cmp.Diff(exp, got, cmpopts.IgnoreFields(Info{}, "EventName", "Action", "ActionPath", "ActionRepository", "Actor", "ActorID", "Job", "Ref", "RefName", "RefProtected", "RefType", "RepoFullName", "RepoID", "RepoOwner", "RepoOwnerID", "RunAttempt", "RunID", "RunNumber", "RunnerArch", "RunnerOs", "SeverURL", "TriggeringActor", "Workflow", "WorkflowRef", "WorkflowSha", "Workspace")); diff != "" {
		t.Errorf("info mismatch (-want +got):\n%s", diff)
	}

	assert.True(t, got.IsGitHubPullRequest())
}
