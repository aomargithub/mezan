package http

import (
	"errors"
	"github.com/aomargithub/mezan/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type userSignUpForm struct {
	Name  string
	Email string
	CommonCreateView
}

type loginForm struct {
	Email string
	CommonCreateView
}

func (s Server) getUserSignUpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		form := userSignUpForm{
			CommonCreateView: CommonCreateView{
				CommonView: CommonView{
					CsrfToken: s.csrfToken(r),
				},
			},
		}
		s.render(w, r, "userSignUp.tmpl", http.StatusOK, form)
	})
}

func (s Server) getLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		form := loginForm{
			CommonCreateView: CommonCreateView{
				CommonView: CommonView{
					CsrfToken: s.csrfToken(r),
					Flash:     s.sessionManager.PopString(r.Context(), "flash"),
				},
			},
		}
		s.render(w, r, "login.tmpl", http.StatusOK, form)
	})
}

func (s Server) postUserSignUpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
		}

		var (
			name     = r.PostForm.Get("name")
			password = r.PostForm.Get("password")
			email    = r.PostForm.Get("email")
		)
		form := userSignUpForm{
			Name:  name,
			Email: email,
			CommonCreateView: CommonCreateView{
				CommonView: CommonView{
					CsrfToken: s.csrfToken(r),
				},
			},
		}

		form.NotBlank("Name", name)
		form.NotBlank("Email", email)
		form.NotBlank("Password", password)
		form.ValidEmail("Email", email)
		form.MinChars("Password", password, 8)

		if !form.Valid() {
			s.render(w, r, "userSignUp.tmpl", http.StatusUnprocessableEntity, form)
			return
		}
		user := domain.User{
			Name:      name,
			Email:     email,
			CreatedAt: time.Now(),
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			s.serverError(w, r, err)
		}

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		s.userService.Rollback(tx)
		err = s.userService.Create(user, string(hashedPassword))
		if err != nil {
			if errors.Is(err, domain.ErrDuplicateEmail) {
				form.AddFieldError("Email", "Email address is already in use")
				s.render(w, r, "userSignUp.tmpl", http.StatusUnprocessableEntity, form)
				return
			}
			s.serverError(w, r, err)
			return
		}
		_ = tx.Commit()
		s.render(w, r, "userSignUp.tmpl", http.StatusOK, nil)
	})
}

func (s Server) postLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
		}

		var (
			password = r.PostForm.Get("password")
			email    = r.PostForm.Get("email")
		)

		form := loginForm{
			Email: email,
			CommonCreateView: CommonCreateView{
				CommonView: CommonView{
					CsrfToken: s.csrfToken(r),
				},
			},
		}
		form.NotBlank("Email", email)
		form.NotBlank("Password", password)

		if !form.Valid() {
			s.render(w, r, "login.tmpl", http.StatusUnprocessableEntity, form)
			return
		}

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.userService.Rollback(tx)
		user, hp, err := s.userService.GetInfoAndHashedPassword(email)
		if err != nil {
			if errors.Is(err, domain.ErrNoRecord) {
				form.AddFormError("Email or Password is incorrect")
				s.render(w, r, "login.tmpl", http.StatusUnprocessableEntity, form)
				return
			}
			s.serverError(w, r, err)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(hp), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				form.AddFormError("Email or Password is incorrect")
				s.render(w, r, "login.tmpl", http.StatusUnprocessableEntity, form)
				return
			}
			s.serverError(w, r, err)
			return
		}
		if err != nil {
			s.serverError(w, r, err)
			return
		}

		err = s.sessionManager.RenewToken(r.Context())
		if err != nil {
			s.serverError(w, r, err)
			return
		}

		s.sessionManager.Put(r.Context(), authenticatedUserIdSessionKey, user.Id)
		s.sessionManager.Put(r.Context(), authenticatedUserNameSessionKey, user.Name)
		_ = tx.Commit()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func (s Server) postLogoutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := s.sessionManager.RenewToken(r.Context())
		if err != nil {
			s.serverError(w, r, err)
			return
		}

		s.sessionManager.Remove(r.Context(), authenticatedUserIdSessionKey)
		s.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}
