package http

import (
	"net/http"
	"time"

	"github.com/aomargithub/mezan/internal/domain"
	"github.com/google/uuid"
)


func (s Server) createGetMezaniCreateHandler() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		err := s.templates["mezaniCreate.tmpl"].ExecuteTemplate(w, "base", nil)
		if (err != nil) {
			s.ServerError(w, r, err)
		}
	}
}

func (s Server) createMezaniCreateHandler() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			s.ClientError(w, http.StatusBadRequest)
		}

		name := r.PostForm.Get("name")

		mezani := domain.Mezani {
			Name: name,
			Id: uuid.New(),
			CreatedAt: time.Now(),
		}

		tx, err := s.Db.Begin()
		if err != nil {
			s.ServerError(w, r, err)
		}
		err = s.mezaniService.Create(mezani)
		defer tx.Rollback()
		if err != nil {
			s.ServerError(w, r, err)
		}

		http.Redirect(w,r,"/", http.StatusSeeOther)
	}
}


func (s Server) createGetMezaniHandler() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		mezaniId := uuid.MustParse(r.PathValue("id"))
		
		tx, err := s.mezaniService.DB.Begin()
		if err != nil {
			s.ServerError(w, r, err)
		}
		mezani,err := s.mezaniService.Get(mezaniId)
		defer tx.Rollback()
		if err != nil {
			s.ServerError(w, r, err)
		}
		

		err = s.templates["mezaniView.tmpl"].ExecuteTemplate(w, "base", mezani)

		if(err != nil) {
			s.ServerError(w, r, err)
		}
	}
}

func (s Server) createHomeHandler() http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		
		
		tx, err := s.mezaniService.DB.Begin()
		if err != nil {
			s.ServerError(w, r, err)
		}
		mezanis, err := s.mezaniService.GetAll()
		defer tx.Rollback()
		if err != nil {
			s.ServerError(w, r, err)
		}

		err = s.templates["home.tmpl"].ExecuteTemplate(w, "base", mezanis)

		if(err != nil) {
			s.ServerError(w, r, err)
		}
	}
}