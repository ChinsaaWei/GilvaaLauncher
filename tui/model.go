package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	quit bool
}

func NewModel() *model {
	return &model{}
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true).
			PaddingLeft(2)

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingLeft(4)
)

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() string {
	if m.quit {
		return ""
	}

	return titleStyle.Render("Gilvaa Launcher") + "\n" + contentStyle.Render("Press 'q' or 'Ctrl+C' to quit\n")
}