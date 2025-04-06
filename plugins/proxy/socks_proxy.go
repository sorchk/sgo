package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// Start 启动SOCKS代理
func (s *SocksProxy) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return nil // 已经在运行中
	}

	// 创建监听器
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener

	// 创建上下文
	s.ctx, s.cancel = context.WithCancel(ctx)

	// 启动代理服务
	go s.serve()

	return nil
}

// Stop 停止SOCKS代理
func (s *SocksProxy) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}

// IsRunning 检查SOCKS代理是否运行中
func (s *SocksProxy) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener != nil
}

// serve 运行SOCKS代理服务
func (s *SocksProxy) serve() {
	for {
		// 接受连接
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				// 服务已停止
				return
			default:
				// 继续接受连接
				continue
			}
		}

		// 处理连接
		go s.handleConnection(conn)
	}
}

// handleConnection 处理SOCKS连接
func (s *SocksProxy) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取第一个字节来确定SOCKS版本
	versionBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, versionBuf); err != nil {
		return
	}

	// 根据版本选择处理方法
	switch versionBuf[0] {
	case 4:
		// SOCKS4协议
		s.handleSocks4(conn, versionBuf[0])
	case 5:
		// SOCKS5协议
		s.handleSocks5(conn)
	default:
		// 不支持的版本
		conn.Close()
	}
}

// handleSocks4 处理SOCKS4连接
func (s *SocksProxy) handleSocks4(conn net.Conn, firstByte byte) {
	// 读取剩余的SOCKS4请求
	buf := make([]byte, 8)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}

	// 获取命令
	cmd := buf[0]
	if cmd != 1 { // 只支持CONNECT命令
		conn.Write([]byte{0, 91, 0, 0, 0, 0, 0, 0})
		return
	}

	// 获取端口
	port := int(buf[1])<<8 | int(buf[2])

	// 获取IP地址
	ip := net.IPv4(buf[3], buf[4], buf[5], buf[6])

	// 读取用户ID
	var userId []byte
	for {
		b := make([]byte, 1)
		if _, err := conn.Read(b); err != nil {
			return
		}
		if b[0] == 0 {
			break
		}
		userId = append(userId, b[0])
	}

	// 连接目标服务器
	targetConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		conn.Write([]byte{0, 91, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	// 发送成功响应
	conn.Write([]byte{0, 90, 0, 0, 0, 0, 0, 0})

	// 双向转发数据
	go func() {
		io.Copy(targetConn, conn)
	}()
	io.Copy(conn, targetConn)
}

// handleSocks5 处理SOCKS5连接
func (s *SocksProxy) handleSocks5(conn net.Conn) {
	// 读取认证方法数量
	methodsBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, methodsBuf); err != nil {
		return
	}

	// 读取认证方法列表
	nmethods := int(methodsBuf[0])
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return
	}

	// 选择认证方法（这里选择无认证）
	conn.Write([]byte{5, 0})

	// 读取请求
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}

	// 检查版本和命令
	if buf[0] != 5 {
		return
	}

	cmd := buf[1]
	if cmd != 1 { // 只支持CONNECT命令
		conn.Write([]byte{5, 7, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}

	// 读取地址类型
	atyp := buf[3]
	var host string
	var port int

	switch atyp {
	case 1: // IPv4
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return
		}
		host = net.IPv4(addr[0], addr[1], addr[2], addr[3]).String()
	case 3: // 域名
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return
		}
		length := int(lenBuf[0])
		domainBuf := make([]byte, length)
		if _, err := io.ReadFull(conn, domainBuf); err != nil {
			return
		}
		host = string(domainBuf)
	case 4: // IPv6
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return
		}
		host = net.IP(addr).String()
	default:
		conn.Write([]byte{5, 8, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}

	// 读取端口
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return
	}
	port = int(portBuf[0])<<8 | int(portBuf[1])

	// 连接目标服务器
	targetConn, err := net.Dial("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		conn.Write([]byte{5, 4, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	// 获取本地地址
	localAddr := targetConn.LocalAddr().(*net.TCPAddr)
	localIP := localAddr.IP.To4()
	if localIP == nil {
		// 如果不是IPv4地址，使用回环地址
		localIP = net.IPv4(127, 0, 0, 1).To4()
	}
	localPort := localAddr.Port

	// 发送成功响应
	resp := []byte{5, 0, 0, 1, localIP[0], localIP[1], localIP[2], localIP[3], byte(localPort >> 8), byte(localPort & 0xff)}
	conn.Write(resp)

	// 设置超时
	deadline := time.Now().Add(5 * time.Minute)
	conn.SetDeadline(deadline)
	targetConn.SetDeadline(deadline)

	// 双向转发数据
	go func() {
		io.Copy(targetConn, conn)
	}()
	io.Copy(conn, targetConn)
}
