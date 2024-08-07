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
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
)

type Info struct {
	Owner            string
	Repo             string
	SHA              string `dump:"GITHUB_SHA"`
	Num              int    // Pull (merge) request number
	Branch           string // Pull (merge) request branch
	Fork             bool
	Action           string `env:"GITHUB_ACTION"`
	ActionPath       string `env:"GITHUB_ACTION_PATH"`
	ActionRepository string `env:"GITHUB_ACTION_REPOSITORY"`
	Actor            string `env:"GITHUB_ACTOR"`
	ActorID          int64  `env:"GITHUB_ACTOR_ID"`
	EventName        string `dump:"GITHUB_EVENT_NAME"`
	Job              string `env:"GITHUB_JOB"`
	Ref              string `env:"GITHUB_REF"`
	RefName          string `env:"GITHUB_REF_NAME"`
	RefProtected     bool   `env:"GITHUB_REF_PROTECTED"`
	RefType          string `env:"GITHUB_REF_TYPE"`
	RepoFullName     string `env:"GITHUB_REPOSITORY"`
	RepoID           int64  `env:"GITHUB_REPOSITORY_ID"`
	RepoOwner        string `env:"GITHUB_REPOSITORY_OWNER"`
	RepoOwnerID      int64  `env:"GITHUB_REPOSITORY_OWNER_ID"`
	RunAttempt       int64  `env:"GITHUB_RUN_ATTEMPT"`
	RunID            int64  `env:"GITHUB_RUN_ID"`
	RunNumber        int64  `env:"GITHUB_RUN_NUMBER"`
	RunnerArch       string `env:"RUNNER_ARCH"`
	RunnerDebug      bool   `env:"RUNNER_DEBUG"`
	RunnerOs         string `env:"RUNNER_OS"`
	SeverURL         string `env:"GITHUB_SERVER_URL"`
	TriggeringActor  string `env:"GITHUB_TRIGGERING_ACTOR"`
	Workflow         string `env:"GITHUB_WORKFLOW"`
	WorkflowRef      string `env:"GITHUB_WORKFLOW_REF"`
	WorkflowSha      string `env:"GITHUB_WORKFLOW_SHA"`
	Workspace        string `env:"GITHUB_WORKSPACE"`
}

func (i *Info) IsGitHubPullRequest() bool {
	return i.Num != 0 && i.Owner != "" && i.Repo != ""
}

// HasReadOnlyGitHubToken tells whether the current process is running in GitHub Actions on a GitHub PullRequest
// sent from a fork, with a read-only token.
//
// See https://docs.github.com/en/actions/reference/events-that-trigger-workflows#pull_request_target.
func (i *Info) HasReadOnlyGitHubToken() bool {
	return i.Fork && i.EventName == "pull_request_target"
}

// NewInfo creates a Info from environment variables.
func NewInfo() (*Info, error) {
	if IsRunningInGitHubAction() {
		return NewInfoFromGitHub()
	}

	// TODO: implement logic for other CI systems

	return nil, fmt.Errorf("CI systems other than GitHub Actions are not supported yet")
}

type Dumper interface {
	Dump() string
}

func (i *Info) Dump() string {
	ret := []string{}
	for _, f := range reflect.VisibleFields(reflect.TypeOf(*i)) {
		name := f.Name
		tag := f.Tag.Get("dump")
		if tag == "" {
			tag = f.Tag.Get("env")
			if tag == "" {
				tag = strcase.ToScreamingSnake(name)
			}
		}
		val := reflect.ValueOf(*i).FieldByName(name)
		if !val.IsValid() {
			continue
		}
		if isEmptyValue(val) {
			continue
		}
		ret = append(ret, fmt.Sprintf("%s=%v", tag, val))
	}
	sort.Strings(ret)

	return strings.Join(ret, "\n")
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Interface, reflect.Pointer:
		return v.IsZero()
	}

	return false
}
