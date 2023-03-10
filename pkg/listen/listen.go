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
	"runtime"
	"strings"

	"github.com/XANi/goneric"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/ua"
)

func getBaseURLFromContext(ctx context.Context) (string, error) {
	cfg, ok := ctx.Value(pkgcontext.ConfigKey).(*flags.ConfigFlags)
	if !ok {
		return "", fmt.Errorf("couldn't obtain configuration options")
	}
	// Everything in the context has been already validated
	// So we assume it's safe to do not validate it again
	return cfg.Endpoint, nil
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

func response(dec *json.Decoder, o *options) (*Response, []byte, error) {
	// Unmarshal the JSON body into a Response
	target := &Response{}
	if err := dec.Decode(target); err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	if o.json.IsJSON() {
		allJSON := new(bytes.Buffer)
		if err := json.NewEncoder(allJSON).Encode(target); err != nil {
			return nil, nil, pkgcontext.OutputError(o.ctx, fmt.Errorf("couldn't JSON encode the response"))
		}

		// Eventually filter the JSON
		out := new(bytes.Buffer)
		if err := o.json.GetOutput(o.ctx, allJSON, out); err != nil {
			return nil, nil, pkgcontext.OutputError(o.ctx, err)
		}

		return nil, out.Bytes(), nil
	}

	return target, nil, nil
}

func request[T Request](r T, o *options, endpointURL, userAgent string) (*json.Decoder, *http.Response, error) {
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
	// Automatically generate user agent
	if userAgent == "" {
		userAgent = ua.Generate(true)
	}
	req.Header.Set("User-Agent", userAgent)

	// Send the request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	dec := json.NewDecoder(res.Body)

	// Bail out if status != 200
	if res.StatusCode != http.StatusOK {
		target := &responseError{}
		if err = dec.Decode(target); err != nil {
			return nil, nil, pkgcontext.OutputError(o.ctx, err)
		}

		return nil, nil, pkgcontext.OutputErrorf(o.ctx, err, target.Message)
	}

	return dec, res, nil
}

func Packages[T Request](r T, opts ...func(*options)) (*Response, []byte, error) {
	o, err := newOptions(opts...)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	endpointURL, err := getEndpointURLFromContext(r, o)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	dec, res, err := request(r, o, endpointURL, "")
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	return response(dec, o)
}

func BulkPackages(requests []*VerdictsRequest, opts ...func(*options)) (*Response, []byte, error) {
	o, err := newOptions(opts...)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	numPackages := len(requests)
	if numPackages == 0 {
		return nil, nil, pkgcontext.OutputError(o.ctx, fmt.Errorf("empty requests set"))
	}

	// TODO: validate that all requests are ok

	endpointURL, err := getEndpointURLFromContext(requests[0], o)
	if err != nil {
		return nil, nil, pkgcontext.OutputError(o.ctx, err)
	}

	userAgent := ua.Generate(true)

	type returnWrap struct {
		res *Package
		err error
	}

	cb := func(req *VerdictsRequest) returnWrap {
		dec, res, reqErr := request(req, o, endpointURL, userAgent)
		if reqErr != nil {
			return returnWrap{nil, reqErr}
		}
		defer res.Body.Close()

		ret := Response{}
		if decodeErr := dec.Decode(&ret); decodeErr != nil {
			return returnWrap{nil, decodeErr}
		}

		// It's impossible to have more that one Package in every Response (a list of Package items) in this case
		// Why? Because every VerdictsRequest contains an exact pacakge version
		// So we only grab the first element of the Response
		return returnWrap{&ret[0], nil}
	}

	returns := goneric.ParallelMapSlice(cb, runtime.NumCPU(), requests)

	numReturns := len(returns)
	if numReturns != numPackages {
		return nil, nil, pkgcontext.OutputError(o.ctx, fmt.Errorf("wrong number of responses: %d responses for %d requests", numReturns, numPackages))
	}

	res := []Package{}
	for _, ret := range returns {
		if ret.err == nil {
			res = append(res, *ret.res)
		} else {
			return nil, nil, pkgcontext.OutputError(o.ctx, err)
		}
	}
	if o.json.IsJSON() {
		allJSON := new(bytes.Buffer)
		if err := json.NewEncoder(allJSON).Encode(res); err != nil {
			return nil, nil, pkgcontext.OutputError(o.ctx, fmt.Errorf("couldn't JSON encode the response"))
		}

		// Eventually filter the JSON
		out := new(bytes.Buffer)
		if err := o.json.GetOutput(o.ctx, allJSON, out); err != nil {
			return nil, nil, pkgcontext.OutputError(o.ctx, err)
		}

		return nil, out.Bytes(), nil
	}
	cast := (Response)(res)

	return &cast, nil, nil
}
