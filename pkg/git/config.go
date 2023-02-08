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
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/imdario/mergo"
	"github.com/muja/goconfig"
)

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

func GetConfig(s config.Scope) (*config.Config, error) {
	files, err := config.Paths(s)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		f, err := osfs.Default.Open(file)
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

	localCfg, _ := repo.Config()
	err = mergo.Merge(c, localCfg, mergo.WithOverride)
	if err != nil {
		return nil, fmt.Errorf("couldn't merge local git config: %#v", err)
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

		// According to spec if there is more than one match, take the logest
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
