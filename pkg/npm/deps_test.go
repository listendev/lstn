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
	"testing"

	"github.com/Masterminds/semver/v3"
	npmdeptype "github.com/listendev/lstn/pkg/npm/deptype"
	"github.com/stretchr/testify/assert"
)

func TestFilterOutByNames(t *testing.T) {
	cases := []struct {
		des        string
		exclusions []string
		expect     *packageJSON
	}{
		{
			des:        "no-exclusions",
			exclusions: []string{},
			expect: &packageJSON{
				Dependencies: map[string]string{
					"tap-parser":       "^11.0.2",
					"@types/react":     "^17.0.52",
					"chokidar":         "^3.3.0",
					"findit":           "^2.0.0",
					"foreground-child": "^2.0.0",
					"fs-exists-cached": "^1.0.0",
					"glob":             "^7.2.3",
				},
				BundleDependencies: []string{
					"ink",
					"treport",
					"@types/react",
				},
			},
		},
		{
			des:        "ignore-packages-not-in-packagejson",
			exclusions: []string{"not", "in", "list"},
			expect: &packageJSON{
				Dependencies: map[string]string{
					"tap-parser":       "^11.0.2",
					"@types/react":     "^17.0.52",
					"chokidar":         "^3.3.0",
					"findit":           "^2.0.0",
					"foreground-child": "^2.0.0",
					"fs-exists-cached": "^1.0.0",
					"glob":             "^7.2.3",
				},
				BundleDependencies: []string{
					"ink",
					"treport",
					"@types/react",
				},
			},
		},
		{
			des:        "remove-2-bundles-2-deps",
			exclusions: []string{"@types/react", "findit", "foreground-child", "treport"},
			expect: &packageJSON{
				Dependencies: map[string]string{
					"tap-parser":       "^11.0.2",
					"chokidar":         "^3.3.0",
					"fs-exists-cached": "^1.0.0",
					"glob":             "^7.2.3",
				},
				BundleDependencies: []string{
					"ink",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		pack := &packageJSON{
			Dependencies: map[string]string{
				"tap-parser":       "^11.0.2",
				"@types/react":     "^17.0.52",
				"chokidar":         "^3.3.0",
				"findit":           "^2.0.0",
				"foreground-child": "^2.0.0",
				"fs-exists-cached": "^1.0.0",
				"glob":             "^7.2.3",
			},
			BundleDependencies: []string{
				"ink",
				"treport",
				"@types/react",
			},
		}

		t.Run(tc.des, func(t *testing.T) {
			pack.FilterOutByNames(tc.exclusions...)
			assert.Equal(t, tc.expect, pack)
		})
	}
}

func TestGetDepsByType(t *testing.T) {
	ed1 := map[string]string{
		"react": "18.0.0",
	}
	eb1 := map[string]string{
		"ink":     "",
		"treport": "",
	}

	p1 := &packageJSON{
		Dependencies:       ed1,
		BundleDependencies: []string{"ink", "treport"},
	}
	assert.Equal(t, ed1, p1.getDepsByType(npmdeptype.Dependencies))
	assert.Equal(t, eb1, p1.getDepsByType(npmdeptype.BundleDependencies))
	assert.Equal(t, map[string]string{}, p1.getDepsByType(npmdeptype.DevDependencies))

	e2 := map[string]string{
		"react":      "18.0.0",
		"node-fetch": "2.6.9",
	}

	p2 := &packageJSON{
		DevDependencies: e2,
	}
	assert.Equal(t, map[string]string{}, p2.getDepsByType(npmdeptype.Dependencies))
	assert.Equal(t, e2, p2.getDepsByType(npmdeptype.DevDependencies))
}

func TestGetDepsMapFromDepList(t *testing.T) {
	ret := map[npmdeptype.Enum]map[string]*semver.Version{}
	getDepsMapFromDepList([]*dep{}, npmdeptype.DevDependencies, ret)
	assert.Empty(t, ret)

	depList := []*dep{
		&dep{
			name:    "core-js",
			version: semver.MustParse("3.29.1"),
		},
		&dep{
			name:    "vue",
			version: semver.MustParse("3.2.47"),
		},
		nil,
		&dep{
			name:    "react",
			version: semver.MustParse("18.2.0"),
		},
	}
	getDepsMapFromDepList(depList, npmdeptype.Dependencies, ret)
	assert.Len(t, ret, 1)
	assert.Len(t, ret[npmdeptype.Dependencies], 3)
	assert.Equal(t, map[npmdeptype.Enum]map[string]*semver.Version{
		npmdeptype.Dependencies: map[string]*semver.Version{
			"react":   semver.MustParse("18.2.0"),
			"vue":     semver.MustParse("3.2.47"),
			"core-js": semver.MustParse("3.29.1"),
		},
	}, ret)
}

func TestGetDepInstance(t *testing.T) {
	v1820, _ := semver.NewConstraint("18.2.0")
	v267, _ := semver.NewConstraint("^2.6.7")

	cases := []struct {
		des string
		pkg string
		ver string
		res *dep
	}{
		{
			des: "exact version",
			pkg: "react",
			ver: "18.2.0",
			res: &dep{
				name:        "react",
				constraints: v1820,
			},
		},
		{
			des: "constraints",
			pkg: "node-fetch",
			ver: "^2.6.7",
			res: &dep{
				name:        "node-fetch",
				constraints: v267,
			},
		},
		// Unsupported
		{
			des: "local path",
			pkg: "dyl",
			ver: "file:.../dyl",
			res: nil,
		},
		{
			des: "URL",
			pkg: "asd",
			ver: "http://asdf.com/asfg.tar.gz",
			res: nil,
		},
		{
			des: "git URLs",
			pkg: "cli",
			ver: "git+ssh://git@github.com:npm/cli#semver:^5.0",
			res: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.des, func(t *testing.T) {
			assert.Equal(t, tc.res, getDepInstance(tc.pkg, tc.ver))
		})
	}
}
