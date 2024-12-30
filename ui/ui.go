package ui

import (
	"fmt"
	"gopf/config"
	"gopf/forwarder"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table      table.Model
	rules      []config.ForwardRule
	forwarders map[string]*forwarder.Forwarder
	language   config.Language
}

var translations = map[config.Language]map[string]string{
	config.Chinese: {
		"name":        "名称",
		"local_port":  "本地端口",
		"remote_addr": "远程地址",
		"status":      "状态",
		"connections": "连接数",
		"bytes_sent":  "发送流量",
		"bytes_recv":  "接收流量",
		"status_ok":   "正常",
		"status_fail": "失败",
		"exit_hint":   "按 q 退出（Press L to switch to English）",
	},
	config.English: {
		"name":        "Name",
		"local_port":  "Local Port",
		"remote_addr": "Remote Address",
		"status":      "Status",
		"connections": "Connections",
		"bytes_sent":  "Bytes Sent",
		"bytes_recv":  "Bytes Recv",
		"status_ok":   "OK",
		"status_fail": "Failed",
		"exit_hint":   "Press q to exit（按 L 切换中文）",
	},
}

func (m model) tr(key string) string {
	if t, ok := translations[m.language][key]; ok {
		return t
	}
	return key
}

func NewModel(rules []config.ForwardRule, forwarders map[string]*forwarder.Forwarder) model {
	m := model{
		rules:      rules,
		forwarders: forwarders,
		language:   config.Chinese,
	}

	m.updateTable()
	return m
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
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "L", "l":
			if m.language == config.Chinese {
				m.language = config.English
			} else {
				m.language = config.Chinese
			}
			m.updateTable()
			return m, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var rows []table.Row
	for _, rule := range m.rules {
		status := m.tr("status_ok")
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

	view := baseStyle.Render(m.table.View())

	// 错误信息显示
	for _, rule := range m.rules {
		if rule.Error != "" {
			view += fmt.Sprintf("\n%s: %s", rule.Name, rule.Error)
		}
	}

	return view + "\n" + m.tr("exit_hint")
}

func StartUI(rules []config.ForwardRule, forwarders map[string]*forwarder.Forwarder) error {
	p := tea.NewProgram(NewModel(rules, forwarders))
	_, err := p.Run()
	return err
}
