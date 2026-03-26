package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type refreshMsg struct{}

func scheduleRefresh() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return refreshMsg{}
	})
}

var debugLog *log.Logger

func initDebugLog() {
	f, err := os.OpenFile("tmp/baconkit.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		debugLog = log.New(f, "", log.Ltime)
	}
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

const rightPanelWidth = 48 // 46 content + 2 border chars

type model struct {
	table           table.Model
	fullRows        []table.Row
	activeView      string
	selectedProcess table.Row
	width           int
	height          int
	sortCol         int
	sortAsc         bool
}

// colNames maps column index to display title (without sort indicator).
var colNames = []string{"PID", "Name", "User", "State"}

func (m *model) sorted() []table.Row {
	rows := make([]table.Row, len(m.fullRows))
	copy(rows, m.fullRows)
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i][m.sortCol], rows[j][m.sortCol]
		less := a < b
		if m.sortCol == 0 { // PID: numeric
			ai, _ := strconv.Atoi(a)
			bi, _ := strconv.Atoi(b)
			less = ai < bi
		}
		if m.sortAsc {
			return less
		}
		return !less
	})
	return rows
}

func (m *model) colTitle(idx int) string {
	title := colNames[idx]
	if idx == m.sortCol {
		if m.sortAsc {
			title += " ▲"
		} else {
			title += " ▼"
		}
	}
	return title
}

func (m *model) resizeTable() {
	if m.width == 0 || m.height == 0 {
		return
	}

	tableW := m.width - rightPanelWidth - 1 - 2 // gap + left+right border
	tableH := m.height - 6
	if tableH < 1 {
		tableH = 1
	}
	if tableW < 2 {
		return
	}

	// Each column renders as col.Width+2 on screen (Padding(0,1) on both Header and Cell).
	// Subtract that overhead before allocating column content widths.
	const (
		pidW     = 8
		userW    = 14
		stateW   = 16
		minNameW = 10
		maxNameW = 30
		cellPad  = 2 // per-column screen overhead
	)
	sorted := m.sorted()
	var cols []table.Column
	var rows []table.Row
	var actualTableW int
	switch {
	case tableW >= pidW+minNameW+userW+stateW+4*cellPad:
		nameW := min(tableW-4*cellPad-pidW-userW-stateW, maxNameW)
		actualTableW = pidW + nameW + userW + stateW + 4*cellPad
		cols = []table.Column{
			{Title: m.colTitle(0), Width: pidW},
			{Title: m.colTitle(1), Width: nameW},
			{Title: m.colTitle(2), Width: userW},
			{Title: m.colTitle(3), Width: stateW},
		}
		for _, r := range sorted {
			rows = append(rows, table.Row{r[0], r[1], r[2], r[3]})
		}
	case tableW >= pidW+minNameW+userW+3*cellPad:
		nameW := min(tableW-3*cellPad-pidW-userW, maxNameW)
		actualTableW = pidW + nameW + userW + 3*cellPad
		cols = []table.Column{
			{Title: m.colTitle(0), Width: pidW},
			{Title: m.colTitle(1), Width: nameW},
			{Title: m.colTitle(2), Width: userW},
		}
		for _, r := range sorted {
			rows = append(rows, table.Row{r[0], r[1], r[2]})
		}
	case tableW >= pidW+minNameW+2*cellPad:
		nameW := min(tableW-2*cellPad-pidW, maxNameW)
		actualTableW = pidW + nameW + 2*cellPad
		cols = []table.Column{
			{Title: m.colTitle(0), Width: pidW},
			{Title: m.colTitle(1), Width: nameW},
		}
		for _, r := range sorted {
			rows = append(rows, table.Row{r[0], r[1]})
		}
	default:
		nameW := tableW - cellPad
		if nameW < 1 {
			nameW = 1
		}
		actualTableW = nameW + cellPad
		cols = []table.Column{
			{Title: m.colTitle(1), Width: nameW},
		}
		for _, r := range sorted {
			rows = append(rows, table.Row{r[1]})
		}
	}

	if debugLog != nil {
		widths := make([]int, len(cols))
		for i, c := range cols {
			widths[i] = c.Width
		}
		debugLog.Printf("window=%dx%d tableW=%d actualTableW=%d tableH=%d cols=%d widths=%v", m.width, m.height, tableW, actualTableW, tableH, len(cols), widths)
	}

	m.table.SetRows([]table.Row{}) // clear first to avoid col/row count mismatch
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
	m.table.SetWidth(actualTableW)
	m.table.SetHeight(tableH)
}

func scanToRows(scanOut map[int]map[string]string) []table.Row {
	var rows []table.Row
	for pid, procDetails := range scanOut {
		rows = append(rows, table.Row{strconv.Itoa(pid), procDetails["Name"], procDetails["User"], procDetails["State"]})
	}
	return rows
}

func (m model) Init() tea.Cmd { return scheduleRefresh() }

func (m *model) syncSelectedProcess() {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.fullRows) {
		return
	}
	m.selectedProcess = m.fullRows[idx]
}

func (m *model) refresh() {
	cursor := m.table.Cursor()
	m.fullRows = scanToRows(loadProcesses())
	m.resizeTable()
	m.table.SetCursor(cursor)
	m.syncSelectedProcess()
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
			case "r":
				m.refresh()
			case "1", "2", "3", "4":
				col := int(msg.String()[0] - '1')
				if col == m.sortCol {
					m.sortAsc = !m.sortAsc
				} else {
					m.sortCol = col
					m.sortAsc = true
				}
				m.resizeTable()
			}
		}
	case refreshMsg:
		m.refresh()
		return m, scheduleRefresh()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeTable()
		if debugLog != nil {
			tableOuterW := m.table.Width() + 2
			rightContentWidth := m.width - tableOuterW - 1 - 2
			if rightContentWidth < 1 {
				rightContentWidth = 1
			}
			left := baseStyle.Width(m.table.Width()).Render(m.table.View() + "\n  " + m.table.HelpView())
			right := baseStyle.Width(rightContentWidth).Render("(right panel)")
			debugLog.Printf("full render (window=%dx%d):\n%s",
				m.width, m.height,
				lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right))
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
			"Process Detail\n\nPID: %s\nName: %s\nUser: %s\nState: %s\n\nPress b or esc to go back",
			m.selectedProcess[0],
			m.selectedProcess[1],
			m.selectedProcess[2],
			m.selectedProcess[3],
		)
		v := tea.NewView(baseStyle.Width(m.width).Height(m.height).Render(body) + "\n")
		v.AltScreen = true
		return v
	}

	rightBody := "No row selected"
	if len(m.selectedProcess) > 0 {
		rightBody = fmt.Sprintf(
			"Selected Process\n\nPID: %s\nName: %s\nUser: %s\nState: %s\n\nPress enter to open full view",
			m.selectedProcess[0],
			m.selectedProcess[1],
			m.selectedProcess[2],
			m.selectedProcess[3],
		)
	}

	tableOuterW := m.table.Width() + 2       // lipgloss Width(n) = outer screen width, inner = n-2
	rightPanelW := m.width - tableOuterW - 1 // fill remaining space after left panel + gap
	if rightPanelW < 3 {
		rightPanelW = 3
	}
	left := baseStyle.Width(tableOuterW).Render(m.table.View() + "\n  " + m.table.HelpView())
	right := baseStyle.Width(rightPanelW).Render(rightBody)
	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right) + "\n")
	v.AltScreen = true
	return v
}

func main() {
	initDebugLog()

	fullRows := scanToRows(loadProcesses())
	t := table.New(table.WithFocused(true))

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

	m := model{table: t, fullRows: fullRows, activeView: "list", sortCol: 0, sortAsc: true}
	m.syncSelectedProcess()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
