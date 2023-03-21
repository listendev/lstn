package reporter

import (
	"context"
	"errors"

	"github.com/google/go-github/v50/github"
	"github.com/listendev/lstn/pkg/reporter/githubprreview"
	"github.com/listendev/lstn/pkg/reporter/request"
)

var (
	ErrReporterNotFound = errors.New("reporter not found")
)

type Reporter interface {
	Report(req *request.Report) error
	WithContext(ctx context.Context)
	WithGithubClient(client *github.Client)
}

func BuildReporter(reporterIdentifier string) (Reporter, error) {
	switch reporterIdentifier {
	case githubprreview.ReporterIdentifier:
		return githubprreview.New(), nil
	default:
		return nil, ErrReporterNotFound
	}
}
