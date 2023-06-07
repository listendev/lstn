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

package packagesprinter

import (
	"bytes"
	"testing"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/stretchr/testify/require"
)

func TestTablePrinter_printVerdictMetadata(t *testing.T) {
	tests := []struct {
		name           string
		metadata       map[string]interface{}
		expectedOutput string
	}{
		{
			name:           "empty metadata does not print anything",
			metadata:       map[string]interface{}{},
			expectedOutput: "",
		},
		{
			name: "metadata with value nil does not print anything",
			metadata: map[string]interface{}{
				"key": nil,
			},
			expectedOutput: "",
		},
		{
			name: "metadata with value prints the key and value",
			metadata: map[string]interface{}{
				"key": "myvalue",
			},
			expectedOutput: "    key: myvalue\n",
		},
		{
			name: "metadata with value prints the key and value while ignoring npm_package_name and npm_package_version",
			metadata: map[string]interface{}{
				"key":                 "myvalue",
				"npm_package_name":    "react",
				"npm_package_version": "0.18.0",
			},
			expectedOutput: "    key: myvalue\n",
		},
		{
			name: "metadata with values prints the values it recognizes and ignores the rest",
			metadata: map[string]interface{}{
				"mystringkey":   "a string",
				"myintkey":      10,
				"somethingelse": float64(10.20),
			},
			expectedOutput: "    myintkey: 10\n    mystringkey: a string\n",
		},
		{
			name: "metadata with value prints the key and value while ignoring empty values",
			metadata: map[string]interface{}{
				"parent_name:":    "node",
				"executable_path": "/bin/sh",
				"commandline":     `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
				"server_ip":       "",
			},
			expectedOutput: "    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name:: node\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{

				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printVerdictMetadata(tt.metadata)

			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}

func TestTablePrinter_printVerdict(t *testing.T) {
	tests := []struct {
		name           string
		p              *listen.Package
		verdict        listen.Verdict
		expectedOutput string
	}{
		{
			name: "verdict with metadata prints the verdict and metadata",
			p: &listen.Package{
				Name:     "my-package",
				Version:  "1.0.0",
				Verdicts: []listen.Verdict{},

				Problems: []listen.Problem{},
			},
			verdict: listen.Verdict{
				Message:  "outbound network connection",
				Severity: "high",
				Metadata: map[string]interface{}{
					"parent_name":     "node",
					"executable_path": "/bin/sh",
					"commandline":     `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
					"server_ip":       "",
				},
			},
			expectedOutput: "  [high] outbound network connection\n    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name: node\n",
		},
		{
			name: "verdict with transitive metadata marks the verdict as transitive",
			p: &listen.Package{
				Name:     "my-package",
				Version:  "1.0.0",
				Verdicts: []listen.Verdict{},

				Problems: []listen.Problem{},
			},
			verdict: listen.Verdict{
				Message:  "outbound network connection",
				Severity: "high",
				Metadata: map[string]interface{}{
					"npm_package_name":    "react",
					"npm_package_version": "0.18.0",
					"parent_name":         "node",
					"executable_path":     "/bin/sh",
					"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
					"server_ip":           "",
				},
			},
			expectedOutput: "  [high] outbound network connection (from transitive dependency react@0.18.0)\n    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name: node\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{
				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printVerdict(tt.p, tt.verdict)

			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}

func TestTablePrinter_printProblem(t *testing.T) {
	tests := []struct {
		name           string
		problem        listen.Problem
		expectedOutput string
	}{
		{
			name: "problem with all details gets printed",
			problem: listen.Problem{
				Type:   "https://listen.dev/probs/invalid-name",
				Title:  "Package name not valid",
				Detail: "Package name not valid",
			},
			expectedOutput: "  - Package name not valid: https://listen.dev/probs/invalid-name\n",
		},
		{
			name: "problem with missing type",
			problem: listen.Problem{
				Type:   "",
				Title:  "Package name not valid",
				Detail: "Package name not valid",
			},
			expectedOutput: "  - Package name not valid: \n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{
				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printProblem(tt.problem)

			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}

func TestTablePrinter_printPackage(t *testing.T) {
	tests := []struct {
		name           string
		p              *listen.Package
		expectedOutput string
	}{
		{
			name: "package with no problems or verdicts gives this information along with the package name and version",
			p: &listen.Package{
				Name:     "my-package",
				Version:  "1.0.0",
				Verdicts: []listen.Verdict{},
				Problems: []listen.Problem{},
			},
			expectedOutput: "There are 0 verdicts and 0 problems for my-package@1.0.0\n\n\n",
		},
		{
			name: "package with a single verdict prints the verdict",
			p: &listen.Package{
				Name:    "my-package",
				Version: "1.0.0",
				Verdicts: []listen.Verdict{
					{
						Message:  "unexpected outbound connection destination",
						Severity: "high",
						Metadata: map[string]interface{}{
							"commandline":      "/usr/local/bin/node",
							"file_descriptor:": "10.0.2.100:47326->142.251.111.128:0",
							"server_ip":        "142.251.111.128",
							"executable_path":  "/usr/local/bin/node",
						},
					},
				},
				Problems: []listen.Problem{},
			},
			expectedOutput: "There is 1 verdict and 0 problems for my-package@1.0.0\n\n  [high] unexpected outbound connection destination\n    commandline: /usr/local/bin/node\n    executable_path: /usr/local/bin/node\n    file_descriptor:: 10.0.2.100:47326->142.251.111.128:0\n    server_ip: 142.251.111.128\n\n",
		},
		{
			name: "package with verdicts prints the verdicts recognizing transient dependencies",
			p: &listen.Package{
				Name:    "my-package",
				Version: "1.0.0",
				Verdicts: []listen.Verdict{
					{
						Message:  "npm spawned a child process",
						Severity: "high",
						Metadata: map[string]interface{}{
							"npm_package_name":    "react",
							"npm_package_version": "0.18.0",
							"parent_name":         "node",
							"executable_path":     "/bin/sh",
							"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
							"server_ip":           "",
						},
					},
					{
						Message:  "unexpected outbound connection destination",
						Severity: "high",
						Metadata: map[string]interface{}{
							"commandline":      "/usr/local/bin/node",
							"file_descriptor:": "10.0.2.100:47326->142.251.111.128:0",
							"server_ip":        "142.251.111.128",
							"executable_path":  "/usr/local/bin/node",
						},
					},
				},
				Problems: []listen.Problem{},
			},
			expectedOutput: "There are 2 verdicts and 0 problems for my-package@1.0.0\n\n  [high] npm spawned a child process (from transitive dependency react@0.18.0)\n    commandline: sh -c  node -e \"try{require('./_postinstall')}catch(e){}\" || exit 0\n    executable_path: /bin/sh\n    parent_name: node\n  [high] unexpected outbound connection destination\n    commandline: /usr/local/bin/node\n    executable_path: /usr/local/bin/node\n    file_descriptor:: 10.0.2.100:47326->142.251.111.128:0\n    server_ip: 142.251.111.128\n\n",
		},
		{
			name: "package with a single problem prints the problem",
			p: &listen.Package{
				Name:     "my-package",
				Version:  "1.0.0",
				Verdicts: []listen.Verdict{},
				Problems: []listen.Problem{
					{
						Type:   "https://listen.dev/probs/invalid-name",
						Title:  "Package name not valid",
						Detail: "Package name not valid",
					},
				},
			},
			expectedOutput: "There are 0 verdicts and 1 problem for my-package@1.0.0\n\n  - Package name not valid: https://listen.dev/probs/invalid-name\n\n",
		},
		{
			name: "package with many problems prints the problems",
			p: &listen.Package{
				Name:     "my-package",
				Version:  "1.0.0",
				Verdicts: []listen.Verdict{},
				Problems: []listen.Problem{
					{
						Type:   "https://listen.dev/probs/invalid-name",
						Title:  "Package name not valid",
						Detail: "Package name not valid",
					},
					{
						Type:   "https://listen.dev/probs/does-not-exist",
						Title:  "A problem that does not exist, just for testing",
						Detail: "A problem that does not exist, just for testing",
					},
				},
			},
			expectedOutput: "There are 0 verdicts and 2 problems for my-package@1.0.0\n\n  - Package name not valid: https://listen.dev/probs/invalid-name\n  - A problem that does not exist, just for testing: https://listen.dev/probs/does-not-exist\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{
				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printPackage(tt.p)
			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}

func TestTablePrinter_printPackages(t *testing.T) {
	tests := []struct {
		name           string
		packages       *listen.Response
		expectedOutput string
	}{
		{
			name: "packages with problems and verdicts prints the verdicts and problems",
			packages: &listen.Response{
				{
					Name:     "my-package",
					Version:  "1.0.0",
					Verdicts: []listen.Verdict{},
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
						{
							Type:   "https://listen.dev/probs/does-not-exist",
							Title:  "A problem that does not exist, just for testing",
							Detail: "A problem that does not exist, just for testing",
						},
					},
				},
				{
					Name:     "my-package",
					Version:  "1.0.0",
					Verdicts: []listen.Verdict{},
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
					},
				},
			},
			expectedOutput: "\nThere are 0 verdicts and 2 problems for my-package@1.0.0\n\n  - Package name not valid: https://listen.dev/probs/invalid-name\n  - A problem that does not exist, just for testing: https://listen.dev/probs/does-not-exist\n\n\nThere are 0 verdicts and 1 problem for my-package@1.0.0\n\n  - Package name not valid: https://listen.dev/probs/invalid-name\n\n",
		},
		{
			name: "empty packages prints nothing",
			packages: &listen.Response{
				{
					Name:     "my-package",
					Version:  "1.0.0",
					Verdicts: []listen.Verdict{},
					Problems: []listen.Problem{},
				},
			},
			expectedOutput: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{
				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			tr.printPackages(tt.packages)
			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}

func TestTablePrinter_printTable(t *testing.T) {
	tests := []struct {
		name           string
		packages       *listen.Response
		expectedOutput string
		wantErr        bool
	}{
		{
			name:           "empty packages prints nothing",
			packages:       &listen.Response{},
			expectedOutput: "",
		},
		{
			name: "inform that there are problems and verdicts in table format",
			packages: &listen.Response{
				{
					Name:    "react",
					Version: "1.0.0",
					Verdicts: []listen.Verdict{
						{
							Message:  "unexpected outbound connection destination",
							Severity: "high",
							Metadata: map[string]interface{}{
								"commandline":      "/usr/local/bin/node",
								"file_descriptor:": "10.0.2.100:47326->142.251.111.128:0",
								"server_ip":        "142.251.111.128",
								"executable_path":  "/usr/local/bin/node",
							},
						},
					},
					Problems: []listen.Problem{},
				},
				{
					Name:    "my-package",
					Version: "1.0.0",
					Verdicts: []listen.Verdict{
						{
							Message:  "npm spawned a child process",
							Severity: "high",
							Metadata: map[string]interface{}{
								"npm_package_name":    "react",
								"npm_package_version": "0.18.0",
								"parent_name":         "node",
								"executable_path":     "/bin/sh",
								"commandline":         `sh -c  node -e "try{require('./_postinstall')}catch(e){}" || exit 0`,
								"server_ip":           "",
							},
						},
						{
							Message:  "unexpected outbound connection destination",
							Severity: "high",
							Metadata: map[string]interface{}{
								"commandline":      "/usr/local/bin/node",
								"file_descriptor:": "10.0.2.100:47326->142.251.111.128:0",
								"server_ip":        "142.251.111.128",
								"executable_path":  "/usr/local/bin/node",
							},
						},
					},
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
						{
							Type:   "https://listen.dev/probs/does-not-exist",
							Title:  "A problem that does not exist, just for testing",
							Detail: "A problem that does not exist, just for testing",
						},
					},
				},
				{
					Name:     "my-package",
					Version:  "1.0.0",
					Verdicts: []listen.Verdict{},
					Problems: []listen.Problem{
						{
							Type:   "https://listen.dev/probs/invalid-name",
							Title:  "Package name not valid",
							Detail: "Package name not valid",
						},
					},
				},
			},
			expectedOutput: "react\t1.0.0\tX 1 verdicts\t✓ 0 problems\nmy-package\t1.0.0\tX 2 verdicts\t! 2 problems\nmy-package\t1.0.0\t✓ 0 verdicts\t! 1 problems\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			tr := &TablePrinter{
				streams: &iostreams.IOStreams{
					Out: outBuf,
				},
			}
			_ = tr.printTable(tt.packages)
			require.Equal(t, tt.expectedOutput, outBuf.String())
		})
	}
}
