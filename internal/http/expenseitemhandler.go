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

type expenseItemCreateForm struct {
	MezaniId    int
	ExpenseId   int
	Name        string
	TotalAmount float32
	Quantity    float32
	Amount      float32
	CommonCreateView
}

func (s Server) getExpenseItemCreateHandler() http.Handler {
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
		defer s.expenseService.Rollback(tx)
		mezaniId, err := s.expenseService.GetMezaniId(expenseId)

		if err != nil {
			if errors.Is(err, db.ErrNoRecord) {
				params := make(map[string]string)
				params["type"] = "Expense"
				params["id"] = strconv.Itoa(expenseId)
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
		expenseItemCreateForm := expenseItemCreateForm{
			ExpenseId:        expenseId,
			MezaniId:         mezaniId,
			CommonCreateView: s.commonCreateView(r),
		}
		_ = tx.Commit()
		s.render(w, r, "expenseItemCreate.tmpl", http.StatusOK, expenseItemCreateForm)
	})
}

func (s Server) postExpenseItemCreateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expenseId, err := strconv.Atoi(r.PathValue("expenseId"))
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

		amount64, err := strconv.ParseFloat(r.PostForm.Get("amount"), 32)
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		amount := float32(amount64)

		totalAmount64, err := strconv.ParseFloat(r.PostForm.Get("totalAmount"), 32)
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		totalAmount := float32(totalAmount64)

		quantity64, err := strconv.ParseFloat(r.PostForm.Get("quantity"), 32)
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		quantity := float32(quantity64)

		mezaniId, err := strconv.Atoi(r.PostForm.Get("mezaniId"))
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		expenseItemCreateForm := expenseItemCreateForm{
			Name:             name,
			ExpenseId:        expenseId,
			MezaniId:         mezaniId,
			Amount:           amount,
			TotalAmount:      totalAmount,
			Quantity:         quantity,
			CommonCreateView: s.commonCreateView(r),
		}
		expenseItemCreateForm.NotBlank("Name", name)
		expenseItemCreateForm.NotNegative("TotalAmount", totalAmount)
		expenseItemCreateForm.NotNegative("amount", amount)
		expenseItemCreateForm.NotNegative("quantity", quantity)

		if !expenseItemCreateForm.Valid() {
			s.render(w, r, "expenseItemCreate.tmpl", http.StatusBadRequest, expenseItemCreateForm)
			return
		}
		expenseItem := domain.ExpenseItem{
			Name:        name,
			Creator:     domain.User{Id: expenseItemCreateForm.Authentication.Id},
			Mezani:      domain.Mezani{Id: mezaniId},
			Amount:      amount,
			TotalAmount: totalAmount,
			CreatedAt:   time.Now(),
			Quantity:    quantity,
			Expense:     domain.Expense{Id: expenseId},
		}
		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.expenseItemService.Rollback(tx)
		err = s.expenseItemService.Create(expenseItem)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		err = tx.Commit()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		s.sessionManager.Put(r.Context(), "flash", "Item successfully created!")
		http.Redirect(w, r, fmt.Sprintf("/expenses/%d", expenseId), http.StatusSeeOther)
	})
}
