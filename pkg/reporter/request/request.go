package request

import (
	"github.com/listendev/lstn/pkg/listen"
)

type GithubPRReviewReportRequest struct {
	Owner string
	Repo  string
	ID    int
}

type Report struct {
	Packages              []listen.Package
	GithubPRReviewRequest GithubPRReviewReportRequest
}
