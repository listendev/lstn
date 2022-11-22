package listen

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func Listen(ctx context.Context, r Request, rawResponseOnly bool) (*Response, []byte, error) {
	pl, err := json.Marshal(r)
	if err != nil {
		return nil, nil, err
	}

	// TODO > get the endpoint URL from config/env/flag
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "localhost:3030", bytes.NewBuffer(pl))
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
