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
	"reflect"
	"runtime"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/XANi/goneric"
)

// Deps gets you the package lock dependencies.
func (p *packageLockJSON) Deps() map[string]PackageLockDependency {
	switch p.LockfileVersion.Value {
	case 1:
		return p.packageLockJSONVersion1.Dependencies
	case 2:
		return p.packageLockJSONVersion2.Dependencies
	case 3:
		return p.packageLockJSONVersion3.Dependencies
	default:
		return nil
	}
}

func (p *packageJSON) getDepsByType(t DependencyType) map[string]string {
	// TODO: assert t != All

	r := reflect.ValueOf(p)
	f := reflect.Indirect(r).FieldByName(t.Name())

	ret := map[string]string{}
	if f.Kind() == reflect.Map {
		for _, e := range f.MapKeys() {
			v := f.MapIndex(e)
			if v.Kind().String() == "string" {
				packageName := e.String()
				versionConstraint := v.Interface().(string)
				ret[packageName] = versionConstraint
			}
		}
	}

	return ret
}

type dep struct {
	name        string
	version     *semver.Version
	constraints *semver.Constraints
}

func getDepInstance(packageName, versionConstraint string) *dep {
	constraints, err := semver.NewConstraint(versionConstraint)
	// TODO: support URLs as dependencies (https://docs.npmjs.com/cli/v9/configuring-npm/package-json#dependencies)
	// TODO: those do not match as semver version constraints...
	if err != nil {
		return nil
	}

	return &dep{
		name:        packageName,
		constraints: constraints,
	}
}

func getDepsMapFromDepList(list []*dep, t DependencyType, out map[DependencyType]map[string]*semver.Version) {
	for _, resol := range list {
		// Ignore dependency if we were unable to resolve its version
		if resol == nil {
			continue
		}
		if resol.version != nil {
			if _, ok := out[t]; !ok {
				out[t] = map[string]*semver.Version{}
			}
			out[t][resol.name] = resol.version
		}
	}
}

func (p *packageJSON) Deps(ctx context.Context, resolve VersionResolutionStrategy, types ...DependencyType) map[DependencyType]map[string]*semver.Version {
	ret := map[DependencyType]map[string]*semver.Version{}

	// No input dependency types means give me all the dependencies for all the dependency types
	if len(types) == 0 {
		types = []DependencyType{All}
	}

	// Sort (ascending) dependency types
	sort.Slice(types, func(i, j int) bool {
		return types[i] < types[j]
	})

	// We assume the All type is the one with the lowest value
	if types[0] == All {
		types = AllDependencyTypes
	}

	for _, t := range types {
		depsByType := p.getDepsByType(t)

		if len(depsByType) == 0 {
			continue
		}

		// TODO: process the overrides field

		// Create a slice of dep instances
		deps := goneric.MapToSlice(getDepInstance, depsByType)

		// Resolve version constraints with parallel requests to the registry
		resolutions := goneric.ParallelMapSlice(func(input *dep) *dep {
			if input == nil {
				return nil
			}
			// Get all the versions matching the constraint
			collect, err := GetVersionsFromRegistry(ctx, input.name, input.constraints)
			// TODO: understand what to do when the HTTP call to the registry fails
			if err != nil {
				return nil
			}

			return &dep{
				name:    input.name,
				version: resolve(collect),
			}
		}, runtime.NumCPU(), deps)

		getDepsMapFromDepList(resolutions, t, ret)
	}

	return ret
}
