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
package npm

import (
	"context"
	"runtime"
	"sync"
	"time"
)

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

type packageInfo struct {
	name    string
	version string
	shasum  string
}

type packageResult struct {
	info *packageInfo
	err  error
}
type resultsFlow <-chan packageResult
type eachDepCallback func(string, string) (*packageInfo, error)

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

// Shasums concurrently queries the NPM registry to get the shasums
// of all the receiving packageLockJSON dependencies.
//
// Note this function is blocking because it waits for all the queries to complete.
//
// It early exits in case the timeout was exceeded, the context got cancelled,
// or when one query was not successful.
func (p *packageLockJSON) Shasums(ctx context.Context, timeout time.Duration) (Packages, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Concurrently run queries over the CPUs
	resultsChannel := p.
		produceDependencies().
		each(ctx, runtime.NumCPU(), func(name, version string) (*packageInfo, error) {
			// NOTE > can call cancel() here if you wanna stop the process
			return requestShasum(ctx, name, version)
		})

	// Wait for results to come...
	packages := make(Packages)
	for res := range resultsChannel {
		if res.info != nil {
			packages[res.info.name] = Package{
				Version: res.info.version,
				Shasum:  res.info.shasum,
			}
		}
		if res.err != nil {
			return nil, res.err
		}
	}

	return packages, nil
}
