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
}

func NewModel(rules []config.ForwardRule, forwarders map[string]*forwarder.Forwarder) model {
	columns := []table.Column{
		{Title: "名称", Width: 20},
		{Title: "本地端口", Width: 10},
		{Title: "远程地址", Width: 30},
		{Title: "状态", Width: 10},
		{Title: "连接数", Width: 10},
		{Title: "发送流量", Width: 15},
		{Title: "接收流量", Width: 15},
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

	return model{
		table:      t,
		rules:      rules,
		forwarders: forwarders,
	}
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
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var rows []table.Row
	for _, rule := range m.rules {
		status := rule.Status
		if rule.Error != "" {
			status = "失败"
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

	return view + "\n按 q 退出"
}

func StartUI(rules []config.ForwardRule, forwarders map[string]*forwarder.Forwarder) error {
	p := tea.NewProgram(NewModel(rules, forwarders))
	_, err := p.Run()
	return err
}
