package http

import (
	"errors"
	"fmt"
	"github.com/aomargithub/mezan/internal/domain"
	"net/http"
	"strconv"
	"time"
)

type expenseItemParticipationCreateForm struct {
	MezaniId               int
	ExpenseId              int
	ExpenseItemId          int
	ShareType              domain.ShareType
	Share                  float32
	Amount                 float32
	ExpenseItemTotalAmount float32
	ShareTypes             []domain.ShareType
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
		mezaniId, expenseId, totalAmount, err := s.expenseItemService.GetExpenseIdMezaniIdTotalAmount(expenseItemId)
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
			ExpenseId:              expenseId,
			MezaniId:               mezaniId,
			ExpenseItemId:          expenseItemId,
			CommonCreateView:       s.commonCreateView(r),
			ShareTypes:             domain.ShareTypes,
			ExpenseItemTotalAmount: totalAmount,
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

		totalAmount, allocatedAmount, err := s.expenseItemService.GetTotalAndAllocatedAmounts(expenseItemId)

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

		err = r.ParseForm()
		if err != nil {
			s.clientError(w, http.StatusBadRequest)
			return
		}
		shareTypeStr := r.PostForm.Get("shareType")
		shareType := domain.NewShareType(shareTypeStr)
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
			MezaniId:               mezaniId,
			ExpenseId:              expenseId,
			ExpenseItemId:          expenseItemId,
			Amount:                 amount,
			Share:                  share,
			ShareType:              *shareType,
			ShareTypes:             domain.ShareTypes,
			ExpenseItemTotalAmount: totalAmount,
			CommonCreateView:       s.commonCreateView(r),
		}

		expenseItemParticipationCreateForm.NotBlank("ShareType", string(*shareType))
		expenseItemParticipationCreateForm.NotLessThan("Share", share, 0)
		if *shareType == domain.PERCENTAGE {
			expenseItemParticipationCreateForm.NotGreaterThan("Share", share, 100)
		} else {
			expenseItemParticipationCreateForm.NotGreaterThan("Share", share, totalAmount-allocatedAmount)
		}

		if !expenseItemParticipationCreateForm.Valid() {
			s.render(w, r, "expenseItemParticipationCreate.tmpl", http.StatusBadRequest, expenseItemParticipationCreateForm)
			return
		}
		now := time.Now()
		expenseItemShare := domain.ExpenseItemShare{
			Mezani: domain.Mezani{
				Id: mezaniId,
			},
			Expense: domain.Expense{
				Id: expenseId,
			},
			ExpenseItem: domain.ExpenseItem{
				Id: expenseItemId,
			},
			Participant: domain.User{
				Id: userId,
			},
			ShareType: *shareType,
			Share:     share,
			Amount:    amount,
			CreatedAt: now,
		}
		oldAmount, err := s.expenseItemShareService.Participate(expenseItemShare)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		amountDelta := amount
		if oldAmount != nil {
			amountDelta = amountDelta - *oldAmount

		}
		err = s.expenseItemService.Participate(amountDelta, now, expenseItemId)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		expenseShare := domain.ExpenseShare{
			Mezani: domain.Mezani{
				Id: mezaniId,
			},
			Expense: domain.Expense{
				Id: expenseId,
			},
			Participant: domain.User{
				Id: userId,
			},
			ShareType: domain.EXACT,
			Share:     amountDelta,
			Amount:    amountDelta,
			CreatedAt: now,
		}
		err = s.expenseShareService.ParticipateInItem(expenseShare)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		err = s.expenseService.Participate(amountDelta, now, expenseId)
		if err != nil {
			s.serverError(w, r, err)
			return
		}

		mezaniShare := domain.MezaniShare{
			Mezani: domain.Mezani{
				Id: mezaniId,
			},
			Participant: domain.User{
				Id: userId,
			},
			ShareType: domain.EXACT,
			Share:     amountDelta,
			Amount:    amountDelta,
			CreatedAt: now,
		}
		err = s.mezaniShareService.ParticipateInChild(mezaniShare)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		err = s.mezaniService.Participate(amountDelta, now, mezaniId)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		_ = tx.Commit()
		s.sessionManager.Put(r.Context(), "flash", "Your participation has been saved successfully!")
		http.Redirect(w, r, fmt.Sprintf("/expenseItems/%d", expenseItemId), http.StatusSeeOther)
	})
}
