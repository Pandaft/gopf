package ui

import (
	"fmt"
	"gopf/config"
	"gopf/forwarder"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

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

	// 禁用状态的按钮样式 - 使用更暗的灰色
	disabledButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF06B7"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#767676")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
)

type inputField struct {
	textinput textinput.Model
	label     string
	validate  func(string) error
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
	version    string
}

var translations = map[config.Language]map[string]string{
	config.Chinese: {
		"name":               "名称",
		"local_port":         "本地端口",
		"remote_addr":        "远程地址",
		"status":             "状态",
		"connections":        "连接数",
		"bytes_sent":         "发送流量",
		"bytes_recv":         "接收流量",
		"status_ok":          "正常",
		"status_fail":        "失败",
		"running":            "运行中",
		"stopped":            "已停止",
		"exit_hint":          "按 %s 退出",
		"normal_hint":        "操作：%s添加 %s编辑 %s删除 %s启动/停止 %s清空统计 %sEnglish %s退出",
		"edit_hint":          "编辑模式：%s确认 %s取消 %s切换字段",
		"add_hint":           "添加模式：%s确认 %s取消 %s切换字段",
		"name_label":         "名称：",
		"lport_label":        "本地端口：",
		"rhost_label":        "远程主机：",
		"rport_label":        "远程端口：",
		"confirm_delete":     "确认删除转发规则 '%s' 吗？[y/N]",
		"yes":                "是",
		"no":                 "否",
		"confirm_title":      "⚠️  删除确认",
		"confirm_warn":       "此操作无法撤销！",
		"confirm_yes":        "是 (Y)",
		"confirm_no":         "否 (N)",
		"error_label":        "错误: ",
		"invalid_lport":      "无效的本地端口",
		"invalid_rport":      "无效的远程端口",
		"confirm_keys":       "← →/h l 切换  Enter 确认  Esc 取消",
		"stats_cleared":      "已清空统计数据",
		"last_active":        "最后活跃",
		"just_now":           "刚刚",
		"seconds_ago":        "%d秒前",
		"minutes_ago":        "%d分钟前",
		"hours_ago":          "%d小时前",
		"days_ago":           "%d天前",
		"never":              "从未活跃",
		"add_title":          "添加转发规则",
		"edit_title":         "编辑转发规则",
		"err_empty_name":     "名称不能为空",
		"err_numeric_port":   "端口必须是数字",
		"err_port_range":     "端口必须在 1-65535 之间",
		"err_empty_host":     "主机地址不能为空",
		"please_select_rule": "请先选择一个规则",
	},
	config.English: {
		"name":               "Name",
		"local_port":         "Local Port",
		"remote_addr":        "Remote Addr",
		"status":             "Status",
		"connections":        "Connections",
		"bytes_sent":         "Bytes Sent",
		"bytes_recv":         "Bytes Recv",
		"status_ok":          "OK",
		"status_fail":        "Failed",
		"running":            "Running",
		"stopped":            "Stopped",
		"exit_hint":          "Press %s to exit",
		"normal_hint":        "Commands: %sAdd %sEdit %sDelete %sStart/Stop %sClear Stats %s中文 %sExit",
		"edit_hint":          "Edit Mode: %sConfirm %sCancel %sSwitch Field",
		"add_hint":           "Add Mode: %sConfirm %sCancel %sSwitch Field",
		"name_label":         "Name: ",
		"lport_label":        "Local Port: ",
		"rhost_label":        "Remote Host: ",
		"rport_label":        "Remote Port: ",
		"confirm_delete":     "Are you sure to delete forwarding rule '%s'? [y/N]",
		"yes":                "Yes",
		"no":                 "No",
		"confirm_title":      "⚠️  Confirm Deletion",
		"confirm_warn":       "This action cannot be undone!",
		"confirm_yes":        "Yes (Y)",
		"confirm_no":         "No (N)",
		"error_label":        "Error: ",
		"invalid_lport":      "Invalid local port",
		"invalid_rport":      "Invalid remote port",
		"confirm_keys":       "← →/h l Switch  Enter Confirm  Esc Cancel",
		"stats_cleared":      "Statistics cleared",
		"last_active":        "Last Active",
		"just_now":           "just now",
		"seconds_ago":        "%ds ago",
		"minutes_ago":        "%dm ago",
		"hours_ago":          "%dh ago",
		"days_ago":           "%dd ago",
		"never":              "never",
		"add_title":          "Add Forwarding Rule",
		"edit_title":         "Edit Forwarding Rule",
		"err_empty_name":     "Name cannot be empty",
		"err_numeric_port":   "Port must be numeric",
		"err_port_range":     "Port must be between 1-65535",
		"err_empty_host":     "Host address cannot be empty",
		"please_select_rule": "Please select a rule first",
	},
}

func (m model) tr(key string) string {
	if t, ok := translations[m.language][key]; ok {
		return t
	}
	return key
}

func formatLastActive(lastActive int64, tr func(string) string) string {
	if lastActive == 0 {
		return tr("never")
	}

	now := time.Now().Unix()
	diff := now - lastActive

	switch {
	case diff < 5:
		return tr("just_now")
	case diff < 60:
		return fmt.Sprintf(tr("seconds_ago"), diff)
	case diff < 3600:
		return fmt.Sprintf(tr("minutes_ago"), diff/60)
	case diff < 86400:
		return fmt.Sprintf(tr("hours_ago"), diff/3600)
	default:
		return fmt.Sprintf(tr("days_ago"), diff/86400)
	}
}

func NewModel(cfg *config.Config, forwarders map[string]*forwarder.Forwarder, version string) model {
	m := model{
		config:     cfg,
		rules:      cfg.Rules,
		forwarders: forwarders,
		language:   config.Chinese,
		mode:       normalMode,
		confirmYes: false,
		version:    version,
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
		{Title: m.tr("last_active"), Width: 15},
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
		{textinput.New(), m.tr("name_label"), m.validateName},
		{textinput.New(), m.tr("lport_label"), m.validatePort},
		{textinput.New(), m.tr("rhost_label"), m.validateHost},
		{textinput.New(), m.tr("rport_label"), m.validatePort},
	}

	for i := range m.inputs {
		m.inputs[i].textinput.Focus()
		m.inputs[i].textinput.PromptStyle = inputStyle
		m.inputs[i].textinput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

		// 设置占位符文本
		switch i {
		case 0:
			m.inputs[i].textinput.Placeholder = "my-forward"
		case 1, 3:
			m.inputs[i].textinput.Placeholder = "1-65535"
		case 2:
			m.inputs[i].textinput.Placeholder = "example.com"
		}
	}
}

func (m *model) updateTable() {
	columns := []table.Column{
		{Title: m.tr("name"), Width: 20},
		{Title: m.tr("local_port"), Width: 10},
		{Title: m.tr("remote_addr"), Width: 30},
		{Title: m.tr("status"), Width: 10},
		{Title: m.tr("connections"), Width: 10},
		{Title: m.tr("bytes_sent"), Width: 15},
		{Title: m.tr("bytes_recv"), Width: 15},
		{Title: m.tr("last_active"), Width: 15},
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
			formatLastActive(rule.LastActive, m.tr),
		})
	}
	m.table.SetRows(rows)

	if len(rows) > 0 {
		cursor := m.table.Cursor()
		if cursor >= len(rows) {
			m.table.SetCursor(len(rows) - 1)
		}
	}
}

func (m *model) startForwarder(rule *config.ForwardRule) error {
	// 如果已经在运行，先停止
	if rule.IsRunning {
		m.stopForwarder(rule)
	}

	// 创建新的转发器
	f := forwarder.NewForwarder(rule)
	if err := f.Start(); err != nil {
		rule.IsRunning = false
		return err
	}

	rule.IsRunning = true
	rule.Error = ""
	m.forwarders[rule.Name] = f
	return nil
}

func (m *model) stopForwarder(rule *config.ForwardRule) {
	// 通过规则名称查找并停止转发器
	if f, ok := m.forwarders[rule.Name]; ok {
		f.Stop()
		delete(m.forwarders, rule.Name)
	}

	// 通过本地端口查找并停止可能存在的其他转发器
	for name, f := range m.forwarders {
		if f.GetLocalPort() == rule.LocalPort {
			f.Stop()
			delete(m.forwarders, name)
		}
	}

	rule.IsRunning = false
	rule.Error = ""
}

func (m *model) clearStats(rule *config.ForwardRule) {
	// 如果转发器正在运行，使用转发器的清空方法
	if f, ok := m.forwarders[rule.Name]; ok {
		f.ClearStats()
	} else {
		// 如果转发器没有运行，直接清空规则中的统计数据
		atomic.StoreUint64(&rule.BytesSent, 0)
		atomic.StoreUint64(&rule.BytesRecv, 0)
		atomic.StoreUint64(&rule.Connections, 0)
	}
}

type tickMsg time.Time

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		m.updateRows()
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.KeyMsg:
		switch m.mode {
		case normalMode:
			switch msg.String() {
			case "up", "down", "k", "j":
				if len(m.rules) > 0 {
					m.table, cmd = m.table.Update(msg)
				}
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
				m.err = nil
				for i := range m.inputs {
					m.inputs[i].textinput.Reset()
					if i == 0 {
						m.inputs[i].textinput.Focus()
					} else {
						m.inputs[i].textinput.Blur()
					}
				}
				return m, nil
			case "e", "d", "s", "c":
				if len(m.rules) == 0 {
					m.err = fmt.Errorf(m.tr("please_select_rule"))
					return m, nil
				}

				switch msg.String() {
				case "e":
					m.mode = editMode
					m.focusIndex = 0
					m.err = nil
					rule := m.rules[m.table.Cursor()]
					for i := range m.inputs {
						m.inputs[i].textinput.Reset()
						if i == 0 {
							m.inputs[i].textinput.Focus()
						} else {
							m.inputs[i].textinput.Blur()
						}
					}
					m.inputs[0].textinput.SetValue(rule.Name)
					m.inputs[1].textinput.SetValue(fmt.Sprintf("%d", rule.LocalPort))
					m.inputs[2].textinput.SetValue(rule.RemoteHost)
					m.inputs[3].textinput.SetValue(fmt.Sprintf("%d", rule.RemotePort))
				case "d":
					idx := m.table.Cursor()
					rule := m.rules[idx]
					m.confirmMsg = fmt.Sprintf(m.tr("confirm_delete"), rule.Name)
					m.mode = confirmMode
					m.confirmYes = false
				case "s":
					idx := m.table.Cursor()
					if idx >= 0 && idx < len(m.rules) {
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
				case "c":
					idx := m.table.Cursor()
					rule := &m.rules[idx]
					m.clearStats(rule)
					m.updateRows()
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
						m.updateRows()
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
				// 验证所有输入
				var hasError bool
				for i := range m.inputs {
					if err := m.inputs[i].validate(m.inputs[i].textinput.Value()); err != nil {
						m.err = err
						hasError = true
						break
					}
				}

				if hasError {
					break
				}

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
					m.rules = m.config.Rules

					newRule := &m.rules[(len(m.rules) - 1)]
					if err := m.startForwarder(newRule); err != nil {
						newRule.Error = err.Error()
					}
				} else {
					// 在更新规则前，先停止旧的转发器
					idx := m.table.Cursor()
					oldRule := &m.rules[idx]
					wasRunning := oldRule.IsRunning
					oldPort := oldRule.LocalPort

					// 先停止旧的转发
					if wasRunning {
						m.stopForwarder(oldRule)
					}

					// 更新规则
					if err := m.config.UpdateRule(idx, rule); err != nil {
						// 如果更新失败，且原来是运行状态，则恢复旧的转发
						if wasRunning {
							oldRule.LocalPort = oldPort // 恢复旧端口
							if err := m.startForwarder(oldRule); err != nil {
								oldRule.Error = err.Error()
							}
						}
						m.err = err
						break
					}

					m.rules = m.config.Rules

					// 如果原来是运行状态，启动新的转发
					if wasRunning {
						newRule := &m.rules[idx]
						if err := m.startForwarder(newRule); err != nil {
							newRule.Error = err.Error()
						}
					}
				}

				m.mode = normalMode
				m.err = nil
				m.updateRows()
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

	// 添加标题和版本信息
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginLeft(2).
		Render("GOPF")
	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Render(m.version)
	view = fmt.Sprintf("%s %s\n", title, version)

	switch m.mode {
	case normalMode:
		view += baseStyle.Render(m.table.View())

		// 错误信息显示
		if m.err != nil {
			view += "\n" + errorStyle.Render(m.err.Error())
		}

		for _, rule := range m.rules {
			if rule.Error != "" {
				view += fmt.Sprintf("\n%s: %s", rule.Name, rule.Error)
			}
		}

		// 检查是否有规则
		hasRules := len(m.rules) > 0

		// 根据是否有规则选择按钮样式
		addKey := keyStyle.Render("[a]")
		var editKey, deleteKey, startStopKey, clearKey, langKey, quitKey string

		if hasRules {
			editKey = keyStyle.Render("[e]")
			deleteKey = keyStyle.Render("[d]")
			startStopKey = keyStyle.Render("[s]")
			clearKey = keyStyle.Render("[c]")
		} else {
			editKey = disabledButtonStyle.Render("[e]")
			deleteKey = disabledButtonStyle.Render("[d]")
			startStopKey = disabledButtonStyle.Render("[s]")
			clearKey = disabledButtonStyle.Render("[c]")
		}
		langKey = keyStyle.Render("[L]")
		quitKey = keyStyle.Render("[q]")

		hint := fmt.Sprintf(m.tr("normal_hint"),
			addKey,
			editKey,
			deleteKey,
			startStopKey,
			clearKey,
			langKey,
			quitKey,
		)
		view += "\n" + hint

	case addMode, editMode:
		var b strings.Builder

		hint := fmt.Sprintf(m.tr("add_hint"),
			keyStyle.Render("[enter]"),
			keyStyle.Render("[esc]"),
			keyStyle.Render("[tab]"),
		)
		if m.mode == editMode {
			hint = fmt.Sprintf(m.tr("edit_hint"),
				keyStyle.Render("[enter]"),
				keyStyle.Render("[esc]"),
				keyStyle.Render("[tab]"),
			)
		}

		// 添加标题
		title := m.tr("add_title")
		if m.mode == editMode {
			title = m.tr("edit_title")
		}
		b.WriteString(labelStyle.Render(title) + "\n\n")

		// 渲染输入框
		for i := range m.inputs {
			// 标签
			b.WriteString(labelStyle.Render(m.inputs[i].label) + "\n")

			// 输入框
			b.WriteString(m.inputs[i].textinput.View() + "\n")

			// 验证错误
			if m.inputs[i].textinput.Value() != "" {
				if err := m.inputs[i].validate(m.inputs[i].textinput.Value()); err != nil {
					b.WriteString(errorStyle.Render("  ↑ "+err.Error()) + "\n")
				}
			}

			b.WriteString("\n")
		}

		if m.err != nil {
			b.WriteString("\n" + errorStyle.Render(m.tr("error_label")+m.err.Error()) + "\n")
		}

		b.WriteString("\n" + hint)
		view = b.String()

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

func StartUI(cfg *config.Config, forwarders map[string]*forwarder.Forwarder, version string) error {
	p := tea.NewProgram(
		NewModel(cfg, forwarders, version),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}

func (m model) validateName(s string) error {
	if s == "" {
		return fmt.Errorf(m.tr("err_empty_name"))
	}
	return nil
}

func (m model) validatePort(s string) error {
	port, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf(m.tr("err_numeric_port"))
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf(m.tr("err_port_range"))
	}
	return nil
}

func (m model) validateHost(s string) error {
	if s == "" {
		return fmt.Errorf(m.tr("err_empty_host"))
	}
	return nil
}
