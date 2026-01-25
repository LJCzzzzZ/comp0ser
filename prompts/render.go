package prompts

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
)

//go:embed *.tmpl
var promptFS embed.FS

type Config struct {
	Subject  string
	Segments int
	MinChars int
	MaxChars int

	// focus for generated narrations
	Focus string

	Hook string
}

type Renderer struct {
	sys *template.Template
}

func NewRenderer() (*Renderer, error) {
	t, err := template.ParseFS(promptFS, "sleep_system.tmpl")
	if err != nil {
		return nil, err
	}

	return &Renderer{sys: t}, nil
}

func (r *Renderer) System(conf Config) (string, error) {
	var buf bytes.Buffer
	if err := r.sys.Execute(&buf, conf); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
