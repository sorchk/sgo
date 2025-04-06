package main

import (
	"strings"
)

// isCommandAllowed 检查命令是否允许执行
func (p *ShellPlugin) isCommandAllowed(cmd string) bool {
	// 如果没有设置允许的命令，则允许所有命令
	if len(p.allowedCommands) == 0 {
		return true
	}

	// 检查命令是否在允许列表中
	for _, allowedCmd := range p.allowedCommands {
		if strings.HasPrefix(cmd, allowedCmd) {
			return true
		}
	}

	return false
}
