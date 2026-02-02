package tui

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// generatePreview returns a string representation of the file content/metadata
func generatePreview(path string, width, height int) string {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if info.IsDir() {
		return fmt.Sprintf("\n  📂 Directory: %s\n  Mod: %s", filepath.Base(path), info.ModTime().Format("2006-01-02 15:04"))
	}

	ext := strings.ToLower(filepath.Ext(path))
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n  📄 %s\n", filepath.Base(path)))
	sb.WriteString(fmt.Sprintf("  Size: %s\n", formatBytes(info.Size())))
	sb.WriteString(fmt.Sprintf("  Mod: %s\n", info.ModTime().Format("2006-01-02 15:04")))

	// Image Handling
	if isImage(ext) {
		file, err := os.Open(path)
		if err == nil {
			defer file.Close()
			cfg, format, err := image.DecodeConfig(file)
			if err == nil {
				sb.WriteString(fmt.Sprintf("  Type: %s Image\n", strings.ToUpper(format)))
				sb.WriteString(fmt.Sprintf("  Dimensions: %dx%d\n", cfg.Width, cfg.Height))
				
				// ASCII Art Placeholder (Real implementation requires scaling)
				// For now, let's just show a nice box with info
				// Implementing a full scaler here might be too much for MVP without external libs,
				// but let's try a very simple sampler if file size is small enough to read quickly.
				
				sb.WriteString("\n")
				sb.WriteString(generateSimpleAscii(path, 40, 20)) // Fixed small size for preview
			}
		}
	}

	return sb.String()
}

func isImage(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		return true
	}
	return false
}

// generateSimpleAscii generates a very basic block preview
// This is a naive implementation; for better quality we'd need a proper resizer.
func generateSimpleAscii(path string, w, h int) string {
    // For now, return a placeholder to ensure UI stability first.
    // Real ASCII art in pure Go without deps that looks good is complex.
    // We will use a stylistic placeholder.
    
    style := lipgloss.NewStyle().
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("63")).
        Padding(1).
        Align(lipgloss.Center).
        Width(30).
        Height(10)
        
    return style.Render("🖼️  Image Preview\n(Visuals coming soon)")
}
