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
package pypi

import (
	"bytes"
	"fmt"
	"io"

	"github.com/listendev/lstn/pkg/fs"
	listentype "github.com/listendev/lstn/pkg/listen/type"
	"github.com/listendev/pkg/lockfile"
	"github.com/pelletier/go-toml/v2"
)

type PoetryLock interface {
	listentype.AnalysisRequester
}

func NewPoetryLockFromBytes(b []byte) (PoetryLock, error) {
	ret := &poetryLock{}
	if err := toml.Unmarshal(b, ret); err != nil {
		return nil, fmt.Errorf("couldn't decode from the input %s contents", lockfile.PoetryLock.String())
	}
	ret.bytes = b

	return ret, nil
}

// NewPoetryLockFromReader creates a PoetryLock instance from by reading the contents of a poetry.lock file.
func NewPoetryLockFromReader(reader io.Reader) (PoetryLock, error) {
	ret := &poetryLock{}
	var b bytes.Buffer
	r := io.TeeReader(reader, &b)
	// NOTE: we aren't actually verify the poetry.lock is well-formed, just decoding it from TOML
	if err := toml.NewDecoder(r).Decode(ret); err != nil {
		return nil, fmt.Errorf("couldn't decode from the input %s contents", lockfile.PoetryLock.String())
	}
	ret.bytes = b.Bytes()

	return ret, nil
}

// GetPoetryLockFromDir creates a PoetryLock instance from the existing poetry.lock in dir, if any.
func GetPoetryLockFromDir(dir string) (PoetryLock, error) {
	reader, err := fs.Read(dir, lockfile.PoetryLock.String())
	if err != nil {
		return nil, err
	}

	return NewPoetryLockFromReader(reader)
}
