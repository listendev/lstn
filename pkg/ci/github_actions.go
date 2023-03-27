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
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v50/github"
)

// NewInfoFromGitHubEvent creates an Info instance using the the file on the GitHub action runner
// that contains the full event webhook payload.
//
// See https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables.
func NewInfoFromGitHubEvent() (*Info, error) {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, fmt.Errorf("couldn't find the GITHUB_EVENT_PATH environment variable")
	}

	evt, err := NewGitHubEventFromPath(eventPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode the GITHUB_EVENT_PATH file")
	}

	info := &Info{}

	info.Owner = *evt.Repo.Owner.Login
	info.Repo = *evt.Repo.Name
	if pullRequestNum := evt.PullRequest.Number; pullRequestNum != nil {
		// Pull request events
		info.Num = *pullRequestNum
	}
	if pullRequestBranch := evt.PullRequest.Head; pullRequestBranch != nil {
		// Pull request events
		if pullRequestShasum := pullRequestBranch.SHA; pullRequestShasum != nil {
			info.SHA = *pullRequestShasum
		}
		if pullRequestRef := pullRequestBranch.Ref; pullRequestRef != nil {
			info.Branch = *pullRequestRef
		}
	} else if headCommitShasum := evt.HeadCommit.ID; headCommitShasum != nil {
		// Push events
		info.SHA = *headCommitShasum
	}

	// Re-run events
	if info.Num == 0 && len(evt.CheckSuite.PullRequests) > 0 && evt.CheckSuite.PullRequests[0] != nil {
		pullRequest := evt.CheckSuite.PullRequests[0]
		info.Num = *pullRequest.Number
		info.Branch = *pullRequest.Head.Ref
		info.SHA = *pullRequest.Head.SHA
	}

	return info, nil
}

type GitHubEvent struct {
	Repo        github.PushEventRepository `json:"repository"`
	HeadCommit  github.HeadCommit          `json:"head_commit"`
	PullRequest github.PullRequest         `json:"pull_request"`
	CheckSuite  github.CheckSuite          `json:"check_suite"`
}

// NewGitHubEventFromPath creates a GitHubEvent by reading the GITHUB_EVENT_PATH file.
func NewGitHubEventFromPath(eventPath string) (*GitHubEvent, error) {
	f, err := os.Open(eventPath)
	if err != nil {
		return nil, err
	}

	e := &GitHubEvent{}
	if err := json.NewDecoder(f).Decode(e); err != nil {
		return nil, err
	}

	return e, nil
}
