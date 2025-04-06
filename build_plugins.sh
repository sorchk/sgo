#!/bin/bash

# 创建插件目录
mkdir -p plugins

# 编译插件管理插件
echo "Building manager plugin..."
go build -buildmode=plugin -o plugins/manager.so ./plugins/manager

# 编译文件传输插件
echo "Building file plugin..."
go build -buildmode=plugin -o plugins/file.so ./plugins/file

# 编译Shell执行插件
echo "Building shell plugin..."
go build -buildmode=plugin -o plugins/shell.so ./plugins/shell

# 编译终端管理插件
echo "Building terminal plugin..."
go build -buildmode=plugin -o plugins/terminal.so ./plugins/terminal

# 编译代理服务插件
echo "Building proxy plugin..."
go build -buildmode=plugin -o plugins/proxy.so ./plugins/proxy

# 代理命令插件已被移除，因为它的功能已经被 manager 插件的服务管理命令完全覆盖

echo "All plugins built successfully!"
