package packagestracker

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/jedib0t/go-pretty/text"
	"github.com/listendev/lstn/pkg/cmd/iostreams"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/npm"
)

type PackagesRetrievalFunc func(depName string, depVersion *semver.Version) (*listen.Response, error)

type Dependency struct {
	Name    string
	Version *semver.Version
}

type ListableDependency interface {
	~[]Dependency | ~map[string]*semver.Version
	List() []Dependency
}

type MapOfDependencies map[string]*semver.Version

func (m MapOfDependencies) List() []Dependency {
	list := []Dependency{}

	for name, version := range m {
		list = append(list, Dependency{
			Name:    name,
			Version: version,
		})
	}

	return list
}

type ListOfDependencies []Dependency

func (l ListOfDependencies) List() []Dependency {
	return l
}

func ConvertToMapOfDependencies[K npm.DependencyType | string](deps map[K]map[string]*semver.Version) map[K]MapOfDependencies {
	md := map[K]MapOfDependencies{}
	for depType, d := range deps {
		md[depType] = MapOfDependencies(d)
	}
	return md
}

func processingMessage(dep Dependency) string {
	if dep.Version == nil {
		return text.Faint.Sprintf("processing %s", dep.Name)
	}
	return text.Faint.Sprintf("processing %s %s", dep.Name, dep.Version)
}

func processingErrorMessage(dep Dependency, err error) string {
	if dep.Version == nil {
		return text.Faint.Sprintf("%s: error processing %s: %s", dep.Name, err.Error())
	}

	return text.Faint.Sprintf("error processing %s %s: %s", dep.Name, dep.Version, err.Error())
}

func TrackPackages[K npm.DependencyType | string, D ListableDependency](
	ctx context.Context,
	deps map[K]D,
	packageRetrievalFunc PackagesRetrievalFunc) (*listen.Response, error) {

	io := ctx.Value(pkgcontext.IOStreamsKey).(*iostreams.IOStreams)

	io.StartProgressTracking()
	defer io.StopProgressTracking()

	// Process one dependency set at once
	combinedResponse := []listen.Package{}
	cs := io.ColorScheme()
	for depType, currentDeps := range deps {
		depTracker := io.CreateProgressTracker(fmt.Sprintf("%s", depType), int64(len(currentDeps)))

		for _, dep := range currentDeps.List() {
			io.LogProgress(processingMessage(dep))

			res, err := packageRetrievalFunc(dep.Name, dep.Version)

			if err != nil {
				io.LogProgress(fmt.Sprintf("%s: %s", cs.FailureIconWithColor(cs.Red), processingErrorMessage(dep, err)))
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
