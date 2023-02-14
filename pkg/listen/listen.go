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
package listen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/ua"
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

func getAPIPrefix(baseURL string) string {
	if strings.HasPrefix(baseURL, "http://127.0.0.1") || strings.HasPrefix(baseURL, "http://localhost") {
		return "/api/npm"
	}

	return "/api"
}

func getEndpointURLFromContext[T any](r T, o *options) (string, error) {
	segment := ""
	switch any(r).(type) {
	case *AnalysisRequest:
		segment = "analysis"
	case *VerdictsRequest:
		segment = "verdicts"
	default:
		return "", fmt.Errorf("unsupported request type")
	}

	if o.baseURL == "" {
		var err error
		o.baseURL, err = getBaseURLFromContext(o.ctx)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s%s/%s", o.baseURL, getAPIPrefix(o.baseURL), segment), nil
}

func Packages[T Request](r T, opts ...func(*options)) (*Response, []byte, error) {
	o, err := newOptions(opts...)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	// Obtain the endpoint base URL
	endpointURL, err := getEndpointURLFromContext(r, o)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	// Prepare the request
	pl, err := json.Marshal(r)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}
	req, err := http.NewRequestWithContext(o.ctx, http.MethodPost, endpointURL, bytes.NewBuffer(pl))
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", ua.Generate(true))

	// Send the request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	// Bail out if status != 200
	if res.StatusCode != http.StatusOK {
		target := &responseError{}
		if err = dec.Decode(target); err != nil {
			return nil, nil, pkgcontext.OutputError(o.ctx, err)
		}

		return nil, nil, pkgcontext.OutputErrorf(o.ctx, err, target.Message)
	}

	return response(dec, res, o)
}

func response(dec *json.Decoder, res *http.Response, o *options) (*Response, []byte, error) {
	if o.json.IsJSON() {
		// Eventually return as JSON
		out := new(bytes.Buffer)
		if err := o.json.GetOutput(o.ctx, res.Body, out); err != nil {
			return nil, nil, pkgcontext.OutputError(o.ctx, err)
		}

		return nil, out.Bytes(), nil
	}
	// Alternatively, unmarshal the JSON body into a Response
	target := &Response{}
	if err := dec.Decode(target); err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	return target, nil, nil
}
