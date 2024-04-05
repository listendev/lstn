package arguments

import (
	"fmt"
	"testing"

	"github.com/listendev/pkg/lockfile"
	"github.com/stretchr/testify/require"
)

func TestGetLockfiles(t *testing.T) {
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

	for _, tc := range cases {
		gotLockfiles, gotErrors := GetLockfiles(tc.inputCWD, tc.inputLockfiles)
		require.Equal(t, tc.wantLockfiles, gotLockfiles)
		require.Equal(t, tc.wantErrors, gotErrors)
	}

}
