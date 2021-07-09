package view

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"strings"
)

type Manager struct {
	fileSystem fs.FS
	ext        string
}

func (m *Manager) Get(name string) (*template.Template, error) {

	sep := string(os.PathSeparator)
	if m.IsEmbed() {
		sep = "/"
	}

	file := fmt.Sprintf("%s.%s", strings.ReplaceAll(name, ".", sep), m.ext)

	return template.ParseFS(m.fileSystem, file)
}

func (m *Manager) IsEmbed() bool {
	_, ok := m.fileSystem.(embed.FS)

	return ok
}

func NewManager(fs fs.FS, ext string) *Manager {
	return &Manager{
		fileSystem: fs,
		ext:        ext,
	}
}
