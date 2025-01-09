package http

import (
	"errors"
	"github.com/aomargithub/mezan/internal/domain"
	"net/http"
	"strconv"
)

type expenseItemParticipationCreateForm struct {
	MezaniId      int
	ExpenseId     int
	ExpenseItemId int
	ShareType     domain.ShareType
	Share         float32
	Amount        float32
	ShareTypes    []domain.ShareType
	CommonCreateView
}

func (s Server) getExpenseItemParticipationCreateHandler() http.Handler {
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
		defer s.expenseItemService.Rollback(tx)
		mezaniId, expenseId, err := s.expenseItemService.GetExpenseId(expenseItemId)
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
				return
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
		expenseItemParticipationCreateForm := expenseItemParticipationCreateForm{
			ExpenseId:        expenseId,
			MezaniId:         mezaniId,
			ExpenseItemId:    expenseItemId,
			CommonCreateView: s.commonCreateView(r),
			ShareTypes:       domain.ShareTypes,
		}
		_ = tx.Commit()
		s.render(w, r, "expenseItemParticipationCreate.tmpl", http.StatusOK, expenseItemParticipationCreateForm)
	})
}

func (s Server) postExpenseItemParticipationCreateHandler() http.Handler {
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
		defer s.expenseItemService.Rollback(tx)

		exists, err := s.expenseItemService.IsExist(expenseItemId)

		if err != nil {
			s.serverError(w, r, err)
			return
		}

		if exists {
			params := make(map[string]string)
			params["type"] = "Expense Item"
			params["id"] = strconv.Itoa(expenseItemId)
			ev := errorView{
				Data:       params,
				CommonView: s.commonView(r),
			}
			s.clientError(w, http.StatusNotFound, ev)
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

		err = r.ParseForm()
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		shareType := domain.ShareType(r.PostForm.Get("shareType"))
		share64, err := strconv.ParseFloat(r.PostForm.Get("share"), 32)
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		share := float32(share64)
		amount64, err := strconv.ParseFloat(r.PostForm.Get("amount"), 32)
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		amount := float32(amount64)

		mezaniId, err := strconv.Atoi(r.PostForm.Get("mezaniId"))
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		expenseId, err := strconv.Atoi(r.PostForm.Get("expenseId"))
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}

		expenseItemParticipationCreateForm := expenseItemParticipationCreateForm{
			MezaniId:      mezaniId,
			ExpenseId:     expenseId,
			ExpenseItemId: expenseItemId,
			Amount:        amount,
			Share:         share,
			ShareType:     shareType,
			ShareTypes:    domain.ShareTypes,
		}

		expenseItemParticipationCreateForm.NotBlank("shareType", string(shareType))
		expenseItemParticipationCreateForm.NotNegative("TotalAmount", totalAmount)
		expenseItemParticipationCreateForm.NotNegative("amount", amount)
		expenseItemParticipationCreateForm.NotNegative("quantity", quantity)
	})
}
