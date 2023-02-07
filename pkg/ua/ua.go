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
	shortVersion, longVersion := version.Version()
	ret := fmt.Sprintf("lstn/%s (%s", shortVersion, longVersion)
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
