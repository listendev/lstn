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
package version

import (
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/listendev/lstn/pkg/validate"
)

// Unknown is the version string lstn ends up having when the build does not have VCS info.
const Unknown = "unknown"

// ImportPath is the package import path for the lstn core.
//
// Notice this identifier may be removed in the future.
const ImportPath = "github.com/listendev/lstn"

// VersionPrefix is a string that overrides lstn's version.
//
// It can be helpful when downstream packagers need to manually set the version.
// If no other version information is available, the short form version
// (see Get()) will be set to VersionPrefix, and the long version
// will include VersionPrefix at the beginning.
//
// For example, set this variable during `go build` with `-ldflags`:
//
// -ldflags '-X github.com/listendev/lstn/pkg/version.VersionPrefix=v1.0.0'.
var VersionPrefix string

// Version returns the lstn version.
//
// This function is experimental.
//
// Notice that lstn MUST be built with its `make/make` tool
// to properly embed complete version information.
//
// This function follows the following logic:
//  1. try to get version info from the build info provided by go.mod dependencies;
//  2. try to get version info from the embedded VCS info
//     (requires building from a git repository)
//  3. when no version is available it returns unknown for both versions (short, long)
//     if VersionPrefix is empty, otherwise it preprends both versions with it
//
// See relevant Go issues:
// - https://github.com/golang/go/issues/29228
// - https://github.com/golang/go/issues/50603
func Get() Version {
	var out Version

	var module *debug.Module
	bi, ok := debug.ReadBuildInfo()
	// Use VersionPrefix (if any) when build info is not available
	if !ok {
		if VersionPrefix != "" {
			out.Long = VersionPrefix
			out.Short = VersionPrefix

			return out
		}
		out.Long = Unknown
		out.Short = Unknown

		return out
	}

	// Detect if used as a module...
	for _, dep := range bi.Deps {
		if dep.Path == ImportPath {
			module = dep

			break
		}
	}

	// When used as a module...
	if module != nil {
		out.Short, out.Long = module.Version, module.Version
		if module.Sum != "" {
			out.Long += " " + module.Sum
		}
		if module.Replace != nil {
			out.Long += " => " + module.Replace.Path
			if module.Replace.Version != "" {
				out.Short = module.Replace.Version + "_custom"
				out.Long += "@" + module.Replace.Version
			}
			if module.Replace.Sum != "" {
				out.Long += " " + module.Replace.Sum
			}
		}
	}

	if out.Long == "" {
		var vcsRevision string
		var vcsTime time.Time
		var vcsModified bool
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs.revision":
				vcsRevision = setting.Value
			case "vcs.time":
				vcsTime, _ = time.Parse(time.RFC3339, setting.Value)
			case "vcs.modified":
				vcsModified, _ = strconv.ParseBool(setting.Value)
			}
		}

		if vcsRevision != "" {
			var modified string
			if vcsModified {
				modified = "dirty"
			}
			out.Long = fmt.Sprintf("%s.%s.%s", vcsRevision, vcsTime.Format("20060102T150405Z07:00"), modified)
			out.Short = vcsRevision

			// use out.Short checksum for out.Short, if hex-only
			if _, err := hex.DecodeString(out.Short); err == nil {
				out.Short = out.Short[:8]
			}

			// append date to out.Short
			if !vcsTime.IsZero() {
				out.Short += "." + vcsTime.Format("20060102")
			}
		}
	}

	if out.Long == "" {
		if VersionPrefix != "" {
			out.Long = VersionPrefix
		} else {
			out.Long = Unknown
		}
	} else if VersionPrefix != "" {
		out.Long = VersionPrefix + "+" + out.Long
	}

	if out.Short == "" || out.Short == "(devel)" {
		if VersionPrefix != "" {
			out.Short = VersionPrefix
		} else {
			out.Short = Unknown
		}
	} else if VersionPrefix != "" {
		out.Short = VersionPrefix + "+" + out.Short
	}

	return out
}

func Changelog(version string) (string, error) {
	basePath := "https://github.com/listendev/lstn"
	v := strings.TrimPrefix(version, "v")

	if errs := validate.Singleton.Var(v, "semver"); errs != nil {
		return "", fmt.Errorf("couldn't find a semver tag")
	}

	return fmt.Sprintf("%s/releases/tag/v%s", basePath, v), nil
}
