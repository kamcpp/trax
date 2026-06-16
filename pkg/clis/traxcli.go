package cli

import (
	"github.com/xshyft/trax/pkg/clis/traxcli"
)

// RunTraxCliWithTraceId starts the interactive trax CLI with optional trace ID
func RunTraxCliWithTraceId(traceId string) {
	ctx := traxcli.Init()
	traxcli.RunInteractive(ctx, traceId)
}

// RunTraxCliCommand executes a single traxcli command non-interactively
func RunTraxCliCommand(command string, args []string) {
	RunTraxCliCommandWithTraceId(command, args, "")
}

// RunTraxCliCommandWithTraceId executes a single traxcli command non-interactively with trace ID
func RunTraxCliCommandWithTraceId(command string, args []string, traceId string) {
	ctx := traxcli.Init()
	traxcli.RunNonInteractive(ctx, command, args, traceId)
}
