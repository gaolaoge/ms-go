package render

import (
	"errors"
	"fmt"
	"net/http"
)

type Redirect struct {
	Location string
	Request  *http.Request
	Code     int
}

func (r Redirect) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	if (r.Code < http.StatusMultipleChoices ||
		r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
		return errors.New(fmt.Sprintf("Cannot redirect with status code %d", r.Code))
	}
	http.Redirect(w, r.Request, r.Location, http.StatusFound)
	return nil
}

func (r Redirect) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "application/xml; charset=utf-8")
}
