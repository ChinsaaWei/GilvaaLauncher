package tui

import (
	"github.com/charmbracelet/bubbletea"
)

func Start() {
	p := tea.NewProgram(NewModel())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
