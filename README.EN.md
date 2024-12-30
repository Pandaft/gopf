# GOPF - Go Port Forwarder

A lightweight port forwarding tool written in Go with a beautiful terminal user interface.

[简体中文](README.md) | **English**

## Features

- 🚀 Lightweight and high-performance port forwarding
- 🎨 Interactive terminal UI for intuitive rule creation and editing
- ⚙️ Simple YAML configuration with both UI and manual editing support
- 🔄 Dynamic rule management with real-time add, modify, and toggle
- 🔌 Support multiple forwarding rules simultaneously

## Interface Preview

![Terminal UI Preview](https://raw.githubusercontent.com/Pandaft/static-files/refs/heads/main/repo/gopf/images/en.webp)

## Installation

### Option 1: Direct Download (Recommended)

Download the latest version for your system from [Github Releases](https://github.com/pandaft/gopf/releases) page and extract it to run.

### Option 2: Via Go

If you have Go environment installed, you can install using:

```bash
go install github.com/pandaft/gopf@latest
```

## Quick Start

1. Run GOPF directly:

   ```bash
   gopf
   ```

2. Create forwarding rules using the interactive UI:
   - Press `a` to add a new rule
   - Fill in rule name, local port, remote host and port
   - Press `s` to start the rule

> Note: Configuration file will be created automatically on first run, no manual editing required.

If you prefer manual configuration, you can edit the `gopf.yaml` file:

```yaml
rules:
  - name: "SSH Forward"
    local_port: 2222
    remote_host: "remote.example.com"
    remote_port: 22
```

## Configuration

The configuration file uses YAML format and supports the following parameters:

```yaml
rules:
  - name: "Rule name"
    local_port: Local port number
    remote_host: "Remote host address"
    remote_port: Remote port number
```

## Usage Examples

```yaml
rules:
  # SSH remote connection forwarding
  - name: "SSH"
    local_port: 2222              # Local listening port
    remote_host: "192.168.1.100"  # Remote host address
    remote_port: 22               # SSH default port

  # Web service forwarding
  - name: "Web"
    local_port: 8080                # Local listening port
    remote_host: "web.example.com"  # Remote web server
    remote_port: 80                 # HTTP default port
```

## Keyboard Shortcuts

- `↑/↓`: Select rules
- `s`: Start/Stop rule
- `a`: Add rule
- `d`: Delete rule
- `q`: Quit

## License

MIT License