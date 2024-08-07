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
package arguments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/listendev/pkg/lockfile"
	"github.com/stretchr/testify/require"
)

func TestGetLockfiles(t *testing.T) {
	absolute1, absolute1Err := filepath.Abs("testdata/package-lock.json")
	require.NoError(t, absolute1Err)

	type testCase struct {
		inputCWD       string
		inputLockfiles []string
		wantLockfiles  map[string]lockfile.Lockfile
		wantErrors     map[lockfile.Lockfile][]error
	}

	cases := []testCase{
		{
			inputCWD:       "testdata",
			inputLockfiles: []string{"package-lock.json"},
			wantLockfiles: map[string]lockfile.Lockfile{
				"testdata/package-lock.json": lockfile.PackageLockJSON,
			},
			wantErrors: map[lockfile.Lockfile][]error{},
		},
		{
			inputCWD:       "testdata",
			inputLockfiles: []string{absolute1},
			wantLockfiles: map[string]lockfile.Lockfile{
				"_CWD_/testdata/package-lock.json": lockfile.PackageLockJSON,
			},
			wantErrors: map[lockfile.Lockfile][]error{},
		},
		{
			inputCWD:       "testdata",
			inputLockfiles: []string{"package-lock.json", "package-lock.json"},
			wantLockfiles: map[string]lockfile.Lockfile{
				"testdata/package-lock.json": lockfile.PackageLockJSON,
			},
			wantErrors: map[lockfile.Lockfile][]error{},
		},
		{
			inputCWD:       "testdata",
			inputLockfiles: []string{"package-lock.json", "poetry.lock"},
			wantLockfiles: map[string]lockfile.Lockfile{
				"testdata/package-lock.json": lockfile.PackageLockJSON,
				"testdata/poetry.lock":       lockfile.PoetryLock,
			},
			wantErrors: map[lockfile.Lockfile][]error{},
		},
		{
			inputCWD:       "testdata",
			inputLockfiles: []string{"package-lock.json", "poetry.lock", "1/poetry.lock"},
			wantLockfiles: map[string]lockfile.Lockfile{
				"testdata/package-lock.json": lockfile.PackageLockJSON,
				"testdata/poetry.lock":       lockfile.PoetryLock,
				"testdata/1/poetry.lock":     lockfile.PoetryLock,
			},
			wantErrors: map[lockfile.Lockfile][]error{},
		},
		{
			inputCWD:       "testdata",
			inputLockfiles: []string{"package-lock.json", "poetry.lock", "1/poetry.lock", "unk/package-lock.json", "not-existing/poetry.lock"},
			wantLockfiles: map[string]lockfile.Lockfile{
				"testdata/package-lock.json": lockfile.PackageLockJSON,
				"testdata/poetry.lock":       lockfile.PoetryLock,
				"testdata/1/poetry.lock":     lockfile.PoetryLock,
			},
			wantErrors: map[lockfile.Lockfile][]error{
				lockfile.PackageLockJSON: {fmt.Errorf("testdata/unk/package-lock.json not found")},
				lockfile.PoetryLock:      {fmt.Errorf("testdata/not-existing/poetry.lock not found")},
			},
		},
		{
			inputCWD:       "testdata",
			inputLockfiles: []string{"unsupported-lockfile.json", "poetry.lock", "1/poetry.lock", "unk/package-lock.json", "not-existing/poetry.lock"},
			wantLockfiles: map[string]lockfile.Lockfile{
				"testdata/poetry.lock":   lockfile.PoetryLock,
				"testdata/1/poetry.lock": lockfile.PoetryLock,
			},
			wantErrors: map[lockfile.Lockfile][]error{
				lockfile.PackageLockJSON: {fmt.Errorf("testdata/unk/package-lock.json not found")},
				lockfile.PoetryLock:      {fmt.Errorf("testdata/not-existing/poetry.lock not found")},
			},
		},
	}

	cwd, _ := os.Getwd()

	for _, tc := range cases {
		gotLockfiles, gotErrors := GetLockfiles(tc.inputCWD, tc.inputLockfiles)

		got := map[string]lockfile.Lockfile{}
		for k, v := range gotLockfiles {
			got[strings.ReplaceAll(k, "_CWD_", cwd)] = v
		}

		want := map[string]lockfile.Lockfile{}
		for k, v := range tc.wantLockfiles {
			want[strings.ReplaceAll(k, "_CWD_", cwd)] = v
		}

		require.Equal(t, want, got)
		require.Equal(t, tc.wantErrors, gotErrors)
	}
}
