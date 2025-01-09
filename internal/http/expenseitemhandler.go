package http

import (
	"errors"
	"fmt"
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

type expenseItemView struct {
	ExpenseItem domain.ExpenseItem
	CommonView
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
			if errors.Is(err, domain.ErrNoRecord) {
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

		userId, _ := s.getCurrentUserInfo(r)
		accessible, err := s.membershipService.ExpenseAccessibleBy(expenseId, userId)
		if !accessible {
			params := make(map[string]string)
			params["type"] = "Expense"
			params["id"] = strconv.Itoa(expenseId)
			ev := errorView{
				Data:       params,
				CommonView: s.commonView(r),
			}
			s.clientError(w, http.StatusForbidden, ev)
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
		tx, err := s.DB.Begin()
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		defer s.expenseItemService.Rollback(tx)

		exists, err := s.expenseService.IsExist(expenseId)
		if err != nil {
			s.serverError(w, r, err)
			return
		}

		if !exists {
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

		userId, _ := s.getCurrentUserInfo(r)
		accessible, err := s.membershipService.ExpenseAccessibleBy(expenseId, userId)
		if !accessible {
			params := make(map[string]string)
			params["type"] = "Expense"
			params["id"] = strconv.Itoa(expenseId)
			ev := errorView{
				Data:       params,
				CommonView: s.commonView(r),
			}
			s.clientError(w, http.StatusForbidden, ev)
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
		err = s.expenseItemService.Create(expenseItem)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		err = s.expenseService.AddAmount(expenseId, totalAmount)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		err = s.mezaniService.AddAmount(mezaniId, totalAmount)
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

func (s Server) getExpenseItemViewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expenseItemId, err := strconv.Atoi(r.PathValue("expenseItemId"))
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
		expenseItem, err := s.expenseItemService.Get(expenseItemId)
		if err != nil {
			if errors.Is(err, domain.ErrNoRecord) {
				params := make(map[string]string)
				params["type"] = "Expense Item"
				params["id"] = strconv.Itoa(expenseItemId)
				ev := errorView{
					Data:       params,
					CommonView: s.commonView(r),
				}
				s.clientError(w, http.StatusNotFound, ev)
			}
			s.serverError(w, r, err)
			return
		}
		userId, _ := s.getCurrentUserInfo(r)
		accessible, err := s.membershipService.ExpenseItemAccessibleBy(expenseItemId, userId)
		if !accessible {
			params := make(map[string]string)
			params["type"] = "Expense Item"
			params["id"] = strconv.Itoa(expenseItemId)
			ev := errorView{
				Data:       params,
				CommonView: s.commonView(r),
			}
			s.clientError(w, http.StatusForbidden, ev)
			return
		}
		_ = tx.Commit()
		view := expenseItemView{
			ExpenseItem: expenseItem,
			CommonView:  s.commonView(r),
		}
		s.render(w, r, "expenseItemView.tmpl", http.StatusOK, view)
	})
}
