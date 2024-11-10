package http

import (
	"net/http"
	"html/template"
	"github.com/aomargithub/mezan/internal/domain"
	"github.com/google/uuid"
)


func (s Server) createMezanHandler() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		files := []string {
			"./ui/html/base.tmpl",
			"./ui/html/partials/nav.tmpl",
			"./ui/html/pages/home.tmpl",
		}
		s.mezanaRepo.Insert(domain.Mezana{
			Id: uuid.New(),
			Name: "hi",
		})
		ts, err := template.ParseFiles(files...)
		if (err != nil) {
			s.ServerError(w, r, err)
			return
		} 

		err = ts.ExecuteTemplate(w, "base", nil)

		if(err != nil) {
			s.ServerError(w, r, err)
		}
	}
}