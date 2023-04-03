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
	"io"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/listendev/lstn/pkg/validate"
	"github.com/thediveo/enumflag/v2"
)

type DependencyType enumflag.Flag

const (
	All DependencyType = (iota + 1) * 22
	Dependencies
	DevDependencies
	PeerDependencies
	BundleDependencies
	OptionalDependencies
)

var AllDependencyTypes = []DependencyType{
	Dependencies,
	DevDependencies,
	PeerDependencies,
	BundleDependencies,
	OptionalDependencies,
}

var DependencyTypeIDs = map[DependencyType][]string{
	Dependencies:         {Dependencies.String()},
	DevDependencies:      {DevDependencies.String()},
	PeerDependencies:     {PeerDependencies.String()},
	BundleDependencies:   {BundleDependencies.String()},
	OptionalDependencies: {OptionalDependencies.String()},
}

func (dt DependencyType) String() string {
	switch dt {
	case Dependencies:
		return "dep"
	case DevDependencies:
		return "dev"
	case PeerDependencies:
		return "peer"
	case BundleDependencies:
		return "bundle"
	case OptionalDependencies:
		return "optional"
	default:
		return "all"
	}
}

func (dt DependencyType) Name() string {
	switch dt {
	case Dependencies:
		return "Dependencies"
	case DevDependencies:
		return "DevDependencies"
	case PeerDependencies:
		return "PeerDependencies"
	case BundleDependencies:
		return "BundleDependencies"
	case OptionalDependencies:
		return "OptionalDependencies"
	default:
		return ""
	}
}

type packageJSON struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	BundleDependencies   []string          `json:"bundleDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

type PackageJSON interface {
	FilterOutByTypes(...DependencyType)
	FilterOutByNames(...string)
	Deps(context.Context, VersionResolutionStrategy) map[DependencyType]map[string]*semver.Version
}

// The VersionResolutionStrategy is a function that, given a version constraints,
// returns back an exact version.
type VersionResolutionStrategy func(semver.Collection) *semver.Version

type LockfileVersion struct {
	Value int `json:"lockfileVersion" name:"lockfile version" validate:"gte=1,lte=3"`
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
	Deps() map[string]PackageLockDependency
	Encode() string
	Ok() bool
	Version() int
}

type PackageLockDependency struct {
	Version  string `json:"version"`
	Resolved string `json:"resolved"`
}

type Package struct {
	Version string `json:"version" name:"version" validate:"semver"`
	Shasum  string `json:"shasum" name:"shasum" validate:"shasum"`
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

// NewPackageLockJSON is a factory to create an empty (and invalid) PackageLockJSON.
func NewPackageLockJSON() PackageLockJSON {
	ret := &packageLockJSON{}

	return ret
}

func (p *packageLockJSON) Version() int {
	return p.LockfileVersion.Value
}

func (p *packageLockJSON) Ok() bool {
	err := validate.Singleton.Struct(p)

	return err == nil
}

// NewPackageLockJSONFromDir creates a PackageLockJSON instance from the package.json in the dir directory (if any).
func NewPackageLockJSONFromDir(ctx context.Context, dir string) (PackageLockJSON, error) {
	JSON, err := generatePackageLock(ctx, dir)
	if err != nil {
		return nil, err
	}

	return NewPackageLockJSONFromBytes(JSON)
}

// NewPackageLockJSONFromReader creates a PackageLockJSON instance from by reading the contents of a package-lock.json file.
func NewPackageLockJSONFromReader(reader io.Reader) (PackageLockJSON, error) {
	ret := &packageLockJSON{}
	if err := json.NewDecoder(reader).Decode(ret); err != nil {
		return nil, fmt.Errorf("couldn't instantiate from the input package-lock.json contents")
	}
	// TODO > set ret.bytes (TeeReader)

	return ret, nil
}

func NewPackageLockJSONFromBytes(b []byte) (PackageLockJSON, error) {
	ret := &packageLockJSON{}
	if err := json.Unmarshal(b, ret); err != nil {
		return nil, fmt.Errorf("couldn't instantiate from the input package-lock.json contents")
	}
	ret.bytes = b

	return ret, nil
}

func GetPackageJSONFromDir(dir string) (PackageJSON, error) {
	reader, err := readPackageJSON(dir)
	if err != nil {
		return nil, err
	}

	return NewPackageJSONFromReader(reader)
}

func NewPackageJSONFromReader(reader io.Reader) (PackageJSON, error) {
	ret := &packageJSON{}
	if err := json.NewDecoder(reader).Decode(ret); err != nil {
		return nil, fmt.Errorf("couldn't instantiate from the input package.json contents")
	}

	return ret, nil
}
