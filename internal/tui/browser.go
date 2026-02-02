package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gmsakibursabbir/tinitui/internal/scanner"
)

type browserModel struct {
	currentDir string
	dirs       list.Model
	files      list.Model
	activePane int // 0 = Dirs, 1 = Files
	selected   map[string]bool
	recursive  bool
	err        error
}

// Item types for lists
type dirItem struct {
	name string
	path string
}
func (d dirItem) Title() string       { return d.name } // Folder icon?
func (d dirItem) Description() string { return "" }
func (d dirItem) FilterValue() string { return d.name }

type fileItem struct {
	name string
	path string
	size int64
}
func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return fmt.Sprintf("%d bytes", f.size) } // Format nice later
func (f fileItem) FilterValue() string { return f.name }

func newBrowserModel() browserModel {
	cwd, _ := os.Getwd()
	
	// Init Lists
	dList := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	dList.Title = "Directories"
	dList.SetShowHelp(false)
	
	fList := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	fList.Title = "Files"
	fList.SetShowHelp(false)

	m := browserModel{
		currentDir: cwd,
		dirs:       dList,
		files:      fList,
		activePane: 0,
		selected:   make(map[string]bool),
	}
	m.refresh()
	return m
}

func (b *browserModel) refresh() {
	entries, err := os.ReadDir(b.currentDir)
	if err != nil {
		b.err = err
		return
	}

	var dirs []list.Item
	var files []list.Item

	// Add ".." if not root
	if filepath.Dir(b.currentDir) != b.currentDir {
		dirs = append(dirs, dirItem{name: "..", path: filepath.Dir(b.currentDir)})
	}

	for _, e := range entries {
		if e.IsDir() {
			// Skip hidden?
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			dirs = append(dirs, dirItem{name: e.Name() + "/", path: filepath.Join(b.currentDir, e.Name())})
		} else {
			// Check support
			// We handle extension check manually or use scanner helper
			// Scanner implementation used internal map, maybe expose it?
			// Or just duplicate logic: png/jpg/webp
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
				info, _ := e.Info()
				files = append(files, fileItem{name: e.Name(), path: filepath.Join(b.currentDir, e.Name()), size: info.Size()})
			}
		}
	}
	
	b.dirs.SetItems(dirs)
	b.files.SetItems(files)
}

func (m MainModel) updateBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Ensure initialized (simple check, or do in Init)
	if m.browser.currentDir == "" {
		m.browser = newBrowserModel()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.browser.activePane = (m.browser.activePane + 1) % 2
		case "enter":
			if m.browser.activePane == 0 {
				// Change directory
				i := m.browser.dirs.SelectedItem()
				if i != nil {
					d := i.(dirItem)
					m.browser.currentDir = d.path
					m.browser.refresh()
					// Reset cursor?
					m.browser.dirs.ResetSelected()
				}
			}
		case "x":
			// Toggle recursive scan for "Add"?
			// Or toggle view mode?
			// Requirement: "Options: [x] recursive".
			// This likely affects what happens when we "Add" a folder?
			// The File Picker desc says: "Done -> add to queue".
			// If we select a FOLDER in left pane, do we add it?
			// My browser implementation currently adds items from "selected" map.
			// Currently I only allow selecting FILES in right pane.
			// "Left: directories ... Right: image files".
			// "Space select/unselect".
			// "Options: [x] recursive"
			// "Done -> add to queue"
			
			// Interpretation:
			// If I select a directory in Left pane, and press Space, do I select it?
			// If so, recursive flag applies to that directory addition.
			
			// Current implementation:
			// Space in activePane==1 (files) toggles selection.
			// Space in activePane==0 (dirs) ?? 
			
			// Let's implement Space in Dirs pane to select the dir.
			// And 'x' to toggle recursive flag in browser model.
			
			if m.browser.activePane == 0 {
				// Support selecting directories?
				// m.browser.selected is map[string]bool.
				// If I add a dir path there, Pipeline.AddFiles needs to handle it.
				// Scanner.Scan(paths, recursive) supports it.
				// So if I pass directory paths to pipeline, and pipeline uses scanner...
				// Wait, Pipeline.AddFiles uses os.Stat -> if file, add.
				// Pipeline.AddFiles logic:
				// "info, err := os.Stat(path) ... if err == nil { size = info.Size() } ... p.jobs = append"
				// Pipeline doesn't call Scanner.
				// Scanner is used in CLI before passing to Pipeline.
				// In TUI, `m.pipeline.AddFiles(paths)` assumes paths are identifiable jobs?
				// Pipeline needs to Scan if we pass directories?
				// Or TUI should Scan before passing to Pipeline.
				
				// Fix: TUI "A" handler should Scan selected paths.
				m.browser.recursive = !m.browser.recursive
			}
		case "backspace":
			// Go up
			parent := filepath.Dir(m.browser.currentDir)
			if parent != m.browser.currentDir {
				m.browser.currentDir = parent
				m.browser.refresh()
			}
		case " ":
			if m.browser.activePane == 1 {
				// Toggle file
				i := m.browser.files.SelectedItem()
				if i != nil {
					f := i.(fileItem)
					if m.browser.selected[f.path] {
						delete(m.browser.selected, f.path)
					} else {
						m.browser.selected[f.path] = true
					}
				}
			} else {
				// Toggle Dir?
				i := m.browser.dirs.SelectedItem()
				if i != nil {
					d := i.(dirItem)
					if m.browser.selected[d.path] {
						delete(m.browser.selected, d.path)
					} else {
						m.browser.selected[d.path] = true
					}
				}
			}
		case "ctrl+a":
			// Select all in current view
			if m.browser.activePane == 1 {
				for _, it := range m.browser.files.Items() {
					f := it.(fileItem)
					m.browser.selected[f.path] = true
				}
			}
		case "a":
			// Done -> Add to queue
			// Scan selected
			var paths []string
			for p := range m.browser.selected {
				paths = append(paths, p)
			}
			if len(paths) > 0 {
				// Scan them!
				res, _ := scanner.Scan(paths, m.browser.recursive)
				if len(res.Images) > 0 {
					m.pipeline.AddFiles(res.Images)
					m.state = StateQueue
					m.browser.selected = make(map[string]bool)
				}
			}
		}
	case tea.WindowSizeMsg:
		// Resize lists
		halfWidth := m.width / 2
		m.browser.dirs.SetWidth(halfWidth - 2)
		m.browser.files.SetWidth(halfWidth - 2)
		m.browser.dirs.SetHeight(m.height - 4)
		m.browser.files.SetHeight(m.height - 4)
	}

	// Update active list
	if m.browser.activePane == 0 {
		m.browser.dirs, cmd = m.browser.dirs.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.browser.files, cmd = m.browser.files.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) viewBrowser() string {
	// Simple split view
	// Highlight active pane border
	
	leftStyle := docStyle.Copy().Width(m.width/2 - 4)
	rightStyle := docStyle.Copy().Width(m.width/2 - 4)
	
	if m.browser.activePane == 0 {
		leftStyle = leftStyle.Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("62"))
		rightStyle = rightStyle.Border(lipgloss.NormalBorder())
	} else {
		leftStyle = leftStyle.Border(lipgloss.NormalBorder())
		rightStyle = rightStyle.Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("62"))
	}
	
	// Custom delegate to show selection for files
	// For now just standard list
	
	// Hack to show selection status in title or description?
	// bubbles/list doesn't support dynamic item update easily without SetItems again.
	// But we render the list.
	// We might need a custom item delegate to render the [x].
	
	// Recursive status
	recStatus := "[ ] Recursive (x)"
	if m.browser.recursive {
		recStatus = "[x] Recursive (x)"
	}
	
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top,
			leftStyle.Render(m.browser.dirs.View()),
			rightStyle.Render(m.browser.files.View()),
		),
		docStyle.Render(recStatus),
	)
}
