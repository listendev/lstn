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
package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/imdario/mergo"
	"github.com/muja/goconfig"
)

// ActiveFS provides a configurable entry point to the billy.Filesystem in active use
// for reading global and system level configuration files.
//
// Override this in tests to mock the filesystem
// (then reset to restore default behavior).
var ActiveFS = DefaultFS()

// DefaultFS provides a billy.Filesystem abstraction over the
// OS filesystem (via osfs.OS) scoped to the root directory.
// in order to enable access to global and system configuration files
// via absolute paths.
func DefaultFS() billy.Filesystem {
	return osfs.New("/")
}

// NewConfigFromMap provides a config.Config instance populating it from
// the input map.
func NewConfigFromMap(src map[string]string) *config.Config {
	cfg := config.NewConfig()

	// We map only the git config sections that we're interested in
	for k, v := range src {
		sect := strings.Split(k, ".")
		switch sect[0] {
		case "user":
			switch sect[1] {
			case "name":
				cfg.User.Name = v
			case "email":
				cfg.User.Email = v
			}
		case "author":
			switch sect[1] {
			case "name":
				cfg.Author.Name = v
			case "email":
				cfg.Author.Email = v
			}
		case "remote":
			// Ignore malformed remote
			if len(sect) < 3 {
				continue
			}
			// Upsert key
			remoteName := strings.Join(sect[1:len(sect)-1], ".")
			if _, ok := cfg.Remotes[remoteName]; !ok {
				cfg.Remotes[remoteName] = &config.RemoteConfig{
					Name: remoteName,
				}
			}
			// Populate
			remoteKey := sect[len(sect)-1]
			switch remoteKey {
			case "url":
				if len(cfg.Remotes[remoteName].URLs) == 0 {
					cfg.Remotes[remoteName].URLs = []string(nil)
				}
				cfg.Remotes[remoteName].URLs = append(cfg.Remotes[remoteName].URLs, v)
			case "fetch":
				if len(cfg.Remotes[remoteName].Fetch) == 0 {
					cfg.Remotes[remoteName].Fetch = []config.RefSpec{}
				}
				rs := config.RefSpec(v)
				if err := rs.Validate(); err != nil {
					// Ignore malformed remote fetch key
					continue
				}
				cfg.Remotes[remoteName].Fetch = append(cfg.Remotes[remoteName].Fetch, rs)
			}
		case "url":
			// Ignore malformed url
			if len(sect) < 3 {
				continue
			}
			// Upsert key
			urlName := strings.Join(sect[1:len(sect)-1], ".")
			if _, ok := cfg.URLs[urlName]; !ok {
				cfg.URLs[urlName] = &config.URL{
					Name: urlName,
				}
			}
			// Populate
			urlKey := sect[len(sect)-1]
			if urlKey == "insteadof" {
				cfg.URLs[urlName].InsteadOf = v
			}
		}
	}

	return cfg
}

// ReadConfig parses a Git configuration file in an io.Reader.
func ReadConfig(r io.Reader) (*config.Config, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	raw, _, err := goconfig.Parse(b)
	if err != nil {
		return nil, err
	}

	return NewConfigFromMap(raw), nil
}

// GetConfig returns an instance of config.Config for the required scope
// (system or global Git configuration).
func GetConfig(s config.Scope) (*config.Config, error) {
	files, err := config.Paths(s)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		f, err := ActiveFS.Open(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}
		defer f.Close()

		return ReadConfig(f)
	}

	return config.NewConfig(), nil
}

// GetFinalConfig retrives system, global, and local Git configurations
// and then merges them into a final config.Config instance.
//
// This function also applies Git URL rules (if any) to the Git remotes.
func GetFinalConfig(repo *git.Repository) (*config.Config, error) {
	var err error
	c := config.NewConfig()
	// Try to retrive and to merge the system Git config
	systemCfg, _ := GetConfig(config.SystemScope)
	err = mergo.Merge(c, systemCfg, mergo.WithOverride)
	if err != nil {
		return nil, fmt.Errorf("couldn't merge system git config: %#v", err)
	}

	// Try to retrieve and to merge the global Git config
	globalCfg, _ := GetConfig(config.GlobalScope)
	err = mergo.Merge(c, globalCfg, mergo.WithOverride)
	if err != nil {
		return nil, fmt.Errorf("couldn't merge global git config: %#v", err)
	}

	// Try to retrive the local Git config
	localCfg, _ := repo.Config()
	localCfgOk := localCfg != nil
	// Fallback to manual reading
	if localCfg == nil {
		// Try to get the local Git filesystem
		fs, isFSBased := repo.Storer.(interface{ Filesystem() billy.Filesystem })
		if !isFSBased {
			return nil, fmt.Errorf("couldn't get the local git filesystem: %#v", err)
		}
		// Read the local Git config file
		f, err := ActiveFS.Open(filepath.Join(fs.Filesystem().Root(), "config"))
		if err == nil {
			var err error
			localCfg, err = ReadConfig(f)
			localCfgOk = err == nil
		}
		defer f.Close()
	}
	// Merge the local Git config
	if localCfgOk {
		err = mergo.Merge(c, localCfg, mergo.WithOverride)
		if err != nil {
			return nil, fmt.Errorf("couldn't merge local git config: %#v", err)
		}
	}

	// Apply URL rules (if any)
	applyURLRules(c)

	return c, nil
}

func applyURLRules(c *config.Config) {
	urlRules := c.URLs

	for _, remote := range c.Remotes {
		for i, url := range remote.URLs {
			if matchingURLRule := findLongestInsteadOfMatch(url, urlRules); matchingURLRule != nil {
				remote.URLs[i] = applyInsteadOf(matchingURLRule, remote.URLs[i])
			}
		}
	}
}

func findLongestInsteadOfMatch(remoteURL string, urls map[string]*config.URL) *config.URL {
	var longestMatch *config.URL
	for _, u := range urls {
		if !strings.HasPrefix(remoteURL, u.InsteadOf) {
			continue
		}

		// According to the spec if there is more than one match, the longest wins
		if longestMatch == nil || len(longestMatch.InsteadOf) < len(u.InsteadOf) {
			longestMatch = u
		}
	}

	return longestMatch
}

func applyInsteadOf(u *config.URL, url string) string {
	if !strings.HasPrefix(url, u.InsteadOf) {
		return url
	}

	return u.Name + url[len(u.InsteadOf):]
}
