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

// ImportPath is the package import path for the lstn core.
//
// Notice this identifier may be removed in the future.
const ImportPath = "github.com/listendev/lstn"

// VersionPrefix is a string that overrides lstn's version.
//
// It can be helpful when downstream packagers need to manually set the version.
// If no other version information is available, the short form version
// (see Version()) will be set to VersionPrefix, and the long version
// will include VersionPrefix at the beginning.
//
// For example, set this variable during `go build` with `-ldflags`:
//
// -ldflags '-X github.com/listendev/lstn/pkg/version.VersionPrefix=v1.0.0'
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
func Version() (short, long string) {
	var module *debug.Module
	bi, ok := debug.ReadBuildInfo()
	// Use VersionPrefix (if any) when build info is not available
	if !ok {
		if VersionPrefix != "" {
			long = VersionPrefix
			short = VersionPrefix
			return
		}
		long = "unknown"
		short = "unknown"
		return
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
		short, long = module.Version, module.Version
		if module.Sum != "" {
			long += " " + module.Sum
		}
		if module.Replace != nil {
			long += " => " + module.Replace.Path
			if module.Replace.Version != "" {
				short = module.Replace.Version + "_custom"
				long += "@" + module.Replace.Version
			}
			if module.Replace.Sum != "" {
				long += " " + module.Replace.Sum
			}
		}
	}

	if long == "" {
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
			long = fmt.Sprintf("%s.%s.%s", vcsRevision, vcsTime.Format("20060102T150405Z07:00"), modified)
			short = vcsRevision

			// use short checksum for short, if hex-only
			if _, err := hex.DecodeString(short); err == nil {
				short = short[:8]
			}

			// append date to short
			if !vcsTime.IsZero() {
				short += "." + vcsTime.Format("20060102")
			}
		}
	}

	if long == "" {
		if VersionPrefix != "" {
			long = VersionPrefix
		} else {
			long = "unknown"
		}
	} else if VersionPrefix != "" {
		long = VersionPrefix + "+" + long
	}

	if short == "" || short == "(devel)" {
		if VersionPrefix != "" {
			short = VersionPrefix
		} else {
			short = "unknown"
		}
	} else if VersionPrefix != "" {
		short = VersionPrefix + "+" + short
	}

	return
}

func Changelog(version string) (string, error) {
	basePath := "https://github.com/listendev/lstn"
	v := strings.TrimPrefix(version, "v")

	if errs := validate.Singleton.Var(v, "semver"); errs != nil {
		return "", fmt.Errorf("couldn't find a semver tag")
	}

	return fmt.Sprintf("%s/releases/tag/v%s", basePath, v), nil
}
 