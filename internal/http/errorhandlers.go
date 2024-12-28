package http

import (
	"net/http"
	"runtime/debug"
)

type errorView struct {
	Data map[string]string
	CommonView
}

func (s Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	trace := string(debug.Stack())
	s.Logger.Error(err.Error(), "uri", r.URL.RequestURI(), "method", r.Method, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (s Server) clientError(w http.ResponseWriter, status int, errorData ...errorView) {

	switch status {
	case http.StatusForbidden:
		s.render(w, nil, "forbidden.tmpl", status, errorData[0])
	case http.StatusNotFound:
		s.render(w, nil, "notFound.tmpl", status, errorData[0])
	default:
		http.Error(w, http.StatusText(status), status)
	}
}
