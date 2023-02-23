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
package context

type contextKey string

// ConfigKey is the key indexing the configuration options in the contexts.
var ConfigKey contextKey = "cfg"

// ContextCancelFuncKey is the key indexing the context cancelation function in the context itself.
var ContextCancelFuncKey contextKey = "ctxcancel"

// InKey is the key indexing the options for the `in` child command.
var InKey contextKey = "in"

// ToKey is the key indexing the options for the `to` child command.
var ToKey contextKey = "to"

// VersionKey is the key indexing the options for the `version` child command.
var VersionKey contextKey = "version"

// VersionTagKey is the key storing the tag part of the version.
var VersionTagKey contextKey = "version_tag"

// VersionShortKey is the key storing the short version.
var VersionShortKey contextKey = "version_short"

// VersionLongKey is the key storing the long version.
var VersionLongKey contextKey = "version_long"

// FIXME: remove
// ShortVersionKey is the key storing the short version.
var ShortVersionKey contextKey = "version_short"
