package packagestracker

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/jedib0t/go-pretty/text"
	"github.com/listendev/lstn/pkg/cmd/iostreams"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/npm"
)

type PackagesRetrievalFunc func(depName string, depVersion *semver.Version) (*listen.Response, error)
type PackagesByType map[npm.DependencyType]map[string]*semver.Version

func TrackPackages(
	ctx context.Context,
	deps PackagesByType,
	packageRetrievalFunc PackagesRetrievalFunc) (*listen.Response, error) {

	io := ctx.Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)

	io.StartProgressTracking()
	defer io.StopProgressTracking()

	// Process one dependency set at once
	combinedResponse := []listen.Package{}

	for depType, currentDeps := range deps {
		depTracker := io.CreateProgressTracker(depType.Name(), int64(len(currentDeps)))

		for depName, depVersion := range currentDeps {
			io.LogProgress(text.Faint.Sprintf("processing %s %s", depName, depVersion))

			res, err := packageRetrievalFunc(depName, depVersion)

			if err != nil {
				depTracker.IncrementWithError(1)
				continue
			}

			if res != nil {
				combinedResponse = append(combinedResponse, *res...)
			}
			depTracker.Increment(1)
		}
	}

	return (*listen.Response)(&combinedResponse), nil
}
