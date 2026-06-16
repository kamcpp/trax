package traxctrl

import (
	"github.com/spf13/cobra"

	"github.com/xshyft/trax/pkg/daemons"
)

func NewTraxCtrlCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "traxctrl",
		Short: "starts agora trax-ctrl daemon",
		Long:  "starts agora trax-ctrl daemon",
		Run: func(cmd *cobra.Command, args []string) {
			useInMemory, _ := cmd.Flags().GetBool("in-memory-store")
			daemons.RunTraxCtrl(useInMemory)
		},
	}

	cmd.Flags().Bool("in-memory-store", false, "use in-memory store instead of PostgreSQL")

	return cmd
}
