package main

import (
	"context"
	"fmt"
	"io"
)

// GetCommands 获取支持的命令列表
func (p *TerminalPlugin) GetCommands() []string {
	return []string{
		"create",
		"list",
		"kill",
		"resize",
		"write",
		"read",
	}
}

// Execute 执行命令
func (p *TerminalPlugin) Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "create":
		return p.createTerminal(ctx, cmdArgs, input, output)
	case "list":
		return p.listTerminals(ctx, output)
	case "kill":
		return p.killTerminal(ctx, cmdArgs, output)
	case "resize":
		return p.resizeTerminal(ctx, cmdArgs, output)
	case "write":
		return p.writeToTerminal(ctx, cmdArgs, input, output)
	case "read":
		return p.readFromTerminal(ctx, cmdArgs, output)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
