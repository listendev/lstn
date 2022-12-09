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

	"github.com/listendev/lstn/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
)

func Listen(ctx context.Context, r *Request, rawResponseOnly bool) (*Response, []byte, error) {
	pl, err := json.Marshal(r)
	if err != nil {
		return nil, nil, err
	}

	cfgOpts, ok := ctx.Value(pkgcontext.ConfigKey).(*flags.ConfigOptions)
	if !ok {
		return nil, nil, fmt.Errorf("couldn't obtain configuration options")
	}
	// Everything in the context has been already validated
	// So we assume it's safe to do not validate it again
	endpointURL := fmt.Sprintf("%s/api/analyse/npm?verbose=true", cfgOpts.Endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewBuffer(pl))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	// Bail out if status != 200
	if res.StatusCode != http.StatusOK {
		target := &responseError{}
		if err := dec.Decode(target); err != nil {
			return nil, nil, err
		}
		// For target.Reason to have a value the verbose query param above needs to be true
		return nil, nil, fmt.Errorf(target.Reason.Message)
	}

	// Return the JSON body
	if rawResponseOnly {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}
		return nil, b, nil
	}

	// Unmarshal the JSON body into a Response
	target := &Response{}
	if err := dec.Decode(target); err != nil {
		return nil, nil, err
	}
	return target, nil, nil
}
