package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aomargithub/mezan/internal/domain"
)

type mezaniCreateForm struct {
	Name string
	Validator
	*Authentication
	CsrfToken string
}

type mezaniView struct {
	Mezani domain.Mezani
	*Authentication
	CsrfToken string
}

type homeView struct {
	Mezanis []domain.Mezani
	*Authentication
	CsrfToken string
}

func (s Server) getMezaniCreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mezaniCreateForm := mezaniCreateForm{
			Authentication: s.createAuthentication(r),
			CsrfToken:      s.csrfToken(r),
		}
		s.render(w, r, "mezaniCreate.tmpl", http.StatusOK, mezaniCreateForm)
	})
}

func (s Server) postMezaniCreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		name := r.PostForm.Get("name")

		mezaniCreateForm := mezaniCreateForm{
			Name:           name,
			Authentication: s.createAuthentication(r),
			CsrfToken:      s.csrfToken(r),
		}

		mezaniCreateForm.NotBlank("userName", name)
		if !mezaniCreateForm.Valid() {
			s.render(w, r, "mezaniCreate.tmpl", http.StatusOK, mezaniCreateForm)
			return
		}
		userId := s.sessionManager.GetInt(r.Context(), authenticatedUserIdSessionKey)
		mezani := domain.Mezani{
			Name:      name,
			CreatorId: userId,
			CreatedAt: time.Now(),
		}

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			defer tx.Rollback()
			return
		}
		err = s.mezaniService.Create(mezani)
		defer tx.Rollback()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		s.sessionManager.Put(r.Context(), "flash", "Mezania successfully created!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func (s Server) getMezaniHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mezaniId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		tx, err := s.mezaniService.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		mezani, err := s.mezaniService.Get(mezaniId)
		defer tx.Rollback()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		view := mezaniView{
			Mezani:         mezani,
			Authentication: s.createAuthentication(r),
			CsrfToken:      s.csrfToken(r),
		}
		s.render(w, r, "mezaniView.tmpl", http.StatusOK, view)
	})
}

func (s Server) homeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx, err := s.mezaniService.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		mezanis, err := s.mezaniService.GetAll()
		defer tx.Rollback()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		view := homeView{
			Mezanis:        mezanis,
			Authentication: s.createAuthentication(r),
			CsrfToken:      s.csrfToken(r),
		}
		s.render(w, r, "home.tmpl", http.StatusOK, view)
	})
}
