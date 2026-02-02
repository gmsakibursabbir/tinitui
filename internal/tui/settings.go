package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tinytui/tinytui/internal/config"
)

type settingsModel struct {
	cursor int
}

func newSettingsModel() settingsModel {
	return settingsModel{
		cursor: 0,
	}
}

func (m MainModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Options:
	// 0. Metadata (Toggle)
	// 1. Output Mode (Toggle Replace/Directory)
	// 2. Mascot (Off/On/Auto)
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.settings.cursor--
			if m.settings.cursor < 0 {
				m.settings.cursor = 2
			}
		case "down", "j":
			m.settings.cursor++
			if m.settings.cursor > 2 {
				m.settings.cursor = 0
			}
		case "enter", " ":
			// Toggle
			switch m.settings.cursor {
			case 0:
				m.config.Metadata = !m.config.Metadata
			case 1:
				if m.config.OutputMode == "replace" {
					m.config.OutputMode = "directory"
				} else {
					m.config.OutputMode = "replace"
				}
			case 2:
				// Cycle Off -> On -> Auto -> Off
				switch m.config.Mascot {
				case config.MascotOff:
					m.config.Mascot = config.MascotOn
				case config.MascotOn:
					m.config.Mascot = config.MascotAuto
				case config.MascotAuto:
					m.config.Mascot = config.MascotOff
				}
			}
			m.config.Save() // Autosave
		}
	}
	return m, nil
}

func (m MainModel) viewSettings() string {
	s := strings.Builder{}
	s.WriteString("Settings\n\n")

	// Helper
	renderItem := func(i int, name, val string) {
		cursor := "  "
		if m.settings.cursor == i {
			cursor = "> "
		}
		
		checked := "[ ]"
		if val == "true" { checked = "[x]" }
		if val == "false" { checked = "[ ]" }
		// Custom values handled by caller
		
		style := lipgloss.NewStyle()
		if m.settings.cursor == i {
			style = style.Foreground(lipgloss.Color("205")).Bold(true)
		}
		
		line := fmt.Sprintf("%s%s %s: %s", cursor, checked, name, val)
		s.WriteString(style.Render(line) + "\n")
	}

	// 0 Metadata
	metaVal := "OFF"
	if m.config.Metadata { metaVal = "ON" }
	renderItem(0, "Preserve Metadata", metaVal)

	// 1 Output
	renderItem(1, "Output Mode", m.config.OutputMode)

	// 2 Mascot
	renderItem(2, "Mascot", string(m.config.Mascot))

	return docStyle.Render(s.String() + "\n(Space/Enter to toggle)")
}
