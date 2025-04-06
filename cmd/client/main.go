package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sorc/tcpserver/internal/crypto"
	"github.com/sorc/tcpserver/pkg/protocol"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	ServerAddr string `json:"server_addr"`
	ClientID   string `json:"client_id"`
	Secret     string `json:"secret"`
}

// Client 客户端
type Client struct {
	config    ClientConfig
	conn      net.Conn
	sessionID string
	cipher    *crypto.XXTEACipher
}

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "client.json", "Path to config file")
	flag.Parse()

	// 读取配置文件
	configData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// 解析配置
	var config ClientConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// 创建客户端
	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 连接服务器
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	// 认证
	if err := client.Authenticate(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("Connected to server and authenticated successfully.")
	fmt.Println("Type 'help' for available commands.")

	// 命令行交互
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		if line == "exit" || line == "quit" {
			break
		}

		if line == "help" {
			printHelp()
			continue
		}

		// 解析命令
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 2 {
			fmt.Println("Invalid command format. Use: <plugin> <command> [args]")
			continue
		}

		plugin := parts[0]
		command := parts[1]
		args := ""
		if len(parts) > 2 {
			args = parts[2]
		}

		// 在新的goroutine中执行命令，设置超时
		go func() {
			// 创建超时通道
			timeoutCh := make(chan bool, 1)
			resultCh := make(chan error, 1)

			// 在新的goroutine中执行命令
			go func() {
				err := client.ExecuteCommand(plugin, command, args)
				resultCh <- err
			}()

			// 设置10秒超时
			go func() {
				time.Sleep(10 * time.Second)
				timeoutCh <- true
			}()

			// 等待命令执行完成或超时
			select {
			case err := <-resultCh:
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			case <-timeoutCh:
				fmt.Println("Command execution timed out. You can continue to use the client.")
			}
		}()
	}
}

// NewClient 创建新的客户端
func NewClient(config ClientConfig) (*Client, error) {
	// 创建加密器
	cipher, err := crypto.NewXXTEACipher([]byte(config.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	return &Client{
		config: config,
		cipher: cipher,
	}, nil
}

// Connect 连接服务器
func (c *Client) Connect() error {
	// 连接服务器
	conn, err := net.Dial("tcp", c.config.ServerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	c.conn = conn

	return nil
}

// Close 关闭连接
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Authenticate 认证
func (c *Client) Authenticate() error {
	// 生成随机数
	nonce := uuid.New().String()
	timestamp := time.Now().Unix()

	// 生成签名
	h := hmac.New(sha256.New, []byte(c.config.Secret))
	h.Write([]byte(fmt.Sprintf("%s:%s:%d", c.config.ClientID, nonce, timestamp)))
	signature := hex.EncodeToString(h.Sum(nil))

	// 创建认证请求
	authMsg, err := protocol.NewAuthRequestMessage(
		uuid.New().String(),
		c.config.ClientID,
		nonce,
		timestamp,
		signature,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	// 发送认证请求
	if err := protocol.WriteMessage(c.conn, authMsg); err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	// 读取认证响应
	respMsg, err := protocol.ReadMessage(c.conn)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	// 检查响应类型
	if respMsg.Header.Type != protocol.AuthResponse {
		return fmt.Errorf("unexpected response type: %d", respMsg.Header.Type)
	}

	// 解析认证响应
	var authResp protocol.AuthResponseBody
	if err := json.Unmarshal(respMsg.Body, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	// 检查认证结果
	if !authResp.Success {
		return fmt.Errorf("authentication failed: %s", authResp.Message)
	}

	// 保存会话ID
	c.sessionID = authResp.SessionID

	return nil
}

// ExecuteCommand 执行命令
func (c *Client) ExecuteCommand(plugin, command string, args string) error {
	// 创建命令请求
	cmdArgs := []string{}
	if args != "" {
		cmdArgs = strings.Split(args, " ")
	}

	fmt.Printf("Executing command: plugin=%s, command=%s, args=%v\n", plugin, command, cmdArgs)

	// 判断是否是交互式命令
	interactive := false
	if command == "interactive" {
		interactive = true
	}

	// 创建命令请求
	requestID := uuid.New().String()
	cmdMsg, err := protocol.NewCommandRequestMessage(
		requestID,
		plugin,
		command,
		cmdArgs,
		interactive,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to create command request: %w", err)
	}

	// 发送命令请求
	if err := protocol.WriteMessage(c.conn, cmdMsg); err != nil {
		return fmt.Errorf("failed to send command request: %w", err)
	}

	// 处理交互式命令
	if interactive {
		return c.handleInteractiveCommand(requestID)
	}

	// 读取命令响应
	for {
		fmt.Printf("Waiting for response...\n")
		respMsg, err := protocol.ReadMessage(c.conn)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		fmt.Printf("Received response: type=%d, requestID=%s\n", respMsg.Header.Type, respMsg.Header.RequestID)

		// 检查请求ID
		if respMsg.Header.RequestID != requestID {
			fmt.Printf("Ignoring response with different requestID: %s (expected %s)\n", respMsg.Header.RequestID, requestID)
			continue
		}

		// 处理响应
		switch respMsg.Header.Type {
		case protocol.CommandResponse:
			fmt.Printf("Processing command response...\n")
			var cmdResp protocol.CommandResponseBody
			if err := json.Unmarshal(respMsg.Body, &cmdResp); err != nil {
				return fmt.Errorf("failed to parse command response: %w", err)
			}

			fmt.Printf("Command response: success=%v, message=%s\n", cmdResp.Success, cmdResp.Message)

			if !cmdResp.Success {
				return fmt.Errorf("command failed: %s", cmdResp.Message)
			}

			if cmdResp.Data != nil {
				fmt.Println(string(cmdResp.Data))
			}

			return nil
		case protocol.DataStream:
			fmt.Printf("Received data stream (%d bytes)\n", len(respMsg.Body))
			fmt.Print(string(respMsg.Body))
		case protocol.ErrorResponse:
			fmt.Printf("Processing error response...\n")
			var errResp protocol.ErrorResponseBody
			if err := json.Unmarshal(respMsg.Body, &errResp); err != nil {
				return fmt.Errorf("failed to parse error response: %w", err)
			}
			return fmt.Errorf("error: %s", errResp.Message)
		default:
			fmt.Printf("Received unknown message type: %d\n", respMsg.Header.Type)
		}
	}
}

// handleInteractiveCommand 处理交互式命令
func (c *Client) handleInteractiveCommand(requestID string) error {
	// 创建通道
	dataCh := make(chan []byte, 10)
	errCh := make(chan error, 1)

	// 读取服务器响应
	go func() {
		for {
			respMsg, err := protocol.ReadMessage(c.conn)
			if err != nil {
				errCh <- fmt.Errorf("failed to read response: %w", err)
				return
			}

			// 检查请求ID
			if respMsg.Header.RequestID != requestID {
				continue
			}

			// 处理响应
			switch respMsg.Header.Type {
			case protocol.CommandResponse:
				var cmdResp protocol.CommandResponseBody
				if err := json.Unmarshal(respMsg.Body, &cmdResp); err != nil {
					errCh <- fmt.Errorf("failed to parse command response: %w", err)
					return
				}

				if !cmdResp.Success {
					errCh <- fmt.Errorf("command failed: %s", cmdResp.Message)
				} else {
					errCh <- nil
				}
				return
			case protocol.DataStream:
				dataCh <- respMsg.Body
			case protocol.ErrorResponse:
				var errResp protocol.ErrorResponseBody
				if err := json.Unmarshal(respMsg.Body, &errResp); err != nil {
					errCh <- fmt.Errorf("failed to parse error response: %w", err)
					return
				}
				errCh <- fmt.Errorf("error: %s", errResp.Message)
				return
			}
		}
	}()

	// 读取用户输入
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "exit" {
				break
			}

			// 发送数据
			dataMsg := protocol.NewDataStreamMessage(requestID, []byte(line+"\n"), false)
			if err := protocol.WriteMessage(c.conn, dataMsg); err != nil {
				errCh <- fmt.Errorf("failed to send data: %w", err)
				return
			}
		}
	}()

	// 处理数据
	for {
		select {
		case data := <-dataCh:
			fmt.Print(string(data))
		case err := <-errCh:
			return err
		}
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("Plugin Management:")
	fmt.Println("  manager list - List installed plugins")
	fmt.Println("  manager install <plugin_path> - Install a plugin")
	fmt.Println("  manager uninstall <plugin_id> - Uninstall a plugin")
	fmt.Println("  manager enable <plugin_id> - Enable a plugin")
	fmt.Println("  manager disable <plugin_id> - Disable a plugin")
	fmt.Println("  manager upgrade <plugin_id> <plugin_path> - Upgrade a plugin")
	fmt.Println("  manager info <plugin_id> - Show plugin information")
	fmt.Println("")
	fmt.Println("Service Management:")
	fmt.Println("  manager start <plugin_id> - Start a service plugin")
	fmt.Println("  manager stop <plugin_id> - Stop a service plugin")
	fmt.Println("  manager restart <plugin_id> - Restart a service plugin")
	fmt.Println("  manager status [plugin_id] - Show service plugin status")
	fmt.Println("  manager config <plugin_id> [config_file] - Configure a service plugin")
	fmt.Println("")
	fmt.Println("File Operations:")
	fmt.Println("  file upload <local_path> <remote_path> [--compress] [--overwrite] - Upload a file or directory")
	fmt.Println("  file upload <request_json> - Upload a file (legacy JSON format)")
	fmt.Println("  file download <remote_path> <local_path> [--compress] [--offset <offset>] [--recursive] - Download a file or directory")
	fmt.Println("  file download <request_json> - Download a file (legacy JSON format)")
	fmt.Println("  file list [path] - List files")
	fmt.Println("  file delete <path> - Delete a file or directory")
	fmt.Println("  file mkdir <path> - Create a directory")
	fmt.Println("")
	fmt.Println("Shell Operations:")
	fmt.Println("  shell exec <command> - Execute a shell command")
	fmt.Println("  shell interactive - Start an interactive shell")
	fmt.Println("")
	fmt.Println("Terminal Management:")
	fmt.Println("  terminal create <request_json> - Create a terminal")
	fmt.Println("  terminal list - List terminals")
	fmt.Println("  terminal kill <terminal_id> - Kill a terminal")
	fmt.Println("  terminal write <request_json> - Write to a terminal")
	fmt.Println("  terminal read <terminal_id> - Read from a terminal")
	// 代理服务命令已被移除，因为它的功能已经被 manager 插件的服务管理命令完全覆盖
	fmt.Println("")
	fmt.Println("Other Commands:")
	fmt.Println("  help - Show this help")
	fmt.Println("  exit/quit - Exit the client")
}
