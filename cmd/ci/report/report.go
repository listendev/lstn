// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2025 The listen.dev team <engineering@garnet.ai>
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
package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/listendev/lstn/internal/project"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/lstn/pkg/cmd/options"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/reporter/factory"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/spf13/cobra"
)

var _, filename, _, _ = runtime.Caller(0)

func New(ctx context.Context) (*cobra.Command, error) {
	c := &cobra.Command{
		Use:                   "report",
		DisableFlagsInUseLine: true,
		Short:                 "Report the most critical findings into GitHub pull requests",
		Annotations: map[string]string{
			"source": project.GetSourceURL(filename),
		},
		RunE: func(c *cobra.Command, _ []string) error {
			ctx = c.Context()
			// Obtain the local options from the context
			optsFromContext, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.CiReportKey)
			if err != nil {
				return err
			}
			opts, ok := optsFromContext.(*options.CiReport)
			if !ok {
				return fmt.Errorf("couldn't obtain options for the current child command")
			}

			// Token options are mandatory in this case
			errs := []error{}
			// GitHub token is mandatory for reporting (posting the comment)
			if err := validate.Singleton.Var(opts.Token.GitHub, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(opts, "ConfigFlags.Token.GitHub")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}
			// The listen.dev token is mandatory for fetching the data to report
			if err := validate.Singleton.Var(opts.Token.JWT, "mandatory"); err != nil {
				tags, _ := flags.GetFieldTag(opts, "ConfigFlags.Token.JWT")
				name, _ := tags.Lookup("name")
				errs = append(errs, flags.Translate(err, name)...)
			}
			if len(errs) > 0 {
				ret := "invalid configuration options/flags"
				for _, e := range errs {
					ret += "\n       "
					ret += e.Error()
				}

				return fmt.Errorf("%s", ret)
			}

			if opts.DebugOptions {
				c.Println(opts.AsJSON())

				return nil
			}
			source := "eavesdrop tool"
			reportingOpts := flags.Reporting{Types: []cmd.ReportType{cmd.GitHubPullCommentReport}}

			_, infoErr := ci.NewInfo()
			if infoErr != nil {
				return factory.ErrReporterUnsupportedEnvironment
			}

			githubRepositoryID := os.Getenv("GITHUB_REPOSITORY_ID")
			githubWorkflow := os.Getenv("GITHUB_WORKFLOW")
			githubJob := os.Getenv("GITHUB_JOB")
			githubRunID := os.Getenv("GITHUB_RUN_ID")
			githubRunNumber := os.Getenv("GITHUB_RUN_NUMBER")
			githubRunAttempt := os.Getenv("GITHUB_RUN_ATTEMPT")

			ghCxt := GHctx{
				RepositoryID: githubRepositoryID,
				Workflow:     githubWorkflow,
				Job:          githubJob,
				RunID:        githubRunID,
				RunNumber:    githubRunNumber,
				RunAttempt:   githubRunAttempt,
			}

			evts, err := events(c.Context(), opts.Endpoint.Core, opts.Token.JWT, ghCxt)
			if err != nil {
				return err
			}

			summaries := summary(evts)

			if len(summaries) == 0 { // no dangerous events, so no need to report
				fmt.Println("No dangerous network events found.")

				commentBody := "✅ **listen.dev** runtime monitor executed successfully! 🎉\n\n"
				commentBody += "No security issues were detected during the scan.\n\n"

				return factory.Exec(c, reportingOpts, heredoc.Doc(commentBody), &source)
			}

			if err := triggerWebhook(c.Context(), opts.Token.JWT, opts.Endpoint.Core, ghCxt); err != nil {
				c.Println("Failed to trigger the webhook")

				// NOTE: We don't return here because we still want to report the findings in GH PR
			}

			link, err := getLinkOfDashboard(c.Context(), opts.Endpoint.Core, opts.Token.JWT, ghCxt)
			if err != nil {
				c.Println("Failed to get the link of the dashboard")

				return err
			}

			commentBody := "# ⚠️ **listen.dev** runtime monitor detected a potential security issue\n"

			domains := "domain"
			if len(summaries) > 1 {
				domains = "domains"
			}

			commentBody += fmt.Sprintf("Suspicious %s detected", domains)
			commentBody += "\n"

			// Start the markdown table
			commentBody += "| Domain        | Status   |\n"
			commentBody += "|--------------|----------|\n"

			for _, s := range summaries {
				commentBody += fmt.Sprintf("| %s | 🚫 BLOCKED |\n", s)
			}

			commentBody += fmt.Sprintln("> These connections were automatically blocked by the runtime monitor to protect your workflows.")
			commentBody += fmt.Sprintf("> [Review and manage these issues in listen.dev dashboard](%s)", link)

			return factory.Exec(c, reportingOpts, heredoc.Doc(commentBody), &source)
		},
	}

	// Create the local options
	reportOpts, err := options.NewCiReport()
	if err != nil {
		return nil, err
	}
	// Local flags will only run when this command is called directly
	reportOpts.Attach(c, []string{"npm-registry", "select", "ignore-deptypes", "ignore-packages", "pypi-endpoint", "npm-endpoint", "lockfiles", "reporter"})

	// Pass the options through the context
	ctx = context.WithValue(ctx, pkgcontext.CiReportKey, reportOpts)
	c.SetContext(ctx)

	return c, nil
}

type GHctx struct {
	RepositoryID string `json:"repository_id"`
	Workflow     string `json:"workflow"`
	Job          string `json:"job"`
	RunID        string `json:"run_id"`
	RunNumber    string `json:"run_number"`
	RunAttempt   string `json:"run_attempt"`
}

// ContextElement defines model for ContextElement.
type ContextElement struct {
	Data any    `json:"data"`
	Type string `json:"type"`
}

type NetPolicyEvent struct {
	Data *struct {
		Body     *DataBody `json:"body,omitempty"`
		UniqueID *string   `json:"unique_id,omitempty"`
	} `json:"data,omitempty"`
	GithubContext *GitHubContext     `json:"github_context,omitempty"`
	ProjectID     *string            `bson:"project_id"               json:"project_id,omitempty"`
	Tags          *map[string]string `json:"tags,omitempty"`
	Type          *string            `json:"type,omitempty"`
}

type DataBody struct {
	Dropped     *DroppedIP   `json:"dropped,omitempty"`
	FullInfo    *FullInfo    `json:"full_info,omitempty"`
	Parent      *ProcessInfo `json:"parent,omitempty"`
	Process     *ProcessInfo `json:"process,omitempty"`
	Resolve     *string      `json:"resolve"`
	ResolveFlow *ResolveFlow `json:"resolve_flow,omitempty"`
}

type DroppedIP struct {
	Icmp *struct {
		Code *string `json:"code,omitempty"`
		Type *string `json:"type,omitempty"`
	} `json:"icmp,omitempty"`
	IPVersion   *int            `json:"ip_version,omitempty"`
	Local       *NetworkInfo    `json:"local,omitempty"`
	Properties  *FlowProperties `json:"properties,omitempty"`
	Proto       *string         `json:"proto,omitempty"`
	Remote      *NetworkInfo    `json:"remote,omitempty"`
	ServicePort *int            `json:"service_port,omitempty"`
}

type FlowProperties struct {
	Egress     *bool `json:"egress,omitempty"`
	Ended      *bool `json:"ended,omitempty"`
	Incoming   *bool `json:"incoming,omitempty"`
	Ingress    *bool `json:"ingress,omitempty"`
	Ongoing    *bool `json:"ongoing,omitempty"`
	Outgoing   *bool `json:"outgoing,omitempty"`
	Started    *bool `json:"started,omitempty"`
	Terminated *bool `json:"terminated,omitempty"`
	Terminator *bool `json:"terminator,omitempty"`
}

type NetworkInfo struct {
	Address *string   `json:"address,omitempty"`
	Name    *string   `json:"name,omitempty"`
	Names   *[]string `json:"names,omitempty"`
	Port    *int      `json:"port,omitempty"`
}

type FullInfo struct {
	Ancestry *[]AncestryInfo         `json:"ancestry,omitempty"`
	Files    *map[string]interface{} `json:"files,omitempty"`
	Flows    *[]ResolveFlow          `json:"flows,omitempty"`
}

type ResolveFlow struct {
	Icmp *struct {
		Code *string `json:"code,omitempty"`
		Type *string `json:"type,omitempty"`
	} `json:"icmp,omitempty"`
	IPVersion   *int            `json:"ip_version,omitempty"`
	Local       *NetworkInfo    `json:"local,omitempty"`
	Properties  *FlowProperties `json:"properties,omitempty"`
	Proto       *string         `json:"proto,omitempty"`
	Remote      *NetworkInfo    `json:"remote,omitempty"`
	ServicePort *int            `json:"service_port,omitempty"`
}

type AncestryInfo struct {
	Args    *string    `json:"args,omitempty"`
	Cmd     *string    `json:"cmd,omitempty"`
	Comm    *string    `json:"comm,omitempty"`
	Exe     *string    `json:"exe,omitempty"`
	Exit    *string    `json:"exit,omitempty"`
	PID     *int       `json:"pid,omitempty"`
	PpID    *int       `json:"ppid,omitempty"`
	Retcode *int       `json:"retcode,omitempty"`
	Start   *time.Time `json:"start,omitempty"`
	UID     *int       `json:"uid,omitempty"`
}

type ProcessInfo struct {
	Args    *string    `json:"args,omitempty"`
	Cmd     *string    `json:"cmd,omitempty"`
	Comm    *string    `json:"comm,omitempty"`
	Exe     *string    `json:"exe,omitempty"`
	Exit    *string    `json:"exit,omitempty"`
	PID     *int       `json:"pid,omitempty"`
	PpID    *int       `json:"ppid,omitempty"`
	Retcode *int       `json:"retcode,omitempty"`
	Start   *time.Time `json:"start,omitempty"`
	UID     *int       `json:"uid,omitempty"`
}

type GitHubContext struct {
	Action            *string `json:"action,omitempty"`
	Actor             *string `json:"actor,omitempty"`
	ActorID           *string `json:"actor_id,omitempty"`
	EventName         *string `json:"event_name,omitempty"`
	Job               *string `json:"job,omitempty"`
	Ref               *string `json:"ref,omitempty"`
	RefName           *string `json:"ref_name,omitempty"`
	RefProtected      *bool   `json:"ref_protected,omitempty"`
	RefType           *string `json:"ref_type,omitempty"`
	Repository        *string `json:"repository,omitempty"`
	RepositoryID      *string `json:"repository_id,omitempty"`
	RepositoryOwner   *string `json:"repository_owner,omitempty"`
	RepositoryOwnerID *string `json:"repository_owner_id,omitempty"`
	RunAttempt        *string `json:"run_attempt,omitempty"`
	RunID             *string `json:"run_id,omitempty"`
	RunNumber         *string `json:"run_number,omitempty"`
	RunnerArch        *string `json:"runner_arch,omitempty"`
	RunnerOs          *string `json:"runner_os,omitempty"`
	ServerURL         *string `json:"server_url,omitempty"`
	Sha               *string `json:"sha,omitempty"`
	TriggeringActor   *string `json:"triggering_actor,omitempty"`
	Workflow          *string `json:"workflow,omitempty"`
	WorkflowRef       *string `json:"workflow_ref,omitempty"`
	WorkflowSha       *string `json:"workflow_sha,omitempty"`
	Workspace         *string `json:"workspace,omitempty"`
}

func getLinkOfDashboard(ctx context.Context, baseURL, token string, ghCtx GHctx) (string, error) {
	url := baseURL + "/api/v1/dashboard/link"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/json")

	qp := request.URL.Query()
	qp.Add("repository_id", ghCtx.RepositoryID)
	qp.Add("workflow", ghCtx.Workflow)
	qp.Add("job", ghCtx.Job)
	qp.Add("run_id", ghCtx.RunID)
	qp.Add("run_number", ghCtx.RunNumber)
	qp.Add("run_attempt", ghCtx.RunAttempt)
	request.URL.RawQuery = qp.Encode()

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Link string `json:"link"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	return result.Link, nil
}

func events(ctx context.Context, baseURL, token string, ghCtx GHctx) ([]NetPolicyEvent, error) {
	url := baseURL + "/api/v1/network_events"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/json")

	qp := request.URL.Query()
	qp.Add("repository_id", ghCtx.RepositoryID)
	qp.Add("workflow", ghCtx.Workflow)
	qp.Add("job", ghCtx.Job)
	qp.Add("run_id", ghCtx.RunID)
	qp.Add("run_number", ghCtx.RunNumber)
	qp.Add("run_attempt", ghCtx.RunAttempt)
	request.URL.RawQuery = qp.Encode()

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var events []NetPolicyEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}

	return events, nil
}

func domainName(event *NetPolicyEvent) (string, error) {
	if event != nil &&
		event.Data != nil &&
		event.Data.Body != nil &&
		event.Data.Body.Dropped != nil &&
		event.Data.Body.Dropped.Remote != nil &&
		event.Data.Body.Dropped.Remote.Name != nil {
		return *event.Data.Body.Dropped.Remote.Name, nil
	}

	if event != nil &&
		event.Data != nil &&
		event.Data.Body != nil &&
		event.Data.Body.Resolve != nil {
		return *event.Data.Body.Resolve, nil
	}

	return "", fmt.Errorf("domain name not found in event")
}

func summary(events []NetPolicyEvent) []string {
	summaries := make([]string, 0, len(events))

	for _, e := range events {
		domain, err := domainName(&e)
		if err != nil {
			fmt.Println("cannot find domain name in event", e.Data.UniqueID)

			continue
		}

		summaries = append(summaries, domain)
	}

	return summaries
}

func triggerWebhook(ctx context.Context, token string, baseURL string, ghCtx GHctx) error {
	url := baseURL + "/api/v1/webhook"

	data, err := json.Marshal(ghCtx)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	// response can be 202 or 204
	if response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}
