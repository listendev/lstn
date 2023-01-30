package groups

import (
	"github.com/spf13/cobra"
)

var Core = cobra.Group{
	ID:    "core",
	Title: "Core Commands",
}
