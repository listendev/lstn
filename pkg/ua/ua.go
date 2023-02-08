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
package ua

import (
	"fmt"
	"runtime"

	"github.com/listendev/lstn/pkg/version"
	goinfo "github.com/matishsiao/goInfo"
)

// Generate creates a user-agent string for the current lstn version.
//
// If the os parameters is true, it also appends the available info like
// the os, the architecture, the kernel and its version, and the hostname.
func Generate(withOS bool, comments ...string) string {
	version := version.Get()
	ret := fmt.Sprintf("lstn/%s (%s", version.Short, version.Long)
	counter, _, _, success := runtime.Caller(1)
	if success {
		ret += fmt.Sprintf("; %s", runtime.FuncForPC(counter).Name())
	}

	for _, comment := range comments {
		ret += fmt.Sprintf("; %s", comment)
	}
	ret += ")"

	if withOS {
		if info, err := goinfo.GetInfo(); err == nil {
			// GOOS/GOARCH (hostname)
			osStr := ""
			if info.GoOS != "" {
				osStr += fmt.Sprintf(" %s/%s", info.GoOS, info.GoARCH)
				if info.Hostname != "" {
					osStr += fmt.Sprintf(" (%s)", info.Hostname)
				}
			}
			ret += osStr

			// Kernel/Version
			kernelStr := ""
			if info.Kernel != "" && info.Kernel != "unknown" {
				kernelStr += fmt.Sprintf(" %s", info.Kernel)
				if info.Core != "" && info.Core != "unknown" {
					kernelStr += fmt.Sprintf("/%s", info.Core)
				}
			}
			ret += kernelStr
		}
	}

	return ret
}
