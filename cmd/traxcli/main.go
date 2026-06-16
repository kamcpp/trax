package main

import (
	"fmt"
	"os"

	traxclicmd "github.com/xshyft/trax/cmd/agora/clis/traxcli"
)

func main() {
	cmd := traxclicmd.NewTraxCli()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
