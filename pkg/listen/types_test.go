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
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/listendev/pkg/ecosystem"
	"github.com/listendev/pkg/models/category"
	"github.com/listendev/pkg/verdictcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNewVerdictsRequest(t *testing.T) {
	tests := []struct {
		desc    string
		args    []string
		req     *VerdictsRequest
		wantErr string
	}{
		{
			"no-args",
			[]string{},
			nil,
			"a verdicts request requires at least one argument (package name)",
		},
		{
			"name-only",
			[]string{"react"},
			&VerdictsRequest{Name: "react"},
			"",
		},
		{
			"name+version",
			[]string{"react", "18.2.0"},
			&VerdictsRequest{Name: "react", Version: "18.2.0"},
			"",
		},
		{
			"name+version+shasum",
			[]string{"react", "18.2.0", "555bd98592883255fa00de14f1151a917b5d77d5"},
			&VerdictsRequest{Name: "react", Version: "18.2.0", Digest: "555bd98592883255fa00de14f1151a917b5d77d5"},
			"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := NewVerdictsRequest(tc.args)
			if err != nil {
				assert.Nil(t, res)
				if assert.Error(t, err) {
					assert.Equal(t, tc.wantErr, err.Error())
				}
			} else {
				assert.Nil(t, err)
				assert.IsType(t, &VerdictsRequest{}, res)
				assert.Equal(t, tc.req.Name, res.Name)
				assert.Equal(t, tc.req.Version, res.Version)
				assert.Equal(t, tc.req.Digest, res.Digest)
			}
		})
	}
}

func TestNewVerdictsRequestWithContext(t *testing.T) {
	c := NewContext()

	tests := []struct {
		desc    string
		args    []string
		req     *VerdictsRequest
		wantErr string
	}{
		{
			"no-args",
			[]string{},
			nil,
			"a verdicts request requires at least one argument (package name)",
		},
		{
			"name-only",
			[]string{"react"},
			&VerdictsRequest{Name: "react", Context: c},
			"",
		},
		{
			"name+version",
			[]string{"react", "18.2.0"},
			&VerdictsRequest{Name: "react", Version: "18.2.0", Context: c},
			"",
		},
		{
			"name+version+shasum",
			[]string{"react", "18.2.0", "555bd98592883255fa00de14f1151a917b5d77d5"},
			&VerdictsRequest{Name: "react", Version: "18.2.0", Digest: "555bd98592883255fa00de14f1151a917b5d77d5", Context: c},
			"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := NewVerdictsRequestWithContext(tc.args, c)
			if err != nil {
				assert.Nil(t, res)
				if assert.Error(t, err) {
					assert.Equal(t, tc.wantErr, err.Error())
				}
			} else {
				assert.Nil(t, err)
				assert.IsType(t, &VerdictsRequest{}, res)
				assert.Equal(t, tc.req.Name, res.Name)
				assert.Equal(t, tc.req.Version, res.Version)
				assert.Equal(t, tc.req.Digest, res.Digest)
				assert.Equal(t, tc.req.Context, res.Context)
			}
		})
	}
}

func TestNewBulkVerdictsRequestsFromMap(t *testing.T) {
	tests := []struct {
		desc    string
		args    map[string]*semver.Version
		reqs    []*VerdictsRequest
		wantErr string
	}{
		{
			"empty",
			map[string]*semver.Version{},
			nil,
			"couldn't create a request set from empty dependencies map",
		},
		{
			"one-package-without-version",
			map[string]*semver.Version{
				"tap": nil,
			},
			[]*VerdictsRequest{
				{Name: "tap"},
			},
			"",
		},
		{
			"one-package-without-empty-name",
			map[string]*semver.Version{
				"": nil,
			},
			nil,
			"a verdicts request requires at least one argument (package name)",
		},
		{
			"one-package-with-version",
			map[string]*semver.Version{
				"tap": semver.MustParse("15.1.2"),
			},
			[]*VerdictsRequest{
				{Name: "tap", Version: "15.1.2"},
			},
			"",
		},
		{
			"more-packages-with-version",
			map[string]*semver.Version{
				"tap":   semver.MustParse("15.1.2"),
				"react": semver.MustParse("18.2.0"),
			},
			[]*VerdictsRequest{
				{Name: "tap", Version: "15.1.2"},
				{Name: "react", Version: "18.2.0"},
			},
			"",
		},
		{
			"more-packages-with-or-without-versions",
			map[string]*semver.Version{
				"tap":     semver.MustParse("15.1.2"),
				"core-js": semver.MustParse("3.25.0"),
				"react":   nil,
			},
			[]*VerdictsRequest{
				{Name: "tap", Version: "15.1.2"},
				{Name: "core-js", Version: "3.25.0"},
				{Name: "react"},
			},
			"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := NewBulkVerdictsRequestsFromMap(tc.args, "")
			if err != nil {
				assert.Nil(t, res)
				if assert.Error(t, err) {
					assert.Equal(t, tc.wantErr, err.Error())
				}
			} else {
				assert.Nil(t, err)
				assert.IsType(t, []*VerdictsRequest{}, res)
				if diff := cmp.Diff(tc.reqs, res, cmpopts.SortSlices(func(x, y *VerdictsRequest) bool {
					return x.Name < y.Name
				}), cmpopts.IgnoreFields(VerdictsRequest{}, "Context")); diff != "" {
					t.Errorf("verdicts request mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestNewContext(t *testing.T) {
	analysisCtx1 := NewContext()
	j1, e1 := json.Marshal(analysisCtx1)

	assert.NotNil(t, analysisCtx1)
	assert.Nil(t, analysisCtx1.Git)
	assert.NotEmpty(t, analysisCtx1.Version.Short)
	assert.NotEmpty(t, analysisCtx1.Version.Long)
	assert.Nil(t, e1)
	assert.NotContains(t, string(j1), "git")

	analysisCtx2 := NewContext(func() (string, error) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return path.Join(cwd, "../../"), nil
	})
	j2, e2 := json.Marshal(analysisCtx2)

	assert.NotNil(t, analysisCtx2)
	assert.NotNil(t, analysisCtx2.Git)
	assert.NotEmpty(t, analysisCtx1.Version.Short)
	assert.NotEmpty(t, analysisCtx1.Version.Long)
	assert.Nil(t, e2)
	assert.Contains(t, string(j2), "git")
}

func TestNewAnalysisRequest(t *testing.T) {
	validPackageLockJSON, _ := npm.NewPackageLockJSONFromBytes([]byte(heredoc.Doc(`{
		"name": "sample",
		"version": "1.0.0",
		"lockfileVersion": 3,
		"requires": true,
		"packages": {
			"": {
				"name": "sample",
				"version": "1.0.0",
				"license": "ISC",
				"dependencies": {
					"react": "18.0.0"
				}
			},
			"node_modules/@babel/runtime": {
				"version": "7.20.13",
				"resolved": "https://registry.npmjs.org/@babel/runtime/-/runtime-7.20.13.tgz",
				"integrity": "sha512-gt3PKXs0DBoL9xCvOIIZ2NEqAGZqHjAnmVbfQtB620V0uReIQutpel14KcneZuer7UioY8ALKZ7iocavvzTNFA==",
				"dependencies": {
					"regenerator-runtime": "^0.13.11"
				},
				"engines": {
					"node": ">=6.9.0"
				}
			}
		}
	}`)))

	tests := []struct {
		desc    string
		lock    npm.PackageLockJSON
		wantErr string
		version int
	}{
		{
			"lock-nil",
			nil,
			"couldn't create the analysis request",
			0,
		},
		{
			"lock-empty",
			npm.NewPackageLockJSON(),
			"couldn't create the analysis request because of invalid package-lock.json",
			0,
		},
		{
			"valid",
			validPackageLockJSON,
			"",
			3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := NewAnalysisRequest(tc.lock, WithRequestContext())
			if err != nil {
				assert.Nil(t, res)
				if assert.Error(t, err) {
					assert.Equal(t, tc.wantErr, err.Error())
				}
			} else {
				assert.Nil(t, err)
				assert.IsType(t, &AnalysisRequest{}, res)
				assert.Equal(t, res.PackageLockJSON.Version(), tc.version)
			}
		})
	}
}

func TestAnalysisRequestMarshal(t *testing.T) {
	validPackageLockJSON1, _ := npm.NewPackageLockJSONFromBytes([]byte(`{"name": "github-gist","version": "1.0.0","lockfileVersion": 1,"requires": true,"dependencies": {"@babel/runtime-corejs3": {"version": "7.12.1","resolved": "https://registry.npmjs.org/@babel/runtime-corejs3/-/runtime-corejs3-7.12.1.tgz","integrity": "sha512-umhPIcMrlBZ2aTWlWjUseW9LjQKxi1dpFlQS8DzsxB//5K+u6GLTC/JliPKHsd5kJVPIU6X/Hy0YvWOYPcMxBw==","dev": true,"requires": {"core-js-pure": "^3.0.0","regenerator-runtime": "^0.13.4"}}}}`))

	validAnalysisRequest1, err := NewAnalysisRequest(validPackageLockJSON1, WithRequestContext())
	assert.Nil(t, err)

	validPackageLockBody1 := []byte(`"eyJuYW1lIjogImdpdGh1Yi1naXN0IiwidmVyc2lvbiI6ICIxLjAuMCIsImxvY2tmaWxlVmVyc2lvbiI6IDEsInJlcXVpcmVzIjogdHJ1ZSwiZGVwZW5kZW5jaWVzIjogeyJAYmFiZWwvcnVudGltZS1jb3JlanMzIjogeyJ2ZXJzaW9uIjogIjcuMTIuMSIsInJlc29sdmVkIjogImh0dHBzOi8vcmVnaXN0cnkubnBtanMub3JnL0BiYWJlbC9ydW50aW1lLWNvcmVqczMvLS9ydW50aW1lLWNvcmVqczMtNy4xMi4xLnRneiIsImludGVncml0eSI6ICJzaGE1MTItdW1oUEljTXJsQloyYVRXbFdqVXNlVzlMalFLeGkxZHBGbFFTOER6c3hCLy81Syt1NkdMVEMvSmxpUEtIc2Q1a0pWUElVNlgvSHkwWXZXT1lQY014Qnc9PSIsImRldiI6IHRydWUsInJlcXVpcmVzIjogeyJjb3JlLWpzLXB1cmUiOiAiXjMuMC4wIiwicmVnZW5lcmF0b3ItcnVudGltZSI6ICJeMC4xMy40In19fX0="`)

	validPackageLockJSON3, _ := npm.NewPackageLockJSONFromBytes([]byte(heredoc.Doc(`{
		"name": "sample",
		"version": "1.0.0",
		"lockfileVersion": 3,
		"requires": true,
		"packages": {
			"": {
				"name": "sample",
				"version": "1.0.0",
				"license": "ISC",
				"dependencies": {
					"react": "18.0.0"
				}
			},
			"node_modules/@babel/runtime": {
				"version": "7.20.13",
				"resolved": "https://registry.npmjs.org/@babel/runtime/-/runtime-7.20.13.tgz",
				"integrity": "sha512-gt3PKXs0DBoL9xCvOIIZ2NEqAGZqHjAnmVbfQtB620V0uReIQutpel14KcneZuer7UioY8ALKZ7iocavvzTNFA==",
				"dependencies": {
					"regenerator-runtime": "^0.13.11"
				},
				"engines": {
					"node": ">=6.9.0"
				}
			}
		}
	}`)))

	validAnalysisRequest3, err := NewAnalysisRequest(validPackageLockJSON3, WithRequestContext())
	assert.Nil(t, err)

	validPackageLockBody3 := []byte(`"ewoJIm5hbWUiOiAic2FtcGxlIiwKCSJ2ZXJzaW9uIjogIjEuMC4wIiwKCSJsb2NrZmlsZVZlcnNpb24iOiAzLAoJInJlcXVpcmVzIjogdHJ1ZSwKCSJwYWNrYWdlcyI6IHsKCQkiIjogewoJCQkibmFtZSI6ICJzYW1wbGUiLAoJCQkidmVyc2lvbiI6ICIxLjAuMCIsCgkJCSJsaWNlbnNlIjogIklTQyIsCgkJCSJkZXBlbmRlbmNpZXMiOiB7CgkJCQkicmVhY3QiOiAiMTguMC4wIgoJCQl9CgkJfSwKCQkibm9kZV9tb2R1bGVzL0BiYWJlbC9ydW50aW1lIjogewoJCQkidmVyc2lvbiI6ICI3LjIwLjEzIiwKCQkJInJlc29sdmVkIjogImh0dHBzOi8vcmVnaXN0cnkubnBtanMub3JnL0BiYWJlbC9ydW50aW1lLy0vcnVudGltZS03LjIwLjEzLnRneiIsCgkJCSJpbnRlZ3JpdHkiOiAic2hhNTEyLWd0M1BLWHMwREJvTDl4Q3ZPSUlaMk5FcUFHWnFIakFubVZiZlF0QjYyMFYwdVJlSVF1dHBlbDE0S2NuZVp1ZXI3VWlvWThBTEtaN2lvY2F2dnpUTkZBPT0iLAoJCQkiZGVwZW5kZW5jaWVzIjogewoJCQkJInJlZ2VuZXJhdG9yLXJ1bnRpbWUiOiAiXjAuMTMuMTEiCgkJCX0sCgkJCSJlbmdpbmVzIjogewoJCQkJIm5vZGUiOiAiPj02LjkuMCIKCQkJfQoJCX0KCX0KfQ=="`)

	tests := []struct {
		desc    string
		areq    *AnalysisRequest
		lock    []byte
		pkgs    []byte
		wantErr string
	}{
		{
			desc:    "valid version 1",
			areq:    validAnalysisRequest1,
			lock:    validPackageLockBody1,
			wantErr: "",
		},
		{
			desc:    "valid version 3",
			areq:    validAnalysisRequest3,
			lock:    validPackageLockBody3,
			wantErr: "",
		},
		{
			desc:    "missing-packagelock",
			areq:    &AnalysisRequest{},
			wantErr: "json: error calling MarshalJSON for type *listen.AnalysisRequest: package lock is mandatory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := json.Marshal(tc.areq)
			if err != nil {
				assert.Nil(t, res)
				if assert.Error(t, err) {
					assert.Equal(t, tc.wantErr, err.Error())
				}
			} else {
				var o map[string]json.RawMessage
				e := json.Unmarshal(res, &o)
				assert.Nil(t, e)

				assert.Equal(t, tc.lock, []byte(o["manifest"]))
			}
		})
	}
}

type TypesSuite struct {
	suite.Suite
}

func TestTypesSuite(t *testing.T) {
	suite.Run(t, new(TypesSuite))
}

func (suite *TypesSuite) TestResponseMarshalJSON() {
	metadata := make(map[string]interface{})
	metadata["number"] = float64(42)
	metadata["string"] = "foo"

	t := suite.T()
	cases := []struct {
		desc     string
		reader   io.Reader
		expected Response
	}{
		{
			desc:     "empty list",
			reader:   strings.NewReader(`[]`),
			expected: Response{},
		},
		{
			desc:   "without verdicts",
			reader: strings.NewReader(`[{"name":"name","version":"version","digest":"shasum","verdicts":[]}]`),
			expected: Response{
				Package{
					Name:     "name",
					Version:  strPtr("version"),
					Digest:   strPtr("shasum"),
					Verdicts: []Verdict{},
				},
			},
		},
		{
			desc:   "with verdicts",
			reader: strings.NewReader(`[{"name":"name","version":"1.0.0","digest":"036bfebd9748309772b79753d2e9924af7707d00","verdicts":[{"pkg": "name", "version": "1.0.0", "digest": "036bfebd9748309772b79753d2e9924af7707d00", "created_at": "2023-06-22T20:12:58.911537+00:00", "categories": ["process"], "code": "STN001", "fingerprint": "fp0001", "file": "static(exfiltrate_env).json", "ecosystem": "npm", "message":"message","severity":"medium","metadata":{}}]}]`),
			expected: Response{
				Package{
					Name:    "name",
					Version: strPtr("1.0.0"),
					Digest:  strPtr("036bfebd9748309772b79753d2e9924af7707d00"),
					Verdicts: []Verdict{
						{
							File:        "static(exfiltrate_env).json",
							Pkg:         "name",
							Digest:      "036bfebd9748309772b79753d2e9924af7707d00",
							Version:     "1.0.0",
							Categories:  []category.Category{category.Process},
							Code:        verdictcode.STN001,
							Fingerprint: "fp0001",
							Ecosystem:   ecosystem.Npm,
							Message:     "message",
							Severity:    "medium",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Metadata: make(map[string]interface{}),
						},
					},
				},
			},
		},
		{
			desc:   "metadata accept any type",
			reader: strings.NewReader(`[{"name":"name","version":"1.0.0","digest":"036bfebd9748309772b79753d2e9924af7707d00","verdicts":[{"pkg": "name", "version": "1.0.0", "digest": "036bfebd9748309772b79753d2e9924af7707d00", "created_at": "2023-06-22T20:12:58.911537+00:00", "categories": ["process"], "fingerprint": "fp0001", "code": "STN001", "file": "static(exfiltrate_env).json", "ecosystem": "npm", "message":"message","severity":"medium","metadata":{"number":42,"string":"foo"}}]}]`),
			expected: Response{
				Package{
					Name:    "name",
					Version: strPtr("1.0.0"),
					Digest:  strPtr("036bfebd9748309772b79753d2e9924af7707d00"),
					Verdicts: []Verdict{
						{
							File:        "static(exfiltrate_env).json",
							Pkg:         "name",
							Digest:      "036bfebd9748309772b79753d2e9924af7707d00",
							Version:     "1.0.0",
							Categories:  []category.Category{category.Process},
							Fingerprint: "fp0001",
							Code:        verdictcode.STN001,
							Ecosystem:   ecosystem.Npm,
							Message:     "message",
							Severity:    "medium",
							CreatedAt: func() *time.Time {
								t, _ := time.Parse(time.RFC3339Nano, "2023-06-22T20:12:58.911537+00:00")

								return &t
							}(),
							Metadata: metadata,
						},
					},
				},
			},
		},
		{
			desc:   "with problems",
			reader: strings.NewReader(`[{"name":"name","version":"version","digest":"shasum","problems":[{"type":"type","title":"title","detail":"detail"}]}]`),
			expected: Response{
				Package{
					Name:    "name",
					Version: strPtr("version"),
					Digest:  strPtr("shasum"),
					Problems: []Problem{
						{
							Type:   "type",
							Title:  "title",
							Detail: "detail",
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		r := new(Response)
		t.Run(tc.desc, func(t *testing.T) {
			dec := json.NewDecoder(tc.reader)
			suite.NoError(dec.Decode(r))
			suite.Equal(*r, tc.expected)
		})
	}
}
