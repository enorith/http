package view

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"regexp"
	"strings"

	"github.com/enorith/http/content"
)

var DefaultManager *Manager

type Manager struct {
	fileSystem  fs.FS
	ext, perfix string
}

func (m *Manager) Template(name string) (*template.Template, error) {
	temp := template.New(name)

	b, e := m.Parse(name)

	if e != nil {
		return nil, e
	}

	return temp.Parse(string(b))
}

func (m *Manager) viewFilePath(name string) string {

	sep := "/"

	if m.perfix != "" {
		name = m.perfix + "." + name
	}

	return fmt.Sprintf("%s.%s", strings.ReplaceAll(name, ".", sep), m.ext)
}

func (m *Manager) Parse(name string) ([]byte, error) {
	tokenExp := regexp.MustCompile("(@.*)")
	file := m.viewFilePath(name)
	b, e := fs.ReadFile(m.fileSystem, file)

	if e != nil {
		return nil, e
	}
	ts := tokenExp.FindAllSubmatch(b, -1)
	for _, v := range ts {
		line := v[1]
		tokens := bytes.Split(line, []byte(" "))
		parser := tokens[0]
		if bytes.Equal([]byte("@import"), parser) || bytes.Equal([]byte("@use"), parser) {
			imp := bytes.TrimSpace(tokens[1])
			bi, e := m.Parse(string(imp))
			if e != nil {
				return nil, e
			}
			b = bytes.ReplaceAll(b, v[1], bi)
		}
	}

	return b, e
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
