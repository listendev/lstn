package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type packageLockJSON struct {
	bytes []byte

	Name            string                           `json:"name"`
	Version         string                           `json:"version"`
	LockfileVersion int                              `json:"lockfileVersion"`
	Dependencies    map[string]PackageLockDependency `json:"dependencies"`
}

type PackageLockJSON interface {
	Shasums(ctx context.Context, timeout time.Duration) (Packages, error)
	Deps() map[string]PackageLockDependency
	Encode() string
}

type PackageLockDependency struct {
	Version  string `json:"version"`
	Resolved string `json:"resolved"`
}

type Package struct {
	Version string
	Shasum  string
}

type Packages map[string]Package

// NewPackageLockJSON is a factory to create an empty PackageLockJSON.
func NewPackageLockJSON() PackageLockJSON {
	ret := &packageLockJSON{}
	return ret
}

// NewPackageLockJSONFrom creates a PackageLockJSON instance from the package.json in the dir directory (if any).
func NewPackageLockJSONFrom(dir string) (PackageLockJSON, error) {
	var err error
	ret := &packageLockJSON{}

	// Get the package-lock.json file contents
	ret.bytes, err = generatePackageLock(dir)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(ret.bytes, ret)
	if err != nil {
		return nil, fmt.Errorf("couldn't instantiate from the input package-lock.json contents")
	}
	return ret, nil
}
