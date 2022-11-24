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
	endpointURL := fmt.Sprintf("%s/api/analyse/npm", cfgOpts.Endpoint)

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

	if rawResponseOnly {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}
		return nil, b, nil
	}

	target := &Response{}
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return nil, nil, err
	}

	return target, nil, nil
}
