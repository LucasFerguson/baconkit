package main

import (
	"fmt"
	"os"

	"baconkit/scans"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table           table.Model
	fullRows        []table.Row
	activeView      string
	selectedProcess table.Row
}

func (m model) Init() tea.Cmd { return nil }

func (m *model) syncSelectedProcess() {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.fullRows) {
		return
	}
	m.selectedProcess = m.fullRows[idx]
}

func (m *model) openSelectedProcess() {
	m.syncSelectedProcess()
	if len(m.selectedProcess) == 0 {
		return
	}
	m.activeView = "process"
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch m.activeView {
		case "process":
			switch msg.String() {
			case "b", "backspace", "esc":
				m.activeView = "list"
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		default:
			switch msg.String() {
			case "esc":
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.openSelectedProcess()
			}
		}
	case tea.MouseClickMsg:
		if m.activeView == "list" && msg.Mouse().Button == tea.MouseLeft {
			m.openSelectedProcess()
		}
	}
	if m.activeView == "list" {
		m.table, cmd = m.table.Update(msg)
		m.syncSelectedProcess()
	}
	return m, cmd
}

func (m model) View() tea.View {
	if m.activeView == "process" && len(m.selectedProcess) > 0 {
		body := fmt.Sprintf(
			"Process Detail\n\nRank: %s\nName: %s\nCountry: %s\nPopulation: %s\n\nPress b or esc to go back",
			m.selectedProcess[0],
			m.selectedProcess[1],
			m.selectedProcess[2],
			m.selectedProcess[3],
		)
		v := tea.NewView(baseStyle.Render(body) + "\n")
		v.MouseMode = tea.MouseModeCellMotion
		return v
	}

	rightBody := "No row selected"
	if len(m.selectedProcess) > 0 {
		rightBody = fmt.Sprintf(
			"Selected Process / Row\n\nRank: %s\nName: %s\nCountry: %s\nPopulation: %s\n\nPress enter to open full view",
			m.selectedProcess[0],
			m.selectedProcess[1],
			m.selectedProcess[2],
			m.selectedProcess[3],
		)
	}

	left := baseStyle.Render(m.table.View() + "\n  " + m.table.HelpView())
	right := baseStyle.Width(46).Render(rightBody)
	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right) + "\n")
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "deb" {
		scans.Deb()
		return
	}

	columns := []table.Column{
		{Title: "Rank", Width: 6},
		{Title: "Task", Width: 24},
	}

	fullRows := sampleRows()
	rows := make([]table.Row, 0, len(fullRows))
	for _, row := range fullRows {
		rows = append(rows, table.Row{row[0], row[1]})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(30),
		table.WithWidth(32),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{table: t, fullRows: fullRows, activeView: "list"}
	m.syncSelectedProcess()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
