package protocol

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

// MessageType 定义消息类型
type MessageType uint8

const (
	// AuthRequest 认证请求
	AuthRequest MessageType = iota + 1
	// AuthResponse 认证响应
	AuthResponse
	// CommandRequest 命令请求
	CommandRequest
	// CommandResponse 命令响应
	CommandResponse
	// DataStream 数据流
	DataStream
	// ErrorResponse 错误响应
	ErrorResponse
	// HeartbeatRequest 心跳请求
	HeartbeatRequest
	// HeartbeatResponse 心跳响应
	HeartbeatResponse
)

// Header 消息头
type Header struct {
	Type      MessageType `json:"type"`
	Length    uint32      `json:"length"`
	RequestID string      `json:"request_id"`
	Encrypted bool        `json:"encrypted"`
}

// Message 消息结构
type Message struct {
	Header Header `json:"header"`
	Body   []byte `json:"body,omitempty"`
}

// AuthRequestBody 认证请求体
type AuthRequestBody struct {
	ClientID  string `json:"client_id"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"`
}

// AuthResponseBody 认证响应体
type AuthResponseBody struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// CommandRequestBody 命令请求体
type CommandRequestBody struct {
	Plugin      string   `json:"plugin"`
	Command     string   `json:"command"`
	Args        []string `json:"args,omitempty"`
	Interactive bool     `json:"interactive"`
}

// CommandResponseBody 命令响应体
type CommandResponseBody struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    []byte `json:"data,omitempty"`
}

// ErrorResponseBody 错误响应体
type ErrorResponseBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HeartbeatRequestBody 心跳请求体
type HeartbeatRequestBody struct {
	Timestamp int64 `json:"timestamp"`
}

// HeartbeatResponseBody 心跳响应体
type HeartbeatResponseBody struct {
	Timestamp  int64   `json:"timestamp"`
	ServerLoad float64 `json:"server_load"`
}

// ReadMessage 从连接中读取消息
func ReadMessage(r io.Reader) (*Message, error) {
	// 读取消息头长度
	var headerLen uint16
	if err := binary.Read(r, binary.BigEndian, &headerLen); err != nil {
		return nil, err
	}

	// 读取消息头
	headerBytes := make([]byte, headerLen)
	if _, err := io.ReadFull(r, headerBytes); err != nil {
		return nil, err
	}

	var header Header
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}

	// 读取消息体
	body := make([]byte, header.Length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}

	return &Message{
		Header: header,
		Body:   body,
	}, nil
}

// WriteMessage 将消息写入连接
func WriteMessage(w io.Writer, msg *Message) error {
	// 序列化消息头
	headerBytes, err := json.Marshal(msg.Header)
	if err != nil {
		return err
	}

	// 写入消息头长度
	headerLen := uint16(len(headerBytes))
	if err := binary.Write(w, binary.BigEndian, headerLen); err != nil {
		return err
	}

	// 写入消息头
	if _, err := w.Write(headerBytes); err != nil {
		return err
	}

	// 写入消息体
	if _, err := w.Write(msg.Body); err != nil {
		return err
	}

	return nil
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, requestID string, body []byte, encrypted bool) *Message {
	return &Message{
		Header: Header{
			Type:      msgType,
			Length:    uint32(len(body)),
			RequestID: requestID,
			Encrypted: encrypted,
		},
		Body: body,
	}
}

// NewAuthRequestMessage 创建认证请求消息
func NewAuthRequestMessage(requestID string, clientID, nonce string, timestamp int64, signature string, encrypted bool) (*Message, error) {
	body := AuthRequestBody{
		ClientID:  clientID,
		Nonce:     nonce,
		Timestamp: timestamp,
		Signature: signature,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return NewMessage(AuthRequest, requestID, bodyBytes, encrypted), nil
}

// NewAuthResponseMessage 创建认证响应消息
func NewAuthResponseMessage(requestID string, success bool, sessionID, message string, encrypted bool) (*Message, error) {
	body := AuthResponseBody{
		Success:   success,
		SessionID: sessionID,
		Message:   message,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return NewMessage(AuthResponse, requestID, bodyBytes, encrypted), nil
}

// NewCommandRequestMessage 创建命令请求消息
func NewCommandRequestMessage(requestID string, plugin, command string, args []string, interactive bool, encrypted bool) (*Message, error) {
	body := CommandRequestBody{
		Plugin:      plugin,
		Command:     command,
		Args:        args,
		Interactive: interactive,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return NewMessage(CommandRequest, requestID, bodyBytes, encrypted), nil
}

// NewCommandResponseMessage 创建命令响应消息
func NewCommandResponseMessage(requestID string, success bool, message string, data []byte, encrypted bool) (*Message, error) {
	body := CommandResponseBody{
		Success: success,
		Message: message,
		Data:    data,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return NewMessage(CommandResponse, requestID, bodyBytes, encrypted), nil
}

// NewErrorResponseMessage 创建错误响应消息
func NewErrorResponseMessage(requestID string, code int, message string, encrypted bool) (*Message, error) {
	body := ErrorResponseBody{
		Code:    code,
		Message: message,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return NewMessage(ErrorResponse, requestID, bodyBytes, encrypted), nil
}

// NewDataStreamMessage 创建数据流消息
func NewDataStreamMessage(requestID string, data []byte, encrypted bool) *Message {
	return NewMessage(DataStream, requestID, data, encrypted)
}

// NewHeartbeatRequestMessage 创建心跳请求消息
func NewHeartbeatRequestMessage(requestID string, timestamp int64, encrypted bool) (*Message, error) {
	body := HeartbeatRequestBody{
		Timestamp: timestamp,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return NewMessage(HeartbeatRequest, requestID, bodyBytes, encrypted), nil
}

// NewHeartbeatResponseMessage 创建心跳响应消息
func NewHeartbeatResponseMessage(requestID string, timestamp int64, serverLoad float64, encrypted bool) (*Message, error) {
	body := HeartbeatResponseBody{
		Timestamp:  timestamp,
		ServerLoad: serverLoad,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return NewMessage(HeartbeatResponse, requestID, bodyBytes, encrypted), nil
}
