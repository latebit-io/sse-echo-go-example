package views

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type HTMLRenderer struct {
	templates *template.Template
}

func NewHTMLRenderer(dir string) *HTMLRenderer {
	return &HTMLRenderer{
		templates: template.Must(template.ParseGlob(dir)),
	}
}

// Render renders a template to the response writer
func (t *HTMLRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}