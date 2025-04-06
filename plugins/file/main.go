package main

import (
	"github.com/sorc/tcpserver/pkg/plugin"
)

// CreateCommandPlugin 创建命令类插件实例
func CreateCommandPlugin() plugin.ICommandPlugin {
	return &FileTransferPlugin{
		BaseCommandPlugin: plugin.NewBaseCommandPlugin("file", "File Transfer", "1.0.0", plugin.OneTimeCommand),
	}
}

// CreatePlugin 创建插件实例
func CreatePlugin() plugin.Plugin {
	return CreateCommandPlugin()
}

func main() {}
