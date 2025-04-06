package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sorc/tcpserver/internal/auth"
	"github.com/sorc/tcpserver/internal/crypto"
	"github.com/sorc/tcpserver/pkg/plugin"
	"github.com/sorc/tcpserver/pkg/protocol"
)

// Server TCP服务器
type Server struct {
	listener      net.Listener
	addr          string
	authManager   *auth.AuthManager
	pluginManager plugin.PluginManager
	clients       map[string]*Client
	clientsMu     sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	pluginsDir    string
	configDir     string
}

// Client 客户端连接
type Client struct {
	conn       net.Conn
	sessionID  string
	clientInfo *auth.Client
	cipher     *crypto.XXTEACipher
	ctx        context.Context
	cancel     context.CancelFunc
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Addr       string `json:"addr"`
	PluginsDir string `json:"plugins_dir"`
	ConfigDir  string `json:"config_dir"`
}

// NewServer 创建新的服务器
func NewServer(config ServerConfig, pluginManager plugin.PluginManager) (*Server, error) {
	// 创建目录
	if err := os.MkdirAll(config.PluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugins directory: %w", err)
	}
	if err := os.MkdirAll(config.ConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		addr:          config.Addr,
		authManager:   auth.NewAuthManager(),
		pluginManager: pluginManager,
		clients:       make(map[string]*Client),
		ctx:           ctx,
		cancel:        cancel,
		pluginsDir:    config.PluginsDir,
		configDir:     config.ConfigDir,
	}, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	// 内置插件已经在main.go中加载

	// 启动TCP监听
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	s.listener = listener

	log.Printf("Server started on %s", s.addr)

	// 接受连接
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptConnections()
	}()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	// 取消上下文
	s.cancel()

	// 关闭监听器
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	// 关闭所有客户端连接
	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.cancel()
		client.conn.Close()
	}
	s.clientsMu.Unlock()

	// 等待所有goroutine结束
	s.wg.Wait()

	log.Println("Server stopped")
	return nil
}

// acceptConnections 接受客户端连接
func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				// 服务器正在关闭
				return
			default:
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
		}

		// 处理新连接
		clientCtx, clientCancel := context.WithCancel(s.ctx)
		client := &Client{
			conn:   conn,
			ctx:    clientCtx,
			cancel: clientCancel,
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer clientCancel()
			defer conn.Close()

			s.handleClient(client)
		}()
	}
}

// handleClient 处理客户端连接
func (s *Server) handleClient(client *Client) {
	log.Printf("New connection from %s", client.conn.RemoteAddr())

	// 等待认证
	if err := s.authenticateClient(client); err != nil {
		log.Printf("Client authentication failed: %v", err)
		return
	}

	// 添加到客户端列表
	s.clientsMu.Lock()
	s.clients[client.sessionID] = client
	s.clientsMu.Unlock()

	defer func() {
		// 从客户端列表中移除
		s.clientsMu.Lock()
		delete(s.clients, client.sessionID)
		s.clientsMu.Unlock()
		log.Printf("Client %s disconnected", client.clientInfo.ID)
	}()

	log.Printf("Client %s authenticated successfully", client.clientInfo.ID)

	// 处理客户端消息
	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			// 读取消息
			msg, err := protocol.ReadMessage(client.conn)
			if err != nil {
				if err == io.EOF {
					log.Printf("Client %s closed connection", client.clientInfo.ID)
				} else {
					log.Printf("Error reading message from client %s: %v", client.clientInfo.ID, err)
				}
				return
			}

			// 处理消息
			if err := s.handleMessage(client, msg); err != nil {
				log.Printf("Error handling message from client %s: %v", client.clientInfo.ID, err)
				// 发送错误响应
				errMsg, _ := protocol.NewErrorResponseMessage(msg.Header.RequestID, 500, err.Error(), false)
				protocol.WriteMessage(client.conn, errMsg)
			}
		}
	}
}

// authenticateClient 认证客户端
func (s *Server) authenticateClient(client *Client) error {
	// 设置认证超时
	client.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	defer client.conn.SetReadDeadline(time.Time{})

	// 读取认证消息
	msg, err := protocol.ReadMessage(client.conn)
	if err != nil {
		return fmt.Errorf("failed to read auth message: %w", err)
	}

	// 验证消息类型
	if msg.Header.Type != protocol.AuthRequest {
		return errors.New("expected auth request message")
	}

	// 解析认证请求
	var authReq protocol.AuthRequestBody
	if err := json.Unmarshal(msg.Body, &authReq); err != nil {
		return fmt.Errorf("failed to parse auth request: %w", err)
	}

	// 认证客户端
	sessionID, err := s.authManager.Authenticate(authReq.ClientID, authReq.Nonce, authReq.Timestamp, authReq.Signature)
	if err != nil {
		// 发送认证失败响应
		respMsg, _ := protocol.NewAuthResponseMessage(msg.Header.RequestID, false, "", err.Error(), false)
		protocol.WriteMessage(client.conn, respMsg)
		return fmt.Errorf("authentication failed: %w", err)
	}

	// 获取客户端信息
	clientInfo, err := s.authManager.GetClient(authReq.ClientID)
	if err != nil {
		return fmt.Errorf("failed to get client info: %w", err)
	}

	// 创建加密器
	cipher, err := crypto.NewXXTEACipher([]byte(clientInfo.Secret))
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// 更新客户端信息
	client.sessionID = sessionID
	client.clientInfo = clientInfo
	client.cipher = cipher

	// 发送认证成功响应
	respMsg, err := protocol.NewAuthResponseMessage(msg.Header.RequestID, true, sessionID, "Authentication successful", false)
	if err != nil {
		return fmt.Errorf("failed to create auth response: %w", err)
	}

	if err := protocol.WriteMessage(client.conn, respMsg); err != nil {
		return fmt.Errorf("failed to send auth response: %w", err)
	}

	return nil
}

// handleMessage 处理客户端消息
func (s *Server) handleMessage(client *Client, msg *protocol.Message) error {
	// 解密消息体（如果需要）
	body := msg.Body
	if msg.Header.Encrypted {
		decrypted, err := client.cipher.Decrypt(body)
		if err != nil {
			return fmt.Errorf("failed to decrypt message: %w", err)
		}
		body = decrypted
	}

	// 根据消息类型处理
	switch msg.Header.Type {
	case protocol.CommandRequest:
		return s.handleCommandRequest(client, msg.Header.RequestID, body, msg.Header.Encrypted)
	case protocol.HeartbeatRequest:
		return s.handleHeartbeatRequest(client, msg.Header.RequestID, body, msg.Header.Encrypted)
	case protocol.DataStream:
		return s.handleDataStream(client, msg.Header.RequestID, body)
	default:
		return fmt.Errorf("unsupported message type: %d", msg.Header.Type)
	}
}

// handleCommandRequest 处理命令请求
func (s *Server) handleCommandRequest(client *Client, requestID string, body []byte, encrypted bool) error {
	var cmdReq protocol.CommandRequestBody
	if err := json.Unmarshal(body, &cmdReq); err != nil {
		return fmt.Errorf("failed to parse command request: %w", err)
	}

	log.Printf("Received command request: plugin=%s, command=%s, args=%v", cmdReq.Plugin, cmdReq.Command, cmdReq.Args)

	// 检查权限
	hasPermission, err := s.authManager.HasPluginPermission(client.clientInfo.ID, cmdReq.Plugin)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("no permission to use plugin: %s", cmdReq.Plugin)
	}

	// 获取插件
	p, err := s.pluginManager.GetPlugin(cmdReq.Plugin)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// 检查插件状态
	if p.State() != plugin.Enabled && p.State() != plugin.Running {
		return fmt.Errorf("plugin %s is not enabled", cmdReq.Plugin)
	}

	// 创建管道用于命令输入输出
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	// 创建响应通道
	respCh := make(chan error, 1)

	// 检查插件类型
	if p.Type() != plugin.CommandPlugin {
		// 不支持的插件类型
		return fmt.Errorf("plugin %s is not a command plugin", cmdReq.Plugin)
	}

	// 获取命令插件
	cmdPlugin, err := s.pluginManager.GetCommandPlugin(cmdReq.Plugin)
	if err != nil {
		return fmt.Errorf("failed to get command plugin: %w", err)
	}

	// 执行命令
	go func() {
		// 创建上下文，并将插件管理器传递给插件
		ctx := context.WithValue(client.ctx, "plugin_manager", s.pluginManager)

		// 执行命令
		err := cmdPlugin.Execute(ctx, append([]string{cmdReq.Command}, cmdReq.Args...), nil, pw)

		// 关闭写入端，表示命令执行完成
		pw.Close()

		// 发送命令执行结果
		respCh <- err
	}()

	// 读取命令输出并发送给客户端
	log.Printf("Reading command output...")
	buf := make([]byte, 4096)
	for {
		n, err := pr.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("Command output completed (EOF)")
				break
			}
			return fmt.Errorf("failed to read command output: %w", err)
		}

		log.Printf("Read %d bytes from command output", n)

		// 发送数据流消息
		dataMsg := protocol.NewDataStreamMessage(requestID, buf[:n], encrypted)
		if err := protocol.WriteMessage(client.conn, dataMsg); err != nil {
			return fmt.Errorf("failed to send data stream: %w", err)
		}
		log.Printf("Data stream sent to client")
	}

	// 等待命令执行完成
	log.Printf("Waiting for command execution to complete...")
	cmdErr := <-respCh
	log.Printf("Command execution completed with error: %v", cmdErr)

	// 发送命令响应
	var respMsg *protocol.Message
	if cmdErr != nil {
		log.Printf("Command execution failed: %v", cmdErr)
		var err error
		respMsg, err = protocol.NewCommandResponseMessage(requestID, false, cmdErr.Error(), nil, encrypted)
		if err != nil {
			log.Printf("Failed to create command response message: %v", err)
			return fmt.Errorf("failed to create command response message: %w", err)
		}
	} else {
		log.Printf("Command executed successfully")
		var err error
		respMsg, err = protocol.NewCommandResponseMessage(requestID, true, "Command executed successfully", nil, encrypted)
		if err != nil {
			log.Printf("Failed to create command response message: %v", err)
			return fmt.Errorf("failed to create command response message: %w", err)
		}
	}

	log.Printf("Sending command response: requestID=%s, success=%v", requestID, cmdErr == nil)
	if err := protocol.WriteMessage(client.conn, respMsg); err != nil {
		log.Printf("Failed to send command response: %v", err)
		return fmt.Errorf("failed to send command response: %w", err)
	}

	log.Printf("Command response sent successfully")

	return nil
}

// handleHeartbeatRequest 处理心跳请求
func (s *Server) handleHeartbeatRequest(client *Client, requestID string, body []byte, encrypted bool) error {
	var heartbeatReq protocol.HeartbeatRequestBody
	if err := json.Unmarshal(body, &heartbeatReq); err != nil {
		return fmt.Errorf("failed to parse heartbeat request: %w", err)
	}

	// 创建心跳响应
	respMsg, err := protocol.NewHeartbeatResponseMessage(requestID, time.Now().Unix(), 0.0, encrypted)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat response: %w", err)
	}

	// 发送心跳响应
	if err := protocol.WriteMessage(client.conn, respMsg); err != nil {
		return fmt.Errorf("failed to send heartbeat response: %w", err)
	}

	return nil
}

// handleDataStream 处理数据流
func (s *Server) handleDataStream(client *Client, requestID string, body []byte) error {
	// 数据流通常是作为命令执行的一部分处理的
	// 这里可以添加额外的处理逻辑
	return nil
}

// loadBuiltinPlugins 加载内置插件
func (s *Server) loadBuiltinPlugins() error {
	// 加载插件管理插件
	managerPluginPath := filepath.Join(s.pluginsDir, "manager.so")
	if _, err := os.Stat(managerPluginPath); os.IsNotExist(err) {
		log.Printf("Manager plugin not found at %s, skipping", managerPluginPath)
	} else if err == nil {
		_, err := s.LoadPlugin(managerPluginPath)
		if err != nil {
			log.Printf("Failed to load manager plugin: %v", err)
		} else {
			log.Printf("Manager plugin loaded successfully")
			s.EnablePlugin("manager")
		}
	}

	// 加载文件传输插件
	filePluginPath := filepath.Join(s.pluginsDir, "file.so")
	if _, err := os.Stat(filePluginPath); os.IsNotExist(err) {
		log.Printf("File plugin not found at %s, skipping", filePluginPath)
	} else if err == nil {
		_, err := s.LoadPlugin(filePluginPath)
		if err != nil {
			log.Printf("Failed to load file plugin: %v", err)
		} else {
			log.Printf("File plugin loaded successfully")
			s.EnablePlugin("file")
		}
	}

	// 加载Shell插件
	shellPluginPath := filepath.Join(s.pluginsDir, "shell.so")
	if _, err := os.Stat(shellPluginPath); os.IsNotExist(err) {
		log.Printf("Shell plugin not found at %s, skipping", shellPluginPath)
	} else if err == nil {
		_, err := s.LoadPlugin(shellPluginPath)
		if err != nil {
			log.Printf("Failed to load shell plugin: %v", err)
		} else {
			log.Printf("Shell plugin loaded successfully")
			s.EnablePlugin("shell")
		}
	}

	// 加载终端插件
	terminalPluginPath := filepath.Join(s.pluginsDir, "terminal.so")
	if _, err := os.Stat(terminalPluginPath); os.IsNotExist(err) {
		log.Printf("Terminal plugin not found at %s, skipping", terminalPluginPath)
	} else if err == nil {
		_, err := s.LoadPlugin(terminalPluginPath)
		if err != nil {
			log.Printf("Failed to load terminal plugin: %v", err)
		} else {
			log.Printf("Terminal plugin loaded successfully")
			s.EnablePlugin("terminal")
		}
	}

	// 加载代理插件
	proxyPluginPath := filepath.Join(s.pluginsDir, "proxy.so")
	if _, err := os.Stat(proxyPluginPath); os.IsNotExist(err) {
		log.Printf("Proxy plugin not found at %s, skipping", proxyPluginPath)
	} else if err == nil {
		_, err := s.LoadPlugin(proxyPluginPath)
		if err != nil {
			log.Printf("Failed to load proxy plugin: %v", err)
		} else {
			log.Printf("Proxy plugin loaded successfully")
			s.EnablePlugin("proxy")
		}
	}

	return nil
}

// RegisterClient 注册客户端
func (s *Server) RegisterClient(client *auth.Client) error {
	return s.authManager.AddClient(client)
}

// UnregisterClient 注销客户端
func (s *Server) UnregisterClient(clientID string) error {
	return s.authManager.RemoveClient(clientID)
}

// LoadPlugin 加载插件
func (s *Server) LoadPlugin(path string) (plugin.Plugin, error) {
	return s.pluginManager.LoadPlugin(path)
}

// UnloadPlugin 卸载插件
func (s *Server) UnloadPlugin(id string) error {
	return s.pluginManager.UnloadPlugin(id)
}

// EnablePlugin 启用插件
func (s *Server) EnablePlugin(id string) error {
	return s.pluginManager.EnablePlugin(id)
}

// DisablePlugin 禁用插件
func (s *Server) DisablePlugin(id string) error {
	return s.pluginManager.DisablePlugin(id)
}

// GetPlugin 获取插件
func (s *Server) GetPlugin(id string) (plugin.Plugin, error) {
	return s.pluginManager.GetPlugin(id)
}

// ListPlugins 列出所有插件
func (s *Server) ListPlugins() []plugin.Plugin {
	return s.pluginManager.ListPlugins()
}
