/*
Copyright © 2022 The listen.dev team <engineering@garnet.ai>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package listen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-querystring/query"
	"github.com/listendev/lstn/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
)

func getBaseURLFromContext(ctx context.Context) (string, error) {
	cfgOpts, ok := ctx.Value(pkgcontext.ConfigKey).(*flags.ConfigOptions)
	if !ok {
		return "", fmt.Errorf("couldn't obtain configuration options")
	}
	// Everything in the context has been already validated
	// So we assume it's safe to do not validate it again
	return cfgOpts.Endpoint, nil
}

func PackageLockAnalysis(ctx context.Context, r *AnalysisRequest, rawResponseOnly bool) (*Response, []byte, error) {
	// Obtain the endpoint base URL
	baseURL, err := getBaseURLFromContext(ctx)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	endpointURL := fmt.Sprintf("%s/api/analysis", baseURL)

	// Prepare the request
	pl, err := json.Marshal(r)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewBuffer(pl))
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send the request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	// Bail out if status != 200
	if res.StatusCode != http.StatusOK {
		target := &responseError{}
		if err := dec.Decode(target); err != nil {
			return nil, nil, pkgcontext.OutputError(ctx, err)
		}
		// For target.Reason to have a value the verbose query param above needs to be true
		return nil, nil, pkgcontext.OutputErrorf(ctx, err, target.Message)
	}

	// Return the JSON body
	if rawResponseOnly {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, pkgcontext.OutputError(ctx, err)
		}
		return nil, b, nil
	}

	// Unmarshal the JSON body into a Response
	target := &Response{}
	if err := dec.Decode(target); err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}

	return target, nil, nil
}

func PackageVerdicts(ctx context.Context, r *VerdictsRequest, rawResponseOnly bool) (*Response, []byte, error) {
	// Obtain the endpoint base URL
	baseURL, err := getBaseURLFromContext(ctx)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	endpointURL := fmt.Sprintf("%s/api/verdicts", baseURL)

	// Prepare the request
	val, err := query.Values(r)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.URL.RawQuery = val.Encode()

	// Send the request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	// Bail out if status != 200
	if res.StatusCode != http.StatusOK {
		target := &responseError{}
		if err := dec.Decode(target); err != nil {
			return nil, nil, pkgcontext.OutputError(ctx, err)
		}
		// For target.Reason to have a value the verbose query param above needs to be true
		return nil, nil, pkgcontext.OutputErrorf(ctx, err, target.Message)
	}

	// Return the JSON body
	if rawResponseOnly {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, pkgcontext.OutputError(ctx, err)
		}
		return nil, b, nil
	}

	// Unmarshal the JSON body into a Response
	target := &Response{}
	if err := dec.Decode(target); err != nil {
		return nil, nil, pkgcontext.OutputError(ctx, err)
	}

	return target, nil, nil
}
