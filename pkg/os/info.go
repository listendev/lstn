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
package os

import (
	"fmt"

	goinfo "github.com/matishsiao/goInfo"
)

func NewInfo() (*Info, error) {
	info, err := goinfo.GetInfo()
	if err != nil {
		return nil, err
	}

	ret := &Info{}
	ret.OS = info.GoOS
	if ret.OS == "" && info.OS != "" {
		ret.OS = info.OS
	}

	ret.Arch = info.GoARCH
	if ret.Arch == "" && info.Platform != "" {
		ret.Arch = info.Platform
	}

	ret.Kernel = info.Kernel

	if info.Core != "unknown" && info.Core != "" {
		ret.KernelVersion = info.Core
	}

	ret.Hostname = info.Hostname

	return ret, nil
}

func (i *Info) FormatAsUserAgent() string {
	ret := ""
	// GOOS/GOARCH (hostname)
	if i.OS != "" {
		ret += i.OS
		if i.Arch != "" {
			ret += fmt.Sprintf("/%s", i.Arch)
		}
		if i.Hostname != "" {
			ret += fmt.Sprintf(" (%s)", i.Hostname)
		}
	}
	// Kernel/Version
	if i.Kernel != "" {
		if ret != "" {
			ret += " "
		}
		ret += i.Kernel
		if i.KernelVersion != "" {
			ret += fmt.Sprintf("/%s", i.KernelVersion)
		}
	}

	return ret
}
