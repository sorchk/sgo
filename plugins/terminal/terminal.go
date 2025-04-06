package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"
)

// createTerminal 创建新终端
func (p *TerminalPlugin) createTerminal(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: create <request_json>")
	}

	// 解析请求
	var req CreateTerminalRequest
	if err := json.Unmarshal([]byte(args[0]), &req); err != nil {
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// 检查终端ID是否已存在
	p.terminalsMu.RLock()
	_, exists := p.terminals[req.ID]
	p.terminalsMu.RUnlock()
	if exists {
		return fmt.Errorf("terminal with ID %s already exists", req.ID)
	}

	// 确定要执行的命令
	command := req.Command
	cmdArgs := req.Args

	// 如果没有指定命令，使用默认Shell
	if command == "" {
		if runtime.GOOS == "windows" {
			command = "cmd"
		} else {
			command = "bash"
			cmdArgs = append([]string{"-i"}, cmdArgs...)
		}
	}

	// 创建上下文
	termCtx, termCancel := context.WithCancel(ctx)

	// 创建命令
	cmd := exec.CommandContext(termCtx, command, cmdArgs...)
	cmd.Dir = p.workingDir

	// 创建管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		termCancel()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		termCancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		termCancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		termCancel()
		return fmt.Errorf("failed to start command: %w", err)
	}

	// 创建终端实例
	terminal := &Terminal{
		ID:        req.ID,
		Command:   command,
		Args:      cmdArgs,
		CreatedAt: time.Now(),
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		ctx:       termCtx,
		cancel:    termCancel,
	}

	// 添加到终端列表
	p.terminalsMu.Lock()
	p.terminals[req.ID] = terminal
	p.terminalsMu.Unlock()

	// 监控命令执行
	go func() {
		// 等待命令完成
		cmd.Wait()

		// 从终端列表中移除
		p.terminalsMu.Lock()
		delete(p.terminals, req.ID)
		p.terminalsMu.Unlock()
	}()

	// 返回终端信息
	termInfo, err := json.Marshal(map[string]interface{}{
		"id":         terminal.ID,
		"command":    terminal.Command,
		"args":       terminal.Args,
		"created_at": terminal.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal terminal info: %w", err)
	}

	fmt.Fprintf(output, "%s\n", termInfo)
	return nil
}

// listTerminals 列出所有终端
func (p *TerminalPlugin) listTerminals(ctx context.Context, output io.Writer) error {
	p.terminalsMu.RLock()
	defer p.terminalsMu.RUnlock()

	// 构建终端列表
	terminals := make([]map[string]interface{}, 0, len(p.terminals))
	for _, term := range p.terminals {
		terminals = append(terminals, map[string]interface{}{
			"id":         term.ID,
			"command":    term.Command,
			"args":       term.Args,
			"created_at": term.CreatedAt,
		})
	}

	// 序列化终端列表
	termList, err := json.Marshal(terminals)
	if err != nil {
		return fmt.Errorf("failed to marshal terminal list: %w", err)
	}

	fmt.Fprintf(output, "%s\n", termList)
	return nil
}

// killTerminal 终止终端
func (p *TerminalPlugin) killTerminal(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: kill <terminal_id>")
	}

	terminalID := args[0]

	// 获取终端
	p.terminalsMu.Lock()
	terminal, exists := p.terminals[terminalID]
	if !exists {
		p.terminalsMu.Unlock()
		return fmt.Errorf("terminal with ID %s not found", terminalID)
	}

	// 取消上下文
	terminal.cancel()

	// 关闭管道
	terminal.stdin.Close()

	// 从终端列表中移除
	delete(p.terminals, terminalID)
	p.terminalsMu.Unlock()

	fmt.Fprintf(output, "{\"success\":true,\"id\":\"%s\"}\n", terminalID)
	return nil
}

// resizeTerminal 调整终端大小
func (p *TerminalPlugin) resizeTerminal(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: resize <terminal_id> <rows> <cols>")
	}

	terminalID := args[0]
	rows, err := parseInt(args[1])
	if err != nil {
		return fmt.Errorf("invalid rows value: %w", err)
	}
	cols, err := parseInt(args[2])
	if err != nil {
		return fmt.Errorf("invalid cols value: %w", err)
	}

	// 获取终端
	p.terminalsMu.RLock()
	_, exists := p.terminals[terminalID]
	p.terminalsMu.RUnlock()
	if !exists {
		return fmt.Errorf("terminal with ID %s not found", terminalID)
	}

	// 调整终端大小（仅在Unix系统上支持）
	if runtime.GOOS != "windows" {
		// 这里需要使用特定的系统调用来调整终端大小
		// 由于Go标准库没有直接提供这个功能，这里只是返回成功
		// 在实际实现中，可以使用syscall包或第三方库来实现
	}

	fmt.Fprintf(output, "{\"success\":true,\"id\":\"%s\",\"rows\":%d,\"cols\":%d}\n", terminalID, rows, cols)
	return nil
}

// writeToTerminal 向终端写入数据
func (p *TerminalPlugin) writeToTerminal(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: write <request_json>")
	}

	// 解析请求
	var req TerminalDataRequest
	if err := json.Unmarshal([]byte(args[0]), &req); err != nil {
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// 获取终端
	p.terminalsMu.RLock()
	terminal, exists := p.terminals[req.ID]
	p.terminalsMu.RUnlock()
	if !exists {
		return fmt.Errorf("terminal with ID %s not found", req.ID)
	}

	// 写入数据
	if _, err := terminal.stdin.Write([]byte(req.Data)); err != nil {
		return fmt.Errorf("failed to write to terminal: %w", err)
	}

	fmt.Fprintf(output, "{\"success\":true,\"id\":\"%s\",\"bytes_written\":%d}\n", req.ID, len(req.Data))
	return nil
}

// readFromTerminal 从终端读取数据
func (p *TerminalPlugin) readFromTerminal(ctx context.Context, args []string, output io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: read <terminal_id>")
	}

	terminalID := args[0]

	// 获取终端
	p.terminalsMu.RLock()
	terminal, exists := p.terminals[terminalID]
	p.terminalsMu.RUnlock()
	if !exists {
		return fmt.Errorf("terminal with ID %s not found", terminalID)
	}

	// 创建缓冲区
	stdoutBuf := make([]byte, 4096)
	stderrBuf := make([]byte, 4096)

	// 设置非阻塞读取的超时时间
	timeout := time.After(100 * time.Millisecond)

	// 读取标准输出
	stdoutCh := make(chan int, 1)
	go func() {
		n, _ := terminal.stdout.Read(stdoutBuf)
		stdoutCh <- n
	}()

	// 读取标准错误
	stderrCh := make(chan int, 1)
	go func() {
		n, _ := terminal.stderr.Read(stderrBuf)
		stderrCh <- n
	}()

	// 等待数据或超时
	var stdoutData, stderrData []byte
	select {
	case n := <-stdoutCh:
		if n > 0 {
			stdoutData = stdoutBuf[:n]
		}
	case n := <-stderrCh:
		if n > 0 {
			stderrData = stderrBuf[:n]
		}
	case <-timeout:
		// 超时，没有数据可读
	}

	// 返回读取的数据
	resp, err := json.Marshal(map[string]interface{}{
		"id":     terminalID,
		"stdout": string(stdoutData),
		"stderr": string(stderrData),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Fprintf(output, "%s\n", resp)
	return nil
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
