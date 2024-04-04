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
package listen

import (
	"context"

	"github.com/creasty/defaults"
	"github.com/listendev/lstn/pkg/cmd/flags"
	"github.com/listendev/pkg/ecosystem"
)

type options struct {
	baseURL   string
	userAgent string
	ecosystem ecosystem.Ecosystem
	ctx       context.Context
	json      flags.JSONFlags
}

func newOptions(opts ...func(*options)) (*options, error) {
	obj := &options{}

	for _, o := range opts {
		o(obj)
	}

	if err := defaults.Set(obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (o *options) SetDefaults() {
	if o.ctx == nil {
		o.ctx = context.Background()
	}
}

func WithJSONOptions(input flags.JSONFlags) func(*options) {
	return func(o *options) {
		o.json = input
	}
}

func WithBaseURL(input string) func(*options) {
	return func(o *options) {
		o.baseURL = input
	}
}

func WithContext(input context.Context) func(*options) {
	return func(o *options) {
		o.ctx = input
	}
}

func WithUserAgent(userAgent string) func(*options) {
	return func(o *options) {
		o.userAgent = userAgent
	}
}

func WithEcosystem(eco ecosystem.Ecosystem) func(*options) {
	return func(o *options) {
		o.ecosystem = eco
	}
}
