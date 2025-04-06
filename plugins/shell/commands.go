package main

import (
	"context"
	"fmt"
	"io"
)

// GetCommands 获取支持的命令列表
func (p *ShellPlugin) GetCommands() []string {
	return []string{
		"exec",
		"interactive",
	}
}

// Execute 执行命令
func (p *ShellPlugin) Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "exec":
		return p.execCommand(ctx, cmdArgs, input, output)
	case "interactive":
		return p.interactiveShell(ctx, cmdArgs, input, output)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
