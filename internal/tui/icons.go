package tui

import (
	"path/filepath"
	"strings"
)

// GetIcon returns an emoji icon based on file type/extension
func getIcon(name string, isDir bool) string {
	if isDir {
		return "📁"
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	// Images
	case ".png", ".jpg", ".jpeg":
		return "🖼️ "
	case ".webp", ".gif", ".bmp", ".tiff":
		return "🎨"
	case ".svg":
		return "✒️ "
	
	// Archives
	case ".zip", ".tar", ".gz", ".rar", ".7z":
		return "📦"
	
	// Code
	case ".go":
		return "🐹"
	case ".js", ".ts", ".jsx", ".tsx":
		return "📜"
	case ".py":
		return "🐍"
	case ".rs":
		return "🦀"
	case ".c", ".cpp", ".h":
		return "Ⓜ️ "
	case ".html", ".css":
		return "🌐"
	case ".json", ".yaml", ".yml", ".toml", ".xml":
		return "⚙️ "
	case ".md", ".txt":
		return "📝"
	case ".sh", ".bash", ".zsh":
		return "💻"
	
	// Media
	case ".mp3", ".wav", ".flac":
		return "🎵"
	case ".mp4", ".mkv", ".mov", ".avi":
		return "🎬"
		
	// Documents
	case ".pdf":
		return "📕"
	case ".doc", ".docx":
		return "📘"
	case ".xls", ".xlsx":
		return "📗"
	
	default:
		return "📄"
	}
}
