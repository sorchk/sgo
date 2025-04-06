package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

// Start 启动HTTP代理
func (h *HTTPProxy) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.listener != nil {
		return nil // 已经在运行中
	}

	// 创建监听器
	listener, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}
	h.listener = listener

	// 创建HTTP服务器
	h.server = &http.Server{
		Handler: http.HandlerFunc(h.handleHTTP),
	}

	// 启动HTTP服务器
	go func() {
		h.server.Serve(listener)
	}()

	return nil
}

// Stop 停止HTTP代理
func (h *HTTPProxy) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.server != nil {
		// 设置关闭超时
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 关闭HTTP服务器
		h.server.Shutdown(ctx)
		h.server = nil
	}

	if h.listener != nil {
		h.listener.Close()
		h.listener = nil
	}
}

// IsRunning 检查HTTP代理是否运行中
func (h *HTTPProxy) IsRunning() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.listener != nil
}

// handleHTTP 处理HTTP代理请求
func (h *HTTPProxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		// 处理HTTPS请求
		h.handleHTTPS(w, r)
	} else {
		// 处理HTTP请求
		h.handlePlainHTTP(w, r)
	}
}

// handleHTTPS 处理HTTPS代理请求
func (h *HTTPProxy) handleHTTPS(w http.ResponseWriter, r *http.Request) {
	// 连接目标服务器
	dstConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer dstConn.Close()

	// 告诉客户端连接已建立
	w.WriteHeader(http.StatusOK)

	// 获取底层连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	// 劫持连接
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// 双向转发数据
	go func() {
		io.Copy(dstConn, clientConn)
	}()
	io.Copy(clientConn, dstConn)
}

// handlePlainHTTP 处理普通HTTP代理请求
func (h *HTTPProxy) handlePlainHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建新的请求
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	io.Copy(w, resp.Body)
}
