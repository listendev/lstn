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
package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/ua"
	"golang.org/x/exp/maps"
)

// GetFromRegistry asks to the npm registry for the details of a package
// by name, and optionally, by version.
func GetFromRegistry(ctx context.Context, name, version string) (io.ReadCloser, string, error) {
	// Obtain the local options from the context
	opts, err := pkgcontext.GetOptionsFromContext(ctx, pkgcontext.ConfigKey)
	if err != nil {
		return nil, "", fmt.Errorf("couldn't find the registry key in the configuration")
	}
	cfgFlags, ok := opts.(*flags.ConfigFlags)
	if !ok {
		return nil, "", fmt.Errorf("couldn't find the registry configuration")
	}
	npmRegistryBaseURL := cfgFlags.Registry.NPM

	if name == "" {
		return nil, npmRegistryBaseURL, pkgcontext.OutputError(ctx, fmt.Errorf("the name is mandatory to query the npm registry"))
	}
	suffix := name
	if version != "" {
		suffix += fmt.Sprintf("/%s", version)
	}

	url := fmt.Sprintf("%s/%s", npmRegistryBaseURL, suffix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, npmRegistryBaseURL, pkgcontext.OutputErrorf(ctx, err, "couldn't prepare the request to %s", url)
	}

	req.Header.Set("User-Agent", ua.Generate(true))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, npmRegistryBaseURL, pkgcontext.OutputErrorf(ctx, err, "couldn't perform the request to %s", req.URL)
	}

	if res.StatusCode != http.StatusOK {
		return nil, npmRegistryBaseURL, pkgcontext.OutputErrorf(ctx, err, "the NPM registry response for %s was not ok", req.URL)
	}

	return res.Body, npmRegistryBaseURL, nil
}

func GetVersionsFromRegistry(ctx context.Context, name string, constraints *semver.Constraints) (semver.Collection, error) {
	body, URL, err := GetFromRegistry(ctx, name, "")
	if URL == "" {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("package %s doesn't exist on registry %s", name, URL)
	}

	return GetVersionsFromRegistryResponse(body, constraints)
}

func GetVersionsFromRegistryResponse(body io.ReadCloser, constraints *semver.Constraints) (semver.Collection, error) {
	defer body.Close()
	type response struct {
		Name     string                 `json:"name"`
		Versions map[string]interface{} `json:"versions"`
	}
	res := &response{}

	if err := json.NewDecoder(body).Decode(&res); err != nil {
		return nil, fmt.Errorf("couldn't decode the registry response")
	}

	raw := maps.Keys(res.Versions)

	versions := semver.Collection{}
	for _, r := range raw {
		v, err := semver.NewVersion(r)
		if err != nil {
			return nil, fmt.Errorf("couln't parse version %s for package %s", r, res.Name)
		}

		if constraints == nil || constraints.Check(v) {
			versions = append(versions, v)
		}
	}
	sort.Sort(versions)

	return versions, nil
}
