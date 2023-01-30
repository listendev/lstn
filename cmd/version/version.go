package version

import (
	"context"
	"fmt"

	pkgcontext "github.com/listendev/lstn/pkg/context"
	"github.com/listendev/lstn/pkg/version"
	"github.com/spf13/cobra"
)

func New(ctx context.Context) (*cobra.Command, error) {
	var c = &cobra.Command{
		Use:                   "version",
		Short:                 "Print out version information",
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()

			// Obtain the (short) version info from the context
			shortVersion := ctx.Value(pkgcontext.ShortVersionKey).(string)

			outputString := fmt.Sprintf("lstn %s", shortVersion)
			changelogURL, _ := version.Changelog(shortVersion)
			if changelogURL != "" {
				outputString += fmt.Sprintf("\n%s", changelogURL)
			}

			fmt.Println(outputString)

			return nil
		},
	}

	return c, nil
}
