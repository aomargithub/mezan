package http

import (
	"errors"
	"fmt"
	"github.com/aomargithub/mezan/internal/db"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/aomargithub/mezan/internal/domain"
)

type mezaniCreateForm struct {
	Name string
	CommonCreateView
}

type mezaniView struct {
	Mezani domain.Mezani
	CommonView
}

type homeView struct {
	Mezanis []domain.Mezani
	CommonView
}

func (s Server) getMezaniCreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mezaniCreateForm := mezaniCreateForm{
			CommonCreateView: s.commonCreateView(r),
		}
		s.render(w, r, "mezaniCreate.tmpl", http.StatusOK, mezaniCreateForm)
	})
}

func (s Server) postMezaniCreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		err := r.ParseForm()
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		name := r.PostForm.Get("name")

		mezaniCreateForm := mezaniCreateForm{
			Name:             name,
			CommonCreateView: s.commonCreateView(r),
		}

		mezaniCreateForm.NotBlank("Name", name)
		if !mezaniCreateForm.Valid() {
			s.render(w, r, "mezaniCreate.tmpl", http.StatusBadRequest, mezaniCreateForm)
			return
		}
		userId := s.sessionManager.GetInt(r.Context(), authenticatedUserIdSessionKey)
		userName := s.sessionManager.GetString(r.Context(), authenticatedUserNameSessionKey)
		mezani := domain.Mezani{
			Name: name,
			Creator: domain.User{
				Id: userId,
			},
			ShareId:   fmt.Sprintf("%s_%s_%d", userName, name, rand.Intn(1000000)+1000000),
			CreatedAt: now,
		}

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.rollback(tx)
		mezaniId, err := s.mezaniService.Create(mezani)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		memberShip := domain.MemberShip{
			Mezani: domain.Mezani{
				Id: mezaniId,
			},
			Member: domain.User{
				Id: userId,
			},
			CreatedAt: now,
		}
		err = s.membershipService.Create(memberShip)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		_ = tx.Commit()
		s.sessionManager.Put(r.Context(), "flash", "Mezania successfully created!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func (s Server) getMezaniViewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mezaniId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.rollback(tx)
		mezani, err := s.mezaniService.Get(mezaniId)
		if err != nil {
			if errors.Is(err, db.ErrNoRecord) {
				http.Redirect(w, r, "/", http.StatusNotFound)
				return
			}
			s.serverError(w, r, err)
			return
		}
		_ = tx.Commit()
		view := mezaniView{
			Mezani:     mezani,
			CommonView: s.commonView(r),
		}
		s.render(w, r, "mezaniView.tmpl", http.StatusOK, view)
	})
}

func (s Server) homeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.rollback(tx)
		mezanis, err := s.mezaniService.GetAll()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		_ = tx.Commit()
		view := homeView{
			Mezanis:    mezanis,
			CommonView: s.commonView(r),
		}
		s.render(w, r, "home.tmpl", http.StatusOK, view)
	})
}
