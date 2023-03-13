package packagesprinter

import "github.com/listendev/lstn/pkg/listen"

type PackagesPrinter interface {
	RenderPackages(pkgs *listen.Response) error
}
