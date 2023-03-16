package report

import "github.com/listendev/lstn/pkg/listen"

type Report interface {
	Render(packages []listen.Package) error
}

type ReportBuilder struct {
	reports []Report
}

func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{}
}

func (b *ReportBuilder) RegisterReport(r Report) {
	b.reports = append(b.reports, r)
}

func (b *ReportBuilder) Render(packages []listen.Package) error {
	for _, r := range b.reports {
		if err := r.Render(packages); err != nil {
			return err
		}
	}

	return nil
}
