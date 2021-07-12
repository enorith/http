package view

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"strings"

	"github.com/enorith/http/content"
)

var DefaultManager *Manager

type Manager struct {
	fileSystem  fs.FS
	ext, perfix string
}

func (m *Manager) Template(name string) (*template.Template, error) {

	sep := string(os.PathSeparator)
	if m.IsEmbed() {
		sep = "/"
	}
	if m.perfix != "" {
		name = m.perfix + "." + name
	}

	file := fmt.Sprintf("%s.%s", strings.ReplaceAll(name, ".", sep), m.ext)

	return template.ParseFS(m.fileSystem, file)
}

func (m *Manager) Get(name string, code int, data interface{}) (*content.TemplateResponse, error) {
	temp, e := m.Template(name)

	return content.TempResponse(temp, code, data), e
}

func (m *Manager) IsEmbed() bool {
	_, ok := m.fileSystem.(embed.FS)

	return ok
}

func NewManager(fs fs.FS, ext string, perfix ...string) *Manager {
	var p string
	if len(perfix) > 0 {
		p = perfix[0]
	}
	return &Manager{
		fileSystem: fs,
		ext:        ext,
		perfix:     p,
	}
}

func WithDefault(fs fs.FS, ext string, perfix ...string) {
	DefaultManager = NewManager(fs, ext, perfix...)
}

func View(name string, code int, data interface{}) (*content.TemplateResponse, error) {
	if DefaultManager == nil {
		return nil, errors.New("uninitialized default manager, call view.WithDefault first")
	}

	return DefaultManager.Get(name, code, data)
}
