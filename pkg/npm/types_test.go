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
package npm

import (
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
)

func TestPackageStruct(t *testing.T) {
	tests := []struct {
		desc string
		pack Package
		okay bool
	}{
		{
			"both-empty",
			Package{Version: "", Shasum: ""},
			false,
		},
		{
			"version-empty",
			Package{Version: "", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			false,
		},
		{
			"shasum-empty",
			Package{Version: "18.0.0", Shasum: ""},
			false,
		},
		{
			"version-invalid",
			Package{Version: "invalid", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			false,
		},
		{
			"shasum-invalid",
			Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb9"},
			false,
		},
		{
			"valid",
			Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.pack.Ok(), tc.okay)
		})
	}
}

func TestPackagesType(t *testing.T) {
	tests := []struct {
		desc string
		pkgs Packages
		okay bool
	}{
		{
			"empty",
			Packages{},
			false,
		},
		{
			"nil",
			nil,
			false,
		},
		{
			"invalid-key",
			Packages{
				"": Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			false,
		},
		{
			"invalid-package-version",
			Packages{
				"react": Package{Version: "", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			false,
		},
		{
			"invalid-package-shasum",
			Packages{
				"": Package{Version: "18.0.0", Shasum: ""},
			},
			false,
		},
		{
			"valid",
			Packages{
				"react": Package{Version: "18.0.0", Shasum: "b468736d1f4a5891f38585ba8e8fb29f91c3cb96"},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.pkgs.Ok(), tc.okay)
		})
	}
}

func TestPackageLockJSONInstantiation(t *testing.T) {
	invalid := NewPackageLockJSON()

	assert.False(t, invalid.Ok())

	lockfileVersionFuture, err := NewPackageLockJSONFromBytes([]byte(heredoc.Doc(`{
		"name": "sample",
		"version": "1.0.0",
		"lockfileVersion": 22,
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

	if assert.Error(t, err) {
		assert.Equal(t, "couldn't instantiate from the input package-lock.json contents", err.Error())
		assert.Nil(t, lockfileVersionFuture)
	}

	valid, err := NewPackageLockJSONFromBytes([]byte(heredoc.Doc(`{
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

	if assert.Nil(t, err) {
		assert.True(t, valid.Ok())
	}
}

func TestNewPackageJSONFromReader(t *testing.T) {
	tests := []struct {
		desc    string
		input   string
		output  *packageJSON
		wantErr string
	}{
		{
			desc:    "empty",
			input:   "",
			output:  nil,
			wantErr: "couldn't instantiate from the input package.json contents",
		},
		{
			desc: "dep-devdep-peer-peermeta-bundle",
			input: heredoc.Doc(`{
	"name": "xxx",
	"dependencies": {
		"@isaacs/import-jsx": "^4.0.1",
		"@types/react": "^17.0.52",
		"chokidar": "^3.3.0",
		"findit": "^2.0.0",
		"foreground-child": "^2.0.0",
		"fs-exists-cached": "^1.0.0",
		"glob": "^7.2.3",
		"ink": "^3.2.0",
		"isexe": "^2.0.0",
		"istanbul-lib-processinfo": "^2.0.3",
		"jackspeak": "^1.4.2",
		"libtap": "^1.4.0",
		"minipass": "^3.3.4",
		"mkdirp": "^1.0.4",
		"nyc": "^15.1.0",
		"opener": "^1.5.1",
		"react": "^17.0.2",
		"rimraf": "^3.0.0",
		"signal-exit": "^3.0.6",
		"source-map-support": "^0.5.16",
		"tap-mocha-reporter": "^5.0.3",
		"tap-parser": "^11.0.2",
		"tap-yaml": "^1.0.2",
		"tcompare": "^5.0.7",
		"treport": "^3.0.4",
		"which": "^2.0.2"
	},
	"devDependencies": {
		"coveralls": "^3.1.1",
		"eslint": "^7.32.0",
		"flow-remove-types": "^2.193.0",
		"node-preload": "^0.2.1",
		"process-on-spawn": "^1.0.0",
		"ts-node": "^8.5.2",
		"typescript": "^3.7.2"
	},
	"peerDependencies": {
		"coveralls": "^3.1.1",
		"flow-remove-types": ">=2.112.0",
		"ts-node": ">=8.5.2",
		"typescript": ">=3.7.2"
	},
	"peerDependenciesMeta": {
		"coveralls": {
		"optional": true
		},
		"flow-remove-types": {
		"optional": true
		},
		"ts-node": {
		"optional": true
		},
		"typescript": {
		"optional": true
		}
	},
	"bundleDependencies": [
		"ink",
		"treport",
		"@types/react",
		"@isaacs/import-jsx",
		"react"
	]
}`),
			output: &packageJSON{
				Dependencies: map[string]string{
					"@isaacs/import-jsx":       "^4.0.1",
					"@types/react":             "^17.0.52",
					"chokidar":                 "^3.3.0",
					"findit":                   "^2.0.0",
					"foreground-child":         "^2.0.0",
					"fs-exists-cached":         "^1.0.0",
					"glob":                     "^7.2.3",
					"ink":                      "^3.2.0",
					"isexe":                    "^2.0.0",
					"istanbul-lib-processinfo": "^2.0.3",
					"jackspeak":                "^1.4.2",
					"libtap":                   "^1.4.0",
					"minipass":                 "^3.3.4",
					"mkdirp":                   "^1.0.4",
					"nyc":                      "^15.1.0",
					"opener":                   "^1.5.1",
					"react":                    "^17.0.2",
					"rimraf":                   "^3.0.0",
					"signal-exit":              "^3.0.6",
					"source-map-support":       "^0.5.16",
					"tap-mocha-reporter":       "^5.0.3",
					"tap-parser":               "^11.0.2",
					"tap-yaml":                 "^1.0.2",
					"tcompare":                 "^5.0.7",
					"treport":                  "^3.0.4",
					"which":                    "^2.0.2",
				},
				DevDependencies: map[string]string{
					"coveralls":         "^3.1.1",
					"eslint":            "^7.32.0",
					"flow-remove-types": "^2.193.0",
					"node-preload":      "^0.2.1",
					"process-on-spawn":  "^1.0.0",
					"ts-node":           "^8.5.2",
					"typescript":        "^3.7.2",
				},
				PeerDependencies: map[string]string{
					"coveralls":         "^3.1.1",
					"flow-remove-types": ">=2.112.0",
					"ts-node":           ">=8.5.2",
					"typescript":        ">=3.7.2",
				},
				BundleDependencies: []string{
					"ink",
					"treport",
					"@types/react",
					"@isaacs/import-jsx",
					"react",
				},
			},
			wantErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := NewPackageJSONFromReader(strings.NewReader(tc.input))
			if err != nil {
				assert.Nil(t, res)
				if assert.Error(t, err) {
					assert.Equal(t, tc.wantErr, err.Error())
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, res)
			}
		})
	}
}
