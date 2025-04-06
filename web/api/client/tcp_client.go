package client

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/xxtea/xxtea-go/xxtea"
)

// TCPClient TCP客户端
type TCPClient struct {
	Addr      string
	ClientID  string
	Secret    string
	conn      net.Conn
	connected bool
	mutex     sync.Mutex
	timeout   time.Duration
}

// MessageType 消息类型
type MessageType byte

const (
	// AuthRequest 认证请求
	AuthRequest MessageType = 1
	// AuthResponse 认证响应
	AuthResponse MessageType = 2
	// CommandRequest 命令请求
	CommandRequest MessageType = 3
	// CommandResponse 命令响应
	CommandResponse MessageType = 4
	// DataRequest 数据请求
	DataRequest MessageType = 5
	// ErrorResponse 错误响应
	ErrorResponse MessageType = 6
)

// NewTCPClient 创建TCP客户端
func NewTCPClient(addr, clientID, secret string) *TCPClient {
	return &TCPClient{
		Addr:     addr,
		ClientID: clientID,
		Secret:   secret,
		timeout:  30 * time.Second,
	}
}

// Connect 连接服务器
func (c *TCPClient) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return nil
	}

	// 连接服务器
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn

	// 发送认证请求
	requestID := generateRequestID()
	authReq := map[string]string{
		"client_id": c.ClientID,
		"secret":    c.Secret,
	}
	authReqJSON, err := encodeJSON(authReq)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to encode auth request: %w", err)
	}

	// 发送认证请求
	err = c.sendMessage(AuthRequest, requestID, authReqJSON)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	// 接收认证响应
	msgType, respRequestID, payload, err := c.receiveMessage()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to receive auth response: %w", err)
	}

	// 检查响应类型
	if msgType == ErrorResponse {
		c.conn.Close()
		return fmt.Errorf("authentication failed: %s", string(payload))
	}

	if msgType != AuthResponse {
		c.conn.Close()
		return fmt.Errorf("unexpected response type: %d", msgType)
	}

	// 检查请求ID
	if respRequestID != requestID {
		c.conn.Close()
		return fmt.Errorf("request ID mismatch")
	}

	c.connected = true
	return nil
}

// Disconnect 断开连接
func (c *TCPClient) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	err := c.conn.Close()
	c.connected = false
	return err
}

// IsConnected 检查是否已连接
func (c *TCPClient) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.connected
}

// ExecuteCommand 执行命令
func (c *TCPClient) ExecuteCommand(plugin, command string, args []string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return "", errors.New("not connected to server")
	}

	// 生成请求ID
	requestID := generateRequestID()

	// 构建命令请求
	cmdReq := map[string]interface{}{
		"plugin":  plugin,
		"command": command,
		"args":    args,
	}
	cmdReqJSON, err := encodeJSON(cmdReq)
	if err != nil {
		return "", fmt.Errorf("failed to encode command request: %w", err)
	}

	// 发送命令请求
	err = c.sendMessage(CommandRequest, requestID, cmdReqJSON)
	if err != nil {
		return "", fmt.Errorf("failed to send command request: %w", err)
	}

	// 接收命令响应
	var output bytes.Buffer
	for {
		msgType, respRequestID, payload, err := c.receiveMessage()
		if err != nil {
			return output.String(), fmt.Errorf("failed to receive command response: %w", err)
		}

		// 检查请求ID
		if respRequestID != requestID {
			continue
		}

		// 处理响应
		switch msgType {
		case CommandResponse:
			// 解析命令响应
			resp, err := decodeJSON(payload)
			if err != nil {
				return output.String(), fmt.Errorf("failed to decode command response: %w", err)
			}

			// 检查命令是否成功
			success, ok := resp["success"].(bool)
			if !ok || !success {
				errMsg, _ := resp["message"].(string)
				return output.String(), fmt.Errorf("command execution failed: %s", errMsg)
			}

			return output.String(), nil

		case DataRequest:
			// 将数据追加到输出
			output.Write(payload)

		case ErrorResponse:
			return output.String(), fmt.Errorf("command execution failed: %s", string(payload))

		default:
			return output.String(), fmt.Errorf("unexpected response type: %d", msgType)
		}
	}
}

// sendMessage 发送消息
func (c *TCPClient) sendMessage(msgType MessageType, requestID string, payload []byte) error {
	// 设置写入超时
	c.conn.SetWriteDeadline(time.Now().Add(c.timeout))

	// 加密负载
	encryptedPayload := xxtea.Encrypt(payload, []byte(c.Secret))

	// 构建消息头
	header := make([]byte, 21)
	header[0] = byte(msgType)
	copy(header[1:17], []byte(requestID))
	binary.BigEndian.PutUint32(header[17:21], uint32(len(encryptedPayload)))

	// 发送消息头
	_, err := c.conn.Write(header)
	if err != nil {
		return err
	}

	// 发送加密负载
	_, err = c.conn.Write(encryptedPayload)
	return err
}

// receiveMessage 接收消息
func (c *TCPClient) receiveMessage() (MessageType, string, []byte, error) {
	// 设置读取超时
	c.conn.SetReadDeadline(time.Now().Add(c.timeout))

	// 读取消息头
	header := make([]byte, 21)
	_, err := io.ReadFull(c.conn, header)
	if err != nil {
		return 0, "", nil, err
	}

	// 解析消息头
	msgType := MessageType(header[0])
	requestID := string(header[1:17])
	payloadLen := binary.BigEndian.Uint32(header[17:21])

	// 读取加密负载
	encryptedPayload := make([]byte, payloadLen)
	_, err = io.ReadFull(c.conn, encryptedPayload)
	if err != nil {
		return 0, "", nil, err
	}

	// 解密负载
	payload := xxtea.Decrypt(encryptedPayload, []byte(c.Secret))
	if payload == nil {
		return 0, "", nil, errors.New("failed to decrypt payload")
	}

	return msgType, requestID, payload, nil
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// encodeJSON 编码JSON
func encodeJSON(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decodeJSON 解码JSON
func decodeJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
