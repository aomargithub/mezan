package http

import (
	"errors"
	"fmt"
	"github.com/aomargithub/mezan/internal/db"
	"github.com/aomargithub/mezan/internal/domain"
	"net/http"
	"strconv"
	"time"
)

type expenseCreateForm struct {
	MezaniId    int
	Name        string
	TotalAmount float32
	CommonCreateView
}

type expenseView struct {
	Expense domain.Expense
	CommonView
}

func (s Server) getExpenseCreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mezaniId, err := strconv.Atoi(r.PathValue("mezaniId"))
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		expenseCreateForm := expenseCreateForm{
			MezaniId:         mezaniId,
			CommonCreateView: s.commonCreateView(r),
		}
		s.render(w, r, "expenseCreate.tmpl", http.StatusOK, expenseCreateForm)
	})
}

func (s Server) postExpenseCreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mezaniId, err := strconv.Atoi(r.PathValue("mezaniId"))
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		err = r.ParseForm()
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		name := r.PostForm.Get("name")
		totalAmount64, err := strconv.ParseFloat(r.PostForm.Get("totalAmount"), 32)
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		totalAmount := float32(totalAmount64)
		expenseCreateForm := expenseCreateForm{
			MezaniId:         mezaniId,
			Name:             name,
			TotalAmount:      totalAmount,
			CommonCreateView: s.commonCreateView(r),
		}
		expenseCreateForm.NotBlank("Name", name)
		expenseCreateForm.NotNegative("TotalAmount", totalAmount)

		if !expenseCreateForm.Valid() {
			s.render(w, r, "expenseCreate.tmpl", http.StatusBadRequest, expenseCreateForm)
			return
		}

		expense := domain.Expense{
			Mezani:      domain.Mezani{Id: mezaniId},
			Name:        name,
			TotalAmount: totalAmount,
			Creator:     domain.User{Id: expenseCreateForm.Authentication.Id},
			CreatedAt:   time.Now(),
		}

		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.rollback(tx)
		err = s.expenseService.Create(expense)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		err = s.mezaniService.AddExpense(mezaniId, expense.TotalAmount)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		_ = tx.Commit()
		s.sessionManager.Put(r.Context(), "flash", "Expense successfully created!")
		http.Redirect(w, r, fmt.Sprintf("/mezanis/%d", mezaniId), http.StatusSeeOther)
	})
}

func (s Server) getExpenseViewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expenseId, err := strconv.Atoi(r.PathValue("expenseId"))
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
		expense, err := s.expenseService.Get(expenseId)
		if err != nil {
			if errors.Is(err, db.ErrNoRecord) {
				http.Redirect(w, r, "/", http.StatusNotFound)
				return
			}
			s.serverError(w, r, err)
			return
		}
		_ = tx.Commit()
		view := expenseView{
			Expense:    expense,
			CommonView: s.commonView(r),
		}
		s.render(w, r, "expenseView.tmpl", http.StatusOK, view)
	})
}
