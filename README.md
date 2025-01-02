# GOPF - Go 端口转发工具

一个使用 Go 语言编写的轻量级端口转发工具，具有美观的终端用户界面。

**简体中文** | [English](README.EN.md)

## 功能特点

- 🚀 轻量级且高性能的端口转发
- 🎨 交互式终端用户界面，直观完成规则的创建和编辑
- ⚙️ 简单的 YAML 配置，支持手动编辑或界面操作
- 🔄 支持动态规则管理，随时添加、修改、启用和禁用规则
- 🔌 支持多条转发规则同时运行

## 界面预览

![截图](https://raw.githubusercontent.com/Pandaft/static-files/refs/heads/main/repo/gopf/images/zh.webp)

## 安装

### 方式一：直接下载（推荐）

从 [Github Releases](https://github.com/pandaft/gopf/releases) 页面下载适合你系统的最新版本，解压后即可运行。

### 方式二：通过 Go 安装

如果你已经安装了 Go 环境，可以通过以下命令安装：

```bash
go install github.com/pandaft/gopf@latest
```

## 快速开始

1. 直接运行 GOPF：

    ```bash
    gopf
    ```

2. 使用交互式界面创建转发规则：
   - 按 `a` 键添加新规则
   - 填写规则名称、本地端口、远程主机和端口
   - 按 `s` 键启动规则

> 提示：首次运行时会自动创建配置文件，无需手动编辑。

如果你更喜欢手动编辑配置，可以修改 `gopf.yaml` 文件：

```yaml
rules:
  - name: "SSH转发"
    local_port: 2222
    remote_host: "remote.example.com"
    remote_port: 22
```

## 配置说明

配置文件使用 YAML 格式，支持以下参数：

```yaml
rules:
  - name: "规则名称"
    local_port: 本地端口号
    remote_host: "远程主机地址"
    remote_port: 远程端口号
```

## 使用示例

```yaml
rules:
  # SSH 远程连接转发
  - name: "SSH"
    local_port: 2222              # 本地监听端口
    remote_host: "192.168.1.100"  # 远程主机地址
    remote_port: 22               # SSH 默认端口

  # Web 服务转发
  - name: "Web"
    local_port: 8080                # 本地监听端口
    remote_host: "web.example.com"  # 远程 Web 服务器
    remote_port: 80                 # HTTP 默认端口
```

## 键盘热键

- `↑/↓`: 选择规则
- `←/→`: 选择选项
- `s`: 启动/停止规则
- `a`: 添加规则
- `d`: 删除规则
- `c`: 清空统计数据
- `q`: 退出程序

## 许可证

MIT License