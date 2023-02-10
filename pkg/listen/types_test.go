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

	"github.com/MakeNowJust/heredoc"
	"github.com/listendev/lstn/pkg/npm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNewAnalysisContext(t *testing.T) {
	analysisCtx1 := NewAnalysisContext()
	j1, e1 := json.Marshal(analysisCtx1)

	assert.NotNil(t, analysisCtx1)
	assert.Nil(t, analysisCtx1.Git)
	assert.NotEmpty(t, analysisCtx1.Version.Short)
	assert.NotEmpty(t, analysisCtx1.Version.Long)
	assert.Nil(t, e1)
	assert.NotContains(t, string(j1), "git")

	analysisCtx2 := NewAnalysisContext(func() (string, error) {
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
		pkgs    npm.Packages
		wantErr string
		version int
	}{
		{
			"both-nil",
			nil,
			nil,
			"couldn't create the analysis request",
			0,
		},
		{
			"lock-nil",
			nil,
			npm.Packages{
				"react": npm.Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			"couldn't create the analysis request",
			0,
		},
		{
			"pkgs-nil",
			npm.NewPackageLockJSON(),
			nil,
			"couldn't create the analysis request",
			0,
		},
		{
			"pkgs-empty",
			npm.NewPackageLockJSON(),
			npm.Packages{},
			"couldn't create the analysis request",
			0,
		},
		{
			"okish-but-invalid-lock",
			npm.NewPackageLockJSON(),
			npm.Packages{
				"react": npm.Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			"couldn't create the analysis request",
			0,
		},
		{
			"valid",
			validPackageLockJSON,
			npm.Packages{
				"@babel/runtime": npm.Package{Version: "7.20.13", Shasum: "7055ab8a7cff2b8f6058bf6ae45ff84ad2aded4b"},
			},
			"",
			3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := NewAnalysisRequest(tc.lock, tc.pkgs)
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

	validPackages := npm.Packages{
		"@babel/runtime": npm.Package{Version: "7.20.13", Shasum: "7055ab8a7cff2b8f6058bf6ae45ff84ad2aded4b"},
	}

	validAnalysisRequest, err := NewAnalysisRequest(validPackageLockJSON, validPackages)
	assert.Nil(t, err)

	validPackageLockBody := []byte(`"ewoJIm5hbWUiOiAic2FtcGxlIiwKCSJ2ZXJzaW9uIjogIjEuMC4wIiwKCSJsb2NrZmlsZVZlcnNpb24iOiAzLAoJInJlcXVpcmVzIjogdHJ1ZSwKCSJwYWNrYWdlcyI6IHsKCQkiIjogewoJCQkibmFtZSI6ICJzYW1wbGUiLAoJCQkidmVyc2lvbiI6ICIxLjAuMCIsCgkJCSJsaWNlbnNlIjogIklTQyIsCgkJCSJkZXBlbmRlbmNpZXMiOiB7CgkJCQkicmVhY3QiOiAiMTguMC4wIgoJCQl9CgkJfSwKCQkibm9kZV9tb2R1bGVzL0BiYWJlbC9ydW50aW1lIjogewoJCQkidmVyc2lvbiI6ICI3LjIwLjEzIiwKCQkJInJlc29sdmVkIjogImh0dHBzOi8vcmVnaXN0cnkubnBtanMub3JnL0BiYWJlbC9ydW50aW1lLy0vcnVudGltZS03LjIwLjEzLnRneiIsCgkJCSJpbnRlZ3JpdHkiOiAic2hhNTEyLWd0M1BLWHMwREJvTDl4Q3ZPSUlaMk5FcUFHWnFIakFubVZiZlF0QjYyMFYwdVJlSVF1dHBlbDE0S2NuZVp1ZXI3VWlvWThBTEtaN2lvY2F2dnpUTkZBPT0iLAoJCQkiZGVwZW5kZW5jaWVzIjogewoJCQkJInJlZ2VuZXJhdG9yLXJ1bnRpbWUiOiAiXjAuMTMuMTEiCgkJCX0sCgkJCSJlbmdpbmVzIjogewoJCQkJIm5vZGUiOiAiPj02LjkuMCIKCQkJfQoJCX0KCX0KfQ=="`)

	validPackagesBody := []byte(`{"@babel/runtime":{"version":"7.20.13","shasum":"7055ab8a7cff2b8f6058bf6ae45ff84ad2aded4b"}}`)

	tests := []struct {
		desc    string
		areq    *AnalysisRequest
		lock    []byte
		pkgs    []byte
		wantErr string
	}{
		{
			desc:    "valid",
			areq:    validAnalysisRequest,
			lock:    validPackageLockBody,
			pkgs:    validPackagesBody,
			wantErr: "",
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
				assert.Nil(t, err)

				var o map[string]json.RawMessage
				e := json.Unmarshal(res, &o)
				assert.Nil(t, e)

				assert.Equal(t, tc.lock, []byte(o["package-lock"]))
				assert.Equal(t, tc.pkgs, []byte(o["packages"]))
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

type expectedFactory = func() interface{}

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
			reader: strings.NewReader(`[{"name":"name","version":"version","shasum":"shasum","verdicts":[]}]`),
			expected: Response{
				Package{
					Name:     "name",
					Version:  "version",
					Shasum:   "shasum",
					Verdicts: []Verdict{},
				},
			},
		},
		{
			desc:   "with verdicts",
			reader: strings.NewReader(`[{"name":"name","version":"version","shasum":"shasum","verdicts":[{"message":"message","priority":"priority","metadata":{}}]}]`),
			expected: Response{
				Package{
					Name:    "name",
					Version: "version",
					Shasum:  "shasum",
					Verdicts: []Verdict{
						{
							Message:  "message",
							Priority: "priority",
							Metadata: make(map[string]interface{}),
						},
					},
				},
			},
		},
		{
			desc:   "metadata accept any type",
			reader: strings.NewReader(`[{"name":"name","version":"version","shasum":"shasum","verdicts":[{"message":"message","priority":"priority","metadata":{"number":42,"string":"foo"}}]}]`),
			expected: Response{
				Package{
					Name:    "name",
					Version: "version",
					Shasum:  "shasum",
					Verdicts: []Verdict{
						{
							Message:  "message",
							Priority: "priority",
							Metadata: metadata,
						},
					},
				},
			},
		},
		{
			desc:   "with problems",
			reader: strings.NewReader(`[{"name":"name","version":"version","shasum":"shasum","problems":[{"type":"type","title":"title","detail":"detail"}]}]`),
			expected: Response{
				Package{
					Name:    "name",
					Version: "version",
					Shasum:  "shasum",
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
