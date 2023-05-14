package render

import (
	"net/http"
	"text/template"

	"github.com/gaolaoge/ms-go/internal/bytesconv"
)

type HTML struct {
	Data       any
	Name       string
	Template   *template.Template
	IsTemplate bool
}

func (h HTML) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)
	if h.IsTemplate {
		return h.Template.ExecuteTemplate(w, h.Name, h.Data)
	}
	_, err := w.Write(bytesconv.StringToBytes(h.Data.(string)))
	return err
}

func (h HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}

type HTMLRender struct {
	Template *template.Template
}
