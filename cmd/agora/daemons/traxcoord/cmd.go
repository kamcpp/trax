package traxcoord

import (
	"github.com/spf13/cobra"

	"github.com/xshyft/trax/pkg/daemons"
)

func NewTraxCoordinatorCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "traxcoord",
		Short: "starts agora trax-coordinator daemon",
		Long:  "starts agora trax-coordinator daemon",
		Run: func(cmd *cobra.Command, args []string) {
			daemons.RunTraxCoordinator()
		},
	}
	return cmd
}
