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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/listendev/lstn/pkg/validate"
)

type LockfileVersion struct {
	Value int `json:"lockfileVersion"`
}

type packageLockJSONVersion1 struct {
	Name         string                           `json:"name"`
	Version      string                           `json:"version"`
	Dependencies map[string]PackageLockDependency `json:"dependencies"`
}

type packageLockJSONVersion2 struct {
	Name         string                           `json:"name"`
	Version      string                           `json:"version"`
	Dependencies map[string]PackageLockDependency `json:"dependencies"`
}

type packageLockJSONVersion3 struct {
	Name         string                           `json:"name"`
	Version      string                           `json:"version"`
	Dependencies map[string]PackageLockDependency `json:"packages"`
}

type packageLockJSON struct {
	LockfileVersion
	*packageLockJSONVersion1
	*packageLockJSONVersion2
	*packageLockJSONVersion3
	bytes []byte
}

func (p *packageLockJSON) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &p.LockfileVersion); err != nil {
		return err
	}

	switch p.LockfileVersion.Value {
	case 1:
		p.packageLockJSONVersion1 = &packageLockJSONVersion1{}

		return json.Unmarshal(data, p.packageLockJSONVersion1)
	case 2:
		p.packageLockJSONVersion2 = &packageLockJSONVersion2{}

		return json.Unmarshal(data, p.packageLockJSONVersion2)
	case 3:
		p.packageLockJSONVersion3 = &packageLockJSONVersion3{}
		if err := json.Unmarshal(data, p.packageLockJSONVersion3); err != nil {
			return err
		}
		for k := range p.packageLockJSONVersion3.Dependencies {
			if k == "" {
				delete(p.packageLockJSONVersion3.Dependencies, k)
			}
			if strings.HasPrefix(k, "node_modules") {
				newk := strings.TrimPrefix(k, "node_modules/")
				p.packageLockJSONVersion3.Dependencies[newk] = p.packageLockJSONVersion3.Dependencies[k]
				delete(p.packageLockJSONVersion3.Dependencies, k)
			}
		}

		return nil
	default:
		return fmt.Errorf("unsupported package-lock.json version")
	}
}

type PackageLockJSON interface {
	Shasums(ctx context.Context) (Packages, error)
	Deps() map[string]PackageLockDependency
	Encode() string
}

type PackageLockDependency struct {
	Version  string `json:"version"`
	Resolved string `json:"resolved"`
}

type Package struct {
	Version string `json:"version" validate:"semver"`
	Shasum  string `json:"shasum" validate:"len=40"`
}

func (p *Package) Ok() bool {
	return validate.Singleton.Struct(p) == nil
}

type Packages map[string]Package

func (p Packages) Ok() bool {
	if len(p) == 0 {
		return false
	}

	for name, pack := range p {
		if validate.Singleton.Var(name, "npm_package_name") != nil {
			return false
		}
		if !pack.Ok() {
			return false
		}
	}

	return true
}

// NewPackageLockJSON is a factory to create an empty PackageLockJSON.
func NewPackageLockJSON() PackageLockJSON {
	ret := &packageLockJSON{}

	return ret
}

// NewPackageLockJSONFrom creates a PackageLockJSON instance from the package.json in the dir directory (if any).
func NewPackageLockJSONFrom(ctx context.Context, dir string) (PackageLockJSON, error) {
	var err error
	ret := &packageLockJSON{}

	// Get the package-lock.json file contents
	ret.bytes, err = generatePackageLock(ctx, dir)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(ret.bytes, ret)
	if err != nil {
		return nil, fmt.Errorf("couldn't instantiate from the input package-lock.json contents")
	}

	return ret, nil
}
