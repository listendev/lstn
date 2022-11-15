package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type packageLockDependency struct {
	Version  string `json:"version"`
	Resolved string `json:"resolved"`
}

type packageLockJSON struct {
	Name            string                           `json:"name"`
	Version         string                           `json:"version"`
	LockfileVersion int                              `json:"lockfileVersion"`
	Dependencies    map[string]packageLockDependency `json:"dependencies"`
}

type dep struct {
	name    string
	version string
}

type depsFlow <-chan dep

// produceDependencies emits the dependencies of the
// receiving packageLockJSON instance on a channel.
func (p *packageLockJSON) produceDependencies() depsFlow {
	depsChannel := make(chan dep)

	go func() {
		for k, v := range p.Dependencies {
			depsChannel <- dep{
				name:    k,
				version: v.Version,
			}
		}
		close(depsChannel)
	}()

	return depsChannel
}

type packageResult struct {
	info *Package
	err  error
}
type resultsFlow <-chan packageResult
type eachDepCallback func(string, string) (*Package, error)

// each emits the results of the eachDepCallback callback
// (together with the corresponding receiving dependency from deps).
//
// It uses a pool of N Go routines to process the listen for deps
// dependencies and execute the callback on them.
//
// In case the context gets cancelled or its timeout gets exceeded,
// this function exits early.
func (deps depsFlow) each(ctx context.Context, numWorkers int, fn eachDepCallback) resultsFlow {
	outChannel := make(chan packageResult)

	go func() {
		defer close(outChannel)
		var wg sync.WaitGroup

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)

			// Read the dependencies concurrently
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						outChannel <- packageResult{err: ctx.Err()}
						return
					default:
					}
					// Using a double select to avoid a workder reading from a channel instead of receiving the cancellation signal
					// See: https://stackoverflow.com/a/46202533
					select {
					case <-ctx.Done():
						outChannel <- packageResult{err: ctx.Err()}
						return
					case dep, hasMore := <-deps:
						// Exit at completion
						if !hasMore {
							return
						}
						res, err := fn(dep.name, dep.version)
						// Early exit in case the callback returns an error
						if err != nil {
							outChannel <- packageResult{err: err}
							return
						}
						// Send the result into the output channel
						outChannel <- packageResult{info: res}
					}
				}
			}()
		}
		// Wait each of the N worker inner goroutines to complete
		wg.Wait()
	}()

	return outChannel
}

type PackageLockJSON interface {
	QueryShasums(ctx context.Context, timeout time.Duration) ([]*Package, error)
	// Dependencies() // TODO
}

type Package struct {
	Name    string
	Version string
	Shasum  string
}

// QueryShasums concurrently queries the NPM registry to get the shasums
// of all the receiving packageLockJSON dependencies.
//
// Note this function is blocking because it waits for all the queries to complete.
//
// It early exits in case the timeout was exceeded, the context got cancelled,
// or when one query was not successful.
func (p *packageLockJSON) QueryShasums(ctx context.Context, timeout time.Duration) ([]*Package, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Concurrently run queries over the CPUs
	resultsChannel := p.
		produceDependencies().
		each(ctx, runtime.NumCPU(), func(name, version string) (*Package, error) {
			// NOTE > can call cancel() here if you wanna stop the process
			return requestShasum(ctx, name, version)
		})

	// Wait for results to come...
	packages := []*Package{}
	for res := range resultsChannel {
		if res.info != nil {
			packages = append(packages, res.info)
		}
		if res.err != nil {
			return nil, res.err
		}
	}

	return packages, nil
}

// NewPackageLockJSON is a factory to create an empty PackageLockJSON.
func NewPackageLockJSON() PackageLockJSON {
	ret := &packageLockJSON{}
	return ret
}

// NewPackageLockJSONFrom creates a PackageLockJSON from the contents of a package-lock.json file.
func NewPackageLockJSONFrom(bytes []byte) (PackageLockJSON, error) {
	ret := NewPackageLockJSON()
	if err := json.Unmarshal(bytes, ret); err != nil {
		return nil, fmt.Errorf("couldn't instantiate from the input package-lock.json contents")
	}
	return ret, nil
}
