package tui

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type queueModel struct {
	table table.Model
}

func newQueueModel() queueModel {
	columns := []table.Column{
		{Title: "File", Width: 30},
		{Title: "Status", Width: 12},
		{Title: "Size", Width: 12},
		{Title: "After", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
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

	return queueModel{table: t}
}

func (m MainModel) updateQueue(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Run compression
			m.state = StateCompress
			m.pipeline.Start() 
			return m, nil 
		case "d":
			if len(m.queue.table.Rows()) > 0 {
				idx := m.queue.table.Cursor()
				jobs := m.pipeline.Jobs() 
				if idx >= 0 && idx < len(jobs) {
					job := jobs[idx]
					m.pipeline.RemoveJob(job.FilePath)
				}
			}
		case "c":
			m.pipeline.ClearCompleted()
		}
	
	}
	
	// Sync table with pipeline jobs
	jobs := m.pipeline.Jobs() // Thread safe copy
	rows := make([]table.Row, len(jobs))
	for i, j := range jobs {
		after := "-"
		if j.CompressedSize > 0 {
		    after = formatBytes(j.CompressedSize)
		}
		
		rows[i] = table.Row{
			filepath.Base(j.FilePath),
			string(j.Status),
			formatBytes(j.OriginalSize),
			after,
		}
	}
	m.queue.table.SetRows(rows)

	m.queue.table, cmd = m.queue.table.Update(msg)
	return m, cmd
}

func (m MainModel) viewQueue() string {
	return docStyle.Render(
		"Queue (" + fmt.Sprintf("%d", len(m.pipeline.Jobs())) + " files)\n" +
		"Press 'r' to start compression.\n\n" +
		m.queue.table.View(),
	)
}
