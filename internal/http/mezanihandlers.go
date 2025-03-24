package http

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"unsafe"

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

func (s Server) mezaniCreateHandler() http.Handler {
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
		userId, userName := s.getCurrentUserInfo(r)
		mezani := domain.Mezani{
			Name: name,
			Creator: domain.User{
				Id: userId,
			},
			ShareId:   fmt.Sprintf("%s_%s_%s", userName, name, randStringBytesMaskImprSrcUnsafe(10)),
			CreatedAt: now,
		}

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.mezaniService.Rollback(tx)
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
		defer s.mezaniService.Rollback(tx)
		mezani, err := s.mezaniService.Get(mezaniId)
		if err != nil {
			if errors.Is(err, domain.ErrNoRecord) {
				params := make(map[string]string)
				params["type"] = "Mezani"
				params["id"] = strconv.Itoa(mezaniId)
				ev := errorView{
					Data:       params,
					CommonView: s.commonView(r),
				}
				s.clientError(w, http.StatusNotFound, ev)
				return
			}
			s.serverError(w, r, err)
			return
		}
		userId, _ := s.getCurrentUserInfo(r)
		accessible, err := s.membershipService.MezaniAccessibleBy(mezaniId, userId)
		if !accessible {
			params := make(map[string]string)
			params["type"] = "Mezani"
			params["id"] = strconv.Itoa(mezaniId)
			ev := errorView{
				Data:       params,
				CommonView: s.commonView(r),
			}
			s.clientError(w, http.StatusForbidden, ev)
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

func (s Server) getMezaniViewByShareIdHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shareId := r.PathValue("shareId")

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.mezaniService.Rollback(tx)
		mezani, err := s.mezaniService.GetByShareId(shareId)
		if err != nil {
			if errors.Is(err, domain.ErrNoRecord) {
				params := make(map[string]string)
				params["type"] = "Mezani"
				params["id"] = shareId
				ev := errorView{
					Data:       params,
					CommonView: s.commonView(r),
				}
				s.clientError(w, http.StatusNotFound, ev)
				return
			}
			s.serverError(w, r, err)
			return
		}
		userId, _ := s.getCurrentUserInfo(r)
		membership := domain.MemberShip{
			Mezani: domain.Mezani{
				Id: mezani.Id,
			},
			Member: domain.User{
				Id: userId,
			},
			CreatedAt: time.Now(),
		}
		err = s.membershipService.Create(membership)

		if err != nil {
			if errors.Is(err, domain.ErrDuplicateRecord) {
				s.sessionManager.Put(r.Context(), "flash", "Your are already a member in that Mezani!")
				http.Redirect(w, r, fmt.Sprintf("/mezanis/%d", mezani.Id), http.StatusSeeOther)
				return
			}
			s.serverError(w, r, err)
			return
		}

		_ = tx.Commit()
		s.sessionManager.Put(r.Context(), "flash", "Your have been added as a member to that Mezani successfully!")
		http.Redirect(w, r, fmt.Sprintf("/mezanis/%d", mezani.Id), http.StatusSeeOther)
	})
}

func (s Server) homeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.mezaniService.Rollback(tx)
		userId, _ := s.getCurrentUserInfo(r)
		mezanis, err := s.mezaniService.GetAll(userId)
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

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
