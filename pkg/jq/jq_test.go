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
package jq

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	t.Setenv("MY_DEP", "postcss-clean")

	type arguments struct {
		json io.Reader
		expr string
	}

	tests := []struct {
		name  string
		args  arguments
		wantW string
		wantE bool
		typeE error
	}{
		{
			name: "invalid-query",
			args: arguments{
				json: strings.NewReader(`{}`),
				expr: `.[] | select(.name == "ciao"`,
			},
			wantE: true,
		},
		{
			name: "empty-input",
			args: arguments{
				json: strings.NewReader(``),
				expr: `.[]`,
			},
			wantE: true,
		},
		{
			name: "with-env-var",
			args: arguments{
				json: strings.NewReader(heredoc.Doc(`[
					{"name":"postcss-clean","shasum":"be7fca7b640b6d6d5326f1a6bfe50028174090b3","verdicts":[],"version":"1.0.0"},
					{"name":"chalk","shasum":"cd42541677a54333cf541a49108c1432b44c9424","verdicts":[],"version":"2.4.2"}
				]`)),
				expr: `.[] | select(.name == env.MY_DEP) | .version`,
			},
			wantW: "1.0.0\n",
		},
		{
			name: "halt",
			args: arguments{
				json: strings.NewReader(heredoc.Doc(`[
					{"name":"chalk","shasum":"cd42541677a54333cf541a49108c1432b44c9424","verdicts":[],"version":"2.4.2"},
					{"name":"chalk","shasum":"d957f370038b75ac572471e83be4c5ca9f8e8c45","verdicts":[],"version":"5.1.2"}
				]`)),
				expr: `halt_error`,
			},
			wantE: true,
			typeE: &HaltError{
				5,
				`[{"name":"chalk","shasum":"cd42541677a54333cf541a49108c1432b44c9424","verdicts":[],"version":"2.4.2"},{"name":"chalk","shasum":"d957f370038b75ac572471e83be4c5ca9f8e8c45","verdicts":[],"version":"5.1.2"}]`,
			},
		},
		{
			name: "halt-5",
			args: arguments{
				json: strings.NewReader(heredoc.Doc(`[
					{"name":"chalk","shasum":"cd42541677a54333cf541a49108c1432b44c9424","verdicts":[],"version":"2.4.2"},
					{"name":"chalk","shasum":"c4b4a62bfb6df0eeeb5dbc52e6a9ecaff14b9976","verdicts":[],"version":"5.1.0"},
					{"name":"chalk","shasum":"546fb2b8fc5b7dad0991ef81b2a98b265fa71e02","verdicts":[],"version":"5.1.1"},
					{"name":"chalk","shasum":"d957f370038b75ac572471e83be4c5ca9f8e8c45","verdicts":[],"version":"5.1.2"}
				]`)),
				expr: `.[] | halt_error`,
			},
			wantE: true,
			typeE: &HaltError{
				5,
				`{"name":"chalk","shasum":"cd42541677a54333cf541a49108c1432b44c9424","verdicts":[],"version":"2.4.2"}`,
			},
		},
		{
			name: "halt-22",
			args: arguments{
				json: strings.NewReader(heredoc.Doc(`[
					{"name":"chalk","shasum":"cd42541677a54333cf541a49108c1432b44c9424","verdicts":[],"version":"2.4.2"},
					{"name":"chalk","shasum":"c4b4a62bfb6df0eeeb5dbc52e6a9ecaff14b9976","verdicts":[],"version":"5.1.0"},
					{"name":"chalk","shasum":"546fb2b8fc5b7dad0991ef81b2a98b265fa71e02","verdicts":[],"version":"5.1.1"},
					{"name":"chalk","shasum":"d957f370038b75ac572471e83be4c5ca9f8e8c45","verdicts":[],"version":"5.1.2"}
				]`)),
				expr: `.[] | halt_error(22)`,
			},
			wantE: true,
			typeE: &HaltError{
				22,
				`{"name":"chalk","shasum":"cd42541677a54333cf541a49108c1432b44c9424","verdicts":[],"version":"2.4.2"}`,
			},
		},
		{
			name: "object-simple",
			args: arguments{
				json: strings.NewReader(`{"name":"postcss-clean", "version": "1.2.2"}`),
				expr: `.name`,
			},
			wantW: "postcss-clean\n",
		},
		{
			name: "object-simple-halt",
			args: arguments{
				json: strings.NewReader(`{"name":"postcss-clean", "version": "1.2.2"}`),
				expr: `.name | halt_error`,
			},
			wantE: true,
			typeE: &HaltError{
				5,
				`postcss-clean`,
			},
		},
		{
			name: "object-int-simple",
			args: arguments{
				json: strings.NewReader(`{"name":"postcss-clean", "integer": 22}`),
				expr: `.integer`,
			},
			wantW: "22\n",
		},
		{
			name: "object-bool-simple",
			args: arguments{
				json: strings.NewReader(`{"name":"postcss-clean", "use": true}`),
				expr: `.use`,
			},
			wantW: "true\n",
		},
		{
			name: "object-multiple",
			args: arguments{
				json: strings.NewReader(`{"name":"postcss-clean", "version": "1.2.2"}`),
				expr: `.name,.version`,
			},
			wantW: "postcss-clean\n1.2.2\n",
		},
		{
			name: "object-json",
			args: arguments{
				json: strings.NewReader(heredoc.Doc(`{
					"name": "chalk",
					"version": "5.1.3",
					"shasum": "d957f370038b75ac572471e83be4c5ca9f8e8c45",
					"verdicts": [
						{
							"message": "npm install spawned a process",
							"severity": "medium",
							"metadata": {
								"npm_package_name": "@coreui/react",
								"npm_package_version": "3.4.6",
								"commandline": "sh -c node npm-postinstall",
								"parent_name": "node"
							}
						},
						{
							"message": "npm install spawned a process",
							"severity": "medium",
							"metadata": {
								"npm_package_name": "core-js",
								"npm_package_version": "3.26.1",
								"commandline": "sh -c node -e \"try{require('./postinstall')}catch(e){}\"",
								"parent_name": "node"
							}
						}
					]
				}`)),
				expr: `.verdicts[0]`,
			},
			wantW: "{\"message\":\"npm install spawned a process\",\"metadata\":{\"commandline\":\"sh -c node npm-postinstall\",\"npm_package_name\":\"@coreui/react\",\"npm_package_version\":\"3.4.6\",\"parent_name\":\"node\"},\"severity\":\"medium\"}\n",
		},
		{
			name: "complex-1",
			args: arguments{
				json: strings.NewReader(heredoc.Doc(`[
					{
						"name": "chalk",
						"version": "5.1.3",
						"shasum": "d957f370038b75ac572471e83be4c5ca9f8e8c45",
						"verdicts": [
							{
								"message": "npm install spawned a process",
								"severity": "medium",
								"metadata": {
									"npm_package_name": "@coreui/react",
									"npm_package_version": "3.4.6",
									"commandline": "sh -c node npm-postinstall",
									"parent_name": "node"
								}
							},
							{
								"message": "npm install spawned a process",
								"severity": "medium",
								"metadata": {
									"npm_package_name": "core-js",
									"npm_package_version": "3.26.1",
									"commandline": "sh -c node -e \"try{require('./postinstall')}catch(e){}\"",
									"parent_name": "node"
								}
							}
						]
					}
				]`)),
				expr: `.[] | [.name,.verdicts[].metadata]`,
			},
			wantW: heredoc.Doc(`["chalk",{"commandline":"sh -c node npm-postinstall","npm_package_name":"@coreui/react","npm_package_version":"3.4.6","parent_name":"node"},{"commandline":"sh -c node -e \"try{require('./postinstall')}catch(e){}\"","npm_package_name":"core-js","npm_package_version":"3.26.1","parent_name":"node"}]
			`),
		},
		{
			name: "tsv",
			args: arguments{
				json: strings.NewReader(heredoc.Doc(`[
					{
						"name": "chalk",
						"version": "5.1.3",
						"shasum": "d957f370038b75ac572471e83be4c5ca9f8e8c45",
						"verdicts": [
							{
								"message": "npm install spawned a process",
								"severity": "medium",
								"metadata": {
									"npm_package_name": "@coreui/react",
									"npm_package_version": "3.4.6",
									"commandline": "sh -c node npm-postinstall",
									"parent_name": "node"
								}
							},
							{
								"message": "npm install spawned a process",
								"severity": "medium",
								"metadata": {
									"npm_package_name": "core-js",
									"npm_package_version": "3.26.1",
									"commandline": "sh -c node -e \"try{require('./postinstall')}catch(e){}\"",
									"parent_name": "node"
								}
							}
						]
					}
				]`)),
				expr: `.[] | select(.version == "5.1.3") | [.name,.version,(.verdicts | map(.metadata.npm_package_name) | join(","))] | @tsv`,
			},
			wantW: "chalk\t5.1.3\t@coreui/react,core-js\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			c := context.Background()
			err := Eval(c, tt.args.json, w, tt.args.expr)
			if tt.wantE {
				assert.Error(t, err)
				if tt.typeE != nil {
					assert.Equal(t, tt.typeE, err)
				}

				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantW, w.String())
		})
	}
}
