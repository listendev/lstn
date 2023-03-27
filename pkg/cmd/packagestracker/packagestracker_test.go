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
package packagestracker

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Masterminds/semver/v3"
	generic "github.com/cli/cli/pkg/iostreams"
	"github.com/listendev/lstn/pkg/cmd/iostreams"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/stretchr/testify/require"
)

func testStreams(outBuf io.Writer) *iostreams.IOStreams {
	s := &iostreams.IOStreams{
		IOStreams: &generic.IOStreams{
			Out: outBuf,
		},
	}
	return s
}

func TestTrackPackages(t *testing.T) {
	tests := []struct {
		name                 string
		deps                 map[string]ListOfDependencies
		packageRetrievalFunc PackagesRetrievalFunc
		want                 *listen.Response
	}{
		{
			name: "dependencies with a single dep type and no errors",
			deps: map[string]ListOfDependencies{
				"dev": {
					{
						Name:    "foo",
						Version: semver.MustParse("1.0.0"),
					},
				},
			},
			packageRetrievalFunc: func(depName string, depVersion *semver.Version) (*listen.Response, error) {
				return &listen.Response{
					{
						Name:     "foo",
						Version:  "1.0.0",
						Shasum:   "1234",
						Verdicts: []listen.Verdict{},
						Problems: []listen.Problem{},
					},
				}, nil
			},
			want: &listen.Response{
				{
					Name:     "foo",
					Version:  "1.0.0",
					Shasum:   "1234",
					Verdicts: []listen.Verdict{},
					Problems: []listen.Problem{},
				},
			},
		},
		{
			name: "dependencies with a single dep type and errors",
			deps: map[string]ListOfDependencies{
				"dev": {
					{
						Name:    "foo",
						Version: semver.MustParse("1.0.0"),
					},
					{
						Name:    "bar",
						Version: semver.MustParse("1.0.0"),
					},
				},
			},
			packageRetrievalFunc: func(depName string, depVersion *semver.Version) (*listen.Response, error) {
				if depName == "foo" {
					return nil, errors.New("error retrieving package foo@1.0.0")
				}
				return &listen.Response{
					{
						Name:     "bar",
						Version:  "1.0.0",
						Shasum:   "1234",
						Verdicts: []listen.Verdict{},
						Problems: []listen.Problem{},
					},
				}, nil
			},
			want: &listen.Response{
				{
					Name:     "bar",
					Version:  "1.0.0",
					Shasum:   "1234",
					Verdicts: []listen.Verdict{},
					Problems: []listen.Problem{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			outBuf := &bytes.Buffer{}
			ios := testStreams(outBuf)
			ctx = context.WithValue(ctx, pkgcontext.IOStreamsKey, ios)

			got, err := TrackPackages(ctx, tt.deps, tt.packageRetrievalFunc)
			require.ErrorIs(t, err, nil)
			require.Equal(t, tt.want, got)

		})
	}
}
