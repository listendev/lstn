package npm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

const npmRegistryBaseURL = "https://registry.npmjs.org"

type npmRegistryPackageVersionResponse struct {
	Dist struct {
		Shasum string
	}
}

func errorOut(input error, format string, a ...any) error {
	if errors.Is(input, context.Canceled) {
		return context.Canceled
	}
	if e, ok := input.(net.Error); ok && e.Timeout() {
		return context.DeadlineExceeded
	}
	return fmt.Errorf(format, a...)
}

// requestShasum queries the NPM registry to obtain the shasum of the input package version.
// TODO > backoff/retry strategy?
func requestShasum(ctx context.Context, name, version string) (*packageInfo, error) {
	url := fmt.Sprintf("%s/%s/%s", npmRegistryBaseURL, name, version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errorOut(err, "couldn't prepare the request to %s for %s/%s", npmRegistryBaseURL, name, version)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errorOut(err, "couldn't perform the request to %s", req.URL)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errorOut(err, "the NPM registry response for %s was not ok", req.URL)
	}

	ret := &npmRegistryPackageVersionResponse{}
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, errorOut(err, "couldn't decode the NPM registry response")
	}

	return &packageInfo{shasum: ret.Dist.Shasum, name: name, version: version}, nil
}
