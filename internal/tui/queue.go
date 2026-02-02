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
		{Title: "File", Width: 40},
		{Title: "Status", Width: 15},
		{Title: "Size", Width: 15},
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
			m.pipeline.Start() // Or Resume/Kickoff
			return m, nil // or some cmd
		case "d":
			// Delete selected
			// We need to know which one is selected in table
			// bubbles/table selected index matches Rows index
			// Our Rows match p.Jobs() order
			if len(m.queue.table.Rows()) > 0 {
				idx := m.queue.table.Cursor()
				jobs := m.pipeline.Jobs() // Valid copy?
				// Race condition if pipeline removes jobs in background?
				// Pipeline only removes if WE tell it to, or finish?
				// It doesn't auto-remove.
				if idx >= 0 && idx < len(jobs) {
					job := jobs[idx]
					m.pipeline.RemoveJob(job.FilePath)
				}
			}
		case "c":
			m.pipeline.ClearCompleted()
		}
	
	// We need to subscribe to pipeline updates mainly in Progress view, 
	// but Queue view might show pending.
	// For "Queue", we assume it's the list of files to be processed.
	}
	
	// Sync table with pipeline jobs
	jobs := m.pipeline.Jobs() // Thread safe copy
	rows := make([]table.Row, len(jobs))
	for i, j := range jobs {
		rows[i] = table.Row{
			filepath.Base(j.FilePath),
			string(j.Status),
			fmt.Sprintf("%d", j.OriginalSize),
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
