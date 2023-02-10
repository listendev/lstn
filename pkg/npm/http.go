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
	"fmt"
	"io"
	"net/http"

	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/ua"
)

const npmRegistryBaseURL = "https://registry.npmjs.org"

// GetFromRegistry asks to the npm registry for the details of a package
// by name, and optionally, by version.
func GetFromRegistry(ctx context.Context, name, version string) (io.ReadCloser, error) {
	if name == "" {
		return nil, pkgcontext.OutputError(ctx, fmt.Errorf("the name is mandatory to query the npm registry"))
	}
	suffix := fmt.Sprintf("/%s", name)
	if version != "" {
		suffix += fmt.Sprintf("/%s", version)
	}

	url := fmt.Sprintf("%s/%s", npmRegistryBaseURL, suffix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, pkgcontext.OutputErrorf(ctx, err, "couldn't prepare the request to %s", url)
	}

	req.Header.Set("User-Agent", ua.Generate(true))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, pkgcontext.OutputErrorf(ctx, err, "couldn't perform the request to %s", req.URL)
	}

	if res.StatusCode != http.StatusOK {
		return nil, pkgcontext.OutputErrorf(ctx, err, "the NPM registry response for %s was not ok", req.URL)
	}

	return res.Body, nil
}
