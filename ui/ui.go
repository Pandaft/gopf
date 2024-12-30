package ui

import (
	"fmt"
	"gopf/config"
	"gopf/forwarder"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 3).
			Align(lipgloss.Center).
			Width(50)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	// 选中状态的按钮样式 - 使用蓝色
	selectedButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")). // 亮白色文字
				Background(lipgloss.Color("27")). // 蓝色背景
				Bold(true).
				Padding(0, 3)

	// 未选中状态的按钮样式 - 使用暗灰色
	unselectedButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("251")). // 浅灰色文字
				Background(lipgloss.Color("237")). // 深灰色背景
				Padding(0, 3)
)

type inputField struct {
	textinput textinput.Model
	label     string
}

type mode int

const (
	normalMode mode = iota
	addMode
	editMode
	confirmMode
)

type model struct {
	table      table.Model
	config     *config.Config
	rules      []config.ForwardRule
	forwarders map[string]*forwarder.Forwarder
	language   config.Language
	mode       mode
	inputs     []inputField
	focusIndex int
	err        error
	confirmMsg string
	confirmYes bool
}

var translations = map[config.Language]map[string]string{
	config.Chinese: {
		"name":           "名称",
		"local_port":     "本地端口",
		"remote_addr":    "远程地址",
		"status":         "状态",
		"connections":    "连接数",
		"bytes_sent":     "发送流量",
		"bytes_recv":     "接收流量",
		"status_ok":      "正常",
		"status_fail":    "失败",
		"running":        "运行中",
		"stopped":        "已停止",
		"exit_hint":      "按 q 退出（Press L to switch to English）",
		"normal_hint":    "操作：[a]添加 [e]编辑 [d]删除 [s]启动/停止 [L]English [q]退出",
		"edit_hint":      "编辑模式：[enter]确认 [esc]取消 [tab]切换字段",
		"add_hint":       "添加模式：[enter]确认 [esc]取消 [tab]切换字段",
		"name_label":     "名称：",
		"lport_label":    "本地端口：",
		"rhost_label":    "远程主机：",
		"rport_label":    "远程端口：",
		"confirm_delete": "确认删除转发规则 '%s' 吗？[y/N]",
		"yes":            "是",
		"no":             "否",
		"confirm_title":  "⚠️  删除确认",
		"confirm_warn":   "此操作无法撤销！",
		"confirm_yes":    "是 (Y)",
		"confirm_no":     "否 (N)",
		"error_label":    "错误: ",
		"invalid_lport":  "无效的本地端口",
		"invalid_rport":  "无效的远程端口",
		"confirm_keys":   "← →/h l 切换  Enter 确认  Esc 取消",
	},
	config.English: {
		"name":           "Name",
		"local_port":     "Local Port",
		"remote_addr":    "Remote Addr",
		"status":         "Status",
		"connections":    "Connections",
		"bytes_sent":     "Bytes Sent",
		"bytes_recv":     "Bytes Recv",
		"status_ok":      "OK",
		"status_fail":    "Failed",
		"running":        "Running",
		"stopped":        "Stopped",
		"exit_hint":      "Press q to exit（按 L 切换中文）",
		"normal_hint":    "Commands: [a]Add [e]Edit [d]Delete [s]Start/Stop [L]中文 [q]Exit",
		"edit_hint":      "Edit Mode: [enter]Confirm [esc]Cancel [tab]Switch Field",
		"add_hint":       "Add Mode: [enter]Confirm [esc]Cancel [tab]Switch Field",
		"name_label":     "Name: ",
		"lport_label":    "Local Port: ",
		"rhost_label":    "Remote Host: ",
		"rport_label":    "Remote Port: ",
		"confirm_delete": "Are you sure to delete forwarding rule '%s'? [y/N]",
		"yes":            "Yes",
		"no":             "No",
		"confirm_title":  "⚠️  Confirm Deletion",
		"confirm_warn":   "This action cannot be undone!",
		"confirm_yes":    "Yes (Y)",
		"confirm_no":     "No (N)",
		"error_label":    "Error: ",
		"invalid_lport":  "Invalid local port",
		"invalid_rport":  "Invalid remote port",
		"confirm_keys":   "← →/h l Switch  Enter Confirm  Esc Cancel",
	},
}

func (m model) tr(key string) string {
	if t, ok := translations[m.language][key]; ok {
		return t
	}
	return key
}

func NewModel(cfg *config.Config, forwarders map[string]*forwarder.Forwarder) model {
	m := model{
		config:     cfg,
		rules:      cfg.Rules,
		forwarders: forwarders,
		language:   config.Chinese,
		mode:       normalMode,
		confirmYes: false,
	}

	// 创建表格
	columns := []table.Column{
		{Title: m.tr("name"), Width: 20},
		{Title: m.tr("local_port"), Width: 10},
		{Title: m.tr("remote_addr"), Width: 30},
		{Title: m.tr("status"), Width: 10},
		{Title: m.tr("connections"), Width: 10},
		{Title: m.tr("bytes_sent"), Width: 15},
		{Title: m.tr("bytes_recv"), Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
		table.WithKeyMap(table.DefaultKeyMap()),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m.table = t
	m.initInputs()
	m.updateRows()
	return m
}

func (m *model) initInputs() {
	m.inputs = []inputField{
		{textinput.New(), m.tr("name_label")},
		{textinput.New(), m.tr("lport_label")},
		{textinput.New(), m.tr("rhost_label")},
		{textinput.New(), m.tr("rport_label")},
	}

	for i := range m.inputs {
		m.inputs[i].textinput.Focus()
		m.inputs[i].textinput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		m.inputs[i].textinput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	}
}

func (m *model) updateTable() {
	// 只更新列标题
	columns := []table.Column{
		{Title: m.tr("name"), Width: 20},
		{Title: m.tr("local_port"), Width: 10},
		{Title: m.tr("remote_addr"), Width: 30},
		{Title: m.tr("status"), Width: 10},
		{Title: m.tr("connections"), Width: 10},
		{Title: m.tr("bytes_sent"), Width: 15},
		{Title: m.tr("bytes_recv"), Width: 15},
	}
	m.table.SetColumns(columns)
	m.updateRows()
}

func (m *model) updateRows() {
	var rows []table.Row
	for _, rule := range m.rules {
		status := m.tr("running")
		if !rule.IsRunning {
			status = m.tr("stopped")
		}
		if rule.Error != "" {
			status = m.tr("status_fail")
		}

		rows = append(rows, table.Row{
			rule.Name,
			fmt.Sprintf("%d", rule.LocalPort),
			fmt.Sprintf("%s:%d", rule.RemoteHost, rule.RemotePort),
			status,
			fmt.Sprintf("%d", rule.Connections),
			formatBytes(rule.BytesSent),
			formatBytes(rule.BytesRecv),
		})
	}
	m.table.SetRows(rows)
}

func (m *model) startForwarder(rule *config.ForwardRule) error {
	f := forwarder.NewForwarder(rule)
	if err := f.Start(); err != nil {
		return err
	}
	rule.IsRunning = true
	m.forwarders[rule.Name] = f
	return nil
}

func (m *model) stopForwarder(rule *config.ForwardRule) {
	if f, ok := m.forwarders[rule.Name]; ok {
		f.Stop()
		delete(m.forwarders, rule.Name)
	}
	rule.IsRunning = false
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case normalMode:
			switch msg.String() {
			case "up", "down", "k", "j":
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			case "q", "ctrl+c":
				return m, tea.Quit
			case "L", "l":
				if m.language == config.Chinese {
					m.language = config.English
				} else {
					m.language = config.Chinese
				}
				m.updateTable()
				// 更新输入框标签
				for i := range m.inputs {
					switch i {
					case 0:
						m.inputs[i].label = m.tr("name_label")
					case 1:
						m.inputs[i].label = m.tr("lport_label")
					case 2:
						m.inputs[i].label = m.tr("rhost_label")
					case 3:
						m.inputs[i].label = m.tr("rport_label")
					}
				}
				return m, nil
			case "a":
				m.mode = addMode
				m.focusIndex = 0
				for i := range m.inputs {
					m.inputs[i].textinput.SetValue("")
					if i == 0 {
						m.inputs[i].textinput.Focus()
					} else {
						m.inputs[i].textinput.Blur()
					}
				}
			case "e":
				if len(m.rules) > 0 {
					m.mode = editMode
					m.focusIndex = 0
					rule := m.rules[m.table.Cursor()]
					m.inputs[0].textinput.SetValue(rule.Name)
					m.inputs[1].textinput.SetValue(fmt.Sprintf("%d", rule.LocalPort))
					m.inputs[2].textinput.SetValue(rule.RemoteHost)
					m.inputs[3].textinput.SetValue(fmt.Sprintf("%d", rule.RemotePort))
					m.inputs[0].textinput.Focus()
					for i := 1; i < len(m.inputs); i++ {
						m.inputs[i].textinput.Blur()
					}
				}
			case "d":
				if len(m.rules) > 0 {
					idx := m.table.Cursor()
					rule := m.rules[idx]
					m.confirmMsg = fmt.Sprintf(m.tr("confirm_delete"), rule.Name)
					m.mode = confirmMode
					m.confirmYes = false
				}
			case "s":
				if len(m.rules) > 0 {
					idx := m.table.Cursor()
					rule := &m.rules[idx]
					if rule.IsRunning {
						m.stopForwarder(rule)
					} else {
						if err := m.startForwarder(rule); err != nil {
							rule.Error = err.Error()
						} else {
							rule.Error = ""
						}
					}
				}
			}
		case confirmMode:
			switch msg.String() {
			case "y", "Y", "right", "l":
				m.confirmYes = false
			case "n", "N", "left", "h":
				m.confirmYes = true
			case "enter":
				if m.confirmYes {
					idx := m.table.Cursor()
					rule := &m.rules[idx]
					m.stopForwarder(rule)
					if err := m.config.DeleteRule(idx); err != nil {
						m.err = err
					} else {
						m.rules = m.config.Rules
					}
				}
				m.mode = normalMode
			case "esc":
				m.mode = normalMode
			}
			return m, nil
		case addMode, editMode:
			switch msg.String() {
			case "esc":
				m.mode = normalMode
				m.err = nil
			case "tab", "shift+tab":
				s := msg.String()
				if s == "tab" {
					m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
				} else {
					m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
				}

				for i := range m.inputs {
					if i == m.focusIndex {
						m.inputs[i].textinput.Focus()
					} else {
						m.inputs[i].textinput.Blur()
					}
				}
			case "enter":
				var rule config.ForwardRule
				var err error

				rule.Name = m.inputs[0].textinput.Value()
				rule.LocalPort, err = strconv.Atoi(m.inputs[1].textinput.Value())
				if err != nil {
					m.err = fmt.Errorf(m.tr("invalid_lport"))
					break
				}
				rule.RemoteHost = m.inputs[2].textinput.Value()
				rule.RemotePort, err = strconv.Atoi(m.inputs[3].textinput.Value())
				if err != nil {
					m.err = fmt.Errorf(m.tr("invalid_rport"))
					break
				}

				if m.mode == addMode {
					if err := m.config.AddRule(rule); err != nil {
						m.err = err
						break
					}
				} else {
					if err := m.config.UpdateRule(m.table.Cursor(), rule); err != nil {
						m.err = err
						break
					}
				}

				m.rules = m.config.Rules
				m.mode = normalMode
				m.err = nil
			}
		}
	}

	if m.mode == addMode || m.mode == editMode {
		for i := range m.inputs {
			m.inputs[i].textinput, cmd = m.inputs[i].textinput.Update(msg)
		}
	} else if m.mode == normalMode {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var view string

	switch m.mode {
	case normalMode:
		m.updateRows() // 更新行数据
		view = baseStyle.Render(m.table.View())

		// 错误信息显示
		for _, rule := range m.rules {
			if rule.Error != "" {
				view += fmt.Sprintf("\n%s: %s", rule.Name, rule.Error)
			}
		}

		view += "\n" + m.tr("normal_hint")
	case confirmMode:
		// 构建确认对话框
		var confirmBox strings.Builder

		// 标题
		confirmBox.WriteString(m.tr("confirm_title") + "\n\n")

		// 确认消息
		confirmBox.WriteString(fmt.Sprintf(m.tr("confirm_delete"), m.rules[m.table.Cursor()].Name) + "\n")
		confirmBox.WriteString(warningStyle.Render(m.tr("confirm_warn")) + "\n\n")

		// 按钮
		var yesButton, noButton string
		if m.confirmYes {
			yesButton = selectedButtonStyle.Render(m.tr("confirm_yes"))
			noButton = unselectedButtonStyle.Render(m.tr("confirm_no"))
		} else {
			yesButton = unselectedButtonStyle.Render(m.tr("confirm_yes"))
			noButton = selectedButtonStyle.Render(m.tr("confirm_no"))
		}

		buttons := fmt.Sprintf("%s  %s", yesButton, noButton)
		confirmBox.WriteString(buttons + "\n\n")

		// 添加操作提示
		hint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(m.tr("confirm_keys"))
		confirmBox.WriteString(hint)

		// 只显示确认框，不显示表格
		view = "\n\n" + confirmStyle.Render(confirmBox.String())
	case addMode, editMode:
		var b strings.Builder
		hint := m.tr("add_hint")
		if m.mode == editMode {
			hint = m.tr("edit_hint")
		}

		for i := range m.inputs {
			b.WriteString(m.inputs[i].label)
			b.WriteString(m.inputs[i].textinput.View())
			b.WriteString("\n")
		}

		if m.err != nil {
			b.WriteString(fmt.Sprintf("\n%s%v\n", m.tr("error_label"), m.err))
		}

		b.WriteString("\n" + hint)
		view = b.String()
	}

	return view
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func StartUI(cfg *config.Config, forwarders map[string]*forwarder.Forwarder) error {
	p := tea.NewProgram(NewModel(cfg, forwarders))
	_, err := p.Run()
	return err
}
