package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
)

// execCommand 执行单个命令
func (p *ShellPlugin) execCommand(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: exec <command>")
	}

	// 获取命令和参数
	cmdStr := strings.Join(args, " ")

	// 检查命令是否允许执行
	if !p.isCommandAllowed(cmdStr) {
		return fmt.Errorf("command not allowed: %s", cmdStr)
	}

	// 创建命令
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}

	// 设置工作目录
	cmd.Dir = p.workingDir

	// 设置标准输入输出
	cmd.Stdin = input
	cmd.Stdout = output
	cmd.Stderr = output

	// 执行命令
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// interactiveShell 交互式Shell
func (p *ShellPlugin) interactiveShell(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	// 获取Shell程序
	var shellCmd string
	var shellArgs []string

	if runtime.GOOS == "windows" {
		shellCmd = "cmd"
		shellArgs = []string{"/Q"}
	} else {
		shellCmd = "sh"
		shellArgs = []string{"-i"}
	}

	// 创建命令
	cmd := exec.CommandContext(ctx, shellCmd, shellArgs...)

	// 设置工作目录
	cmd.Dir = p.workingDir

	// 创建管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// 处理输入
	go func() {
		defer stdin.Close()
		io.Copy(stdin, input)
	}()

	// 处理输出
	go func() {
		io.Copy(output, stdout)
	}()

	// 处理错误
	go func() {
		io.Copy(output, stderr)
	}()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("shell execution failed: %w", err)
	}

	return nil
}
