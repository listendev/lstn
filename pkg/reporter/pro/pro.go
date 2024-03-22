package pro

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
	"github.com/listendev/lstn/pkg/ci"
	"github.com/listendev/lstn/pkg/cmd/flags"
	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/listen"
	"github.com/listendev/lstn/pkg/reporter"
)

type rep struct {
	ctx  context.Context
	opts *flags.ConfigFlags
	info *ci.Info
}

func New(ctx context.Context, opts ...reporter.Option) (reporter.Reporter, error) {
	// Retrieve the config options from the context
	// Those are mandatory because they contain the GitHub reporting options
	cfgOpts, ok := ctx.Value(pkgcontext.ConfigKey).(*flags.ConfigFlags)
	if cfgOpts == nil || !ok {
		return nil, fmt.Errorf("couldn't retrieve the config options")
	}

	ret := &rep{
		ctx:  ctx,
		opts: cfgOpts,
	}

	for _, opt := range opts {
		ret = opt(ret).(*rep)
	}

	return ret, nil
}

func (r *rep) Run(res listen.Response) error {
	fmt.Println("TODO: for every listen.Package call the Dependency API")

	return nil
}

func (r *rep) WithConfigOptions(opts *flags.ConfigFlags) {
	r.opts = opts
}

func (r *rep) WithGitHubClient(client *github.Client) {
	// Do nothing
}

func (r *rep) WithContinuousIntegrationInfo(info *ci.Info) {
	r.info = info
}
