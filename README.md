# TCP网络服务

这是一个基于Go语言的TCP网络服务，支持插件扩展，客户端认证和加密通信。

## 特性

- 仅开放一个端口提供服务
- 支持插件系统（服务类和命令类）
- 支持插件的安装、卸载、启用、禁用和升级
- 支持客户端认证和权限控制
- 使用XXTEA加密算法进行数据加密传输
- 提供多种实用插件：
  - 文件传输插件：支持文件上传下载，断点续传，压缩传输
  - Shell执行插件：支持执行服务器命令
  - 终端管理插件：支持管理远程终端
  - 代理服务插件：支持HTTP、SOCKS4和SOCKS5代理

## 安装

```bash
# 克隆仓库
git clone https://github.com/sorc/tcpserver.git
cd tcpserver

# 编译服务器
go build -o server ./cmd/server

# 编译客户端
go build -o client ./cmd/client
```

## 配置

### 服务器配置

服务器配置文件为`config.json`，示例：

```json
{
  "server": {
    "addr": ":8888",
    "plugins_dir": "plugins",
    "config_dir": "config"
  },
  "clients": [
    {
      "id": "client1",
      "secret": "secret1",
      "name": "Default Client",
      "permissions": [
        "plugin:manage",
        "service:manage",
        "plugin:use"
      ]
    }
  ]
}
```

### 客户端配置

客户端配置文件为`client.json`，示例：

```json
{
  "server_addr": "localhost:8888",
  "client_id": "client1",
  "secret": "secret1"
}
```

## 使用

### 启动服务器

```bash
./server -config config.json
```

### 启动命令行客户端

```bash
./client -config client.json
```

### 客户端命令

客户端支持以下命令：

- `manager list` - 列出已安装的插件
- `manager install <plugin_path>` - 安装插件
- `manager uninstall <plugin_id>` - 卸载插件
- `manager enable <plugin_id>` - 启用插件
- `manager disable <plugin_id>` - 禁用插件
- `manager info <plugin_id>` - 显示插件信息
- `file upload <request_json>` - 上传文件
- `file download <request_json>` - 下载文件
- `file list [path]` - 列出文件
- `file delete <path>` - 删除文件或目录
- `file mkdir <path>` - 创建目录
- `shell exec <command>` - 执行Shell命令
- `shell interactive` - 启动交互式Shell
- `terminal create <request_json>` - 创建终端
- `terminal list` - 列出终端
- `terminal kill <terminal_id>` - 终止终端
- `terminal write <request_json>` - 向终端写入数据
- `terminal read <terminal_id>` - 从终端读取数据
- `proxy status` - 显示代理状态
- `proxy start <proxy_type>` - 启动代理服务
- `proxy stop <proxy_type>` - 停止代理服务
- `help` - 显示帮助信息
- `exit/quit` - 退出客户端

## 插件开发

要开发新的插件，需要实现`plugin.Plugin`接口，并根据插件类型实现`plugin.ServicePlugin`或`plugin.CommandPlugin`接口。

### 服务类插件示例

```go
package main

import (
    "context"
    "io"
    
    "github.com/sorc/tcpserver/internal/plugin"
)

// MyServicePlugin 自定义服务插件
type MyServicePlugin struct {
    *plugin.BaseServicePlugin
    // 自定义字段
}

// Plugin 导出的插件实例
var Plugin *MyServicePlugin

func init() {
    Plugin = &MyServicePlugin{
        BaseServicePlugin: plugin.NewBaseServicePlugin("my-service", "My Service", "1.0.0"),
    }
}

// Start 启动服务
func (p *MyServicePlugin) Start(ctx context.Context) error {
    // 实现服务启动逻辑
    return p.BaseServicePlugin.Start(ctx)
}

// Stop 停止服务
func (p *MyServicePlugin) Stop() error {
    // 实现服务停止逻辑
    return p.BaseServicePlugin.Stop()
}

func main() {}
```

### 命令类插件示例

```go
package main

import (
    "context"
    "fmt"
    "io"
    
    "github.com/sorc/tcpserver/internal/plugin"
)

// MyCommandPlugin 自定义命令插件
type MyCommandPlugin struct {
    *plugin.BaseCommandPlugin
    // 自定义字段
}

// Plugin 导出的插件实例
var Plugin *MyCommandPlugin

func init() {
    Plugin = &MyCommandPlugin{
        BaseCommandPlugin: plugin.NewBaseCommandPlugin("my-command", "My Command", "1.0.0", plugin.OneTimeCommand),
    }
}

// GetCommands 获取支持的命令列表
func (p *MyCommandPlugin) GetCommands() []string {
    return []string{
        "hello",
        "world",
    }
}

// Execute 执行命令
func (p *MyCommandPlugin) Execute(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
    if len(args) == 0 {
        return fmt.Errorf("no command specified")
    }
    
    command := args[0]
    switch command {
    case "hello":
        fmt.Fprintln(output, "Hello!")
    case "world":
        fmt.Fprintln(output, "World!")
    default:
        return fmt.Errorf("unknown command: %s", command)
    }
    
    return nil
}

func main() {}
```

## 许可证

MIT
