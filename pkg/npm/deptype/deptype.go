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
package deptype

import (
	"fmt"

	"github.com/thediveo/enumflag/v2"
)

type Enum enumflag.Flag

const (
	All Enum = (iota + 1) * 22
	Dependencies
	DevDependencies
	PeerDependencies
	BundleDependencies
	OptionalDependencies
)

var AllTypes = []Enum{
	Dependencies,
	DevDependencies,
	PeerDependencies,
	BundleDependencies,
	OptionalDependencies,
}

var IDs = map[Enum][]string{
	Dependencies:         {Dependencies.String()},
	DevDependencies:      {DevDependencies.String()},
	PeerDependencies:     {PeerDependencies.String()},
	BundleDependencies:   {BundleDependencies.String()},
	OptionalDependencies: {OptionalDependencies.String()},
}

func (e Enum) String() string {
	switch e {
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

func (e Enum) Name() string {
	switch e {
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

func Parse(s string) (Enum, error) {
	for t, vals := range IDs {
		for _, v := range vals {
			if s == v {
				return t, nil
			}
		}
	}

	return All, fmt.Errorf(`couldn't parse "%s" in an NPM dependency type`, s)
}

func ParseMultiple(in ...string) ([]Enum, error) {
	res := []Enum{}
	for _, i := range in {
		val, err := Parse(i)
		if err != nil {
			return nil, err
		}
		res = append(res, val)
	}

	return res, nil
}
