package http

import (
	"errors"
	"fmt"
	"github.com/aomargithub/mezan/internal/domain"
	"net/http"
	"strconv"
	"time"
)

type expenseParticipationCreateForm struct {
	MezaniId           int
	ExpenseId          int
	ShareType          domain.ShareType
	Share              float32
	Amount             float32
	ExpenseTotalAmount float32
	ShareTypes         []domain.ShareType
	CommonCreateView
}

func (s Server) getExpenseParticipationCreateHandler() http.Handler {
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
		mezaniId, totalAmount, err := s.expenseService.GetMezaniIdTotalAmount(expenseId)
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
		expenseShare, _ := s.expenseShareService.GetByExpenseIdParticipantId(expenseId, userId)
		expenseParticipationCreateForm := expenseParticipationCreateForm{
			ExpenseId:          expenseId,
			MezaniId:           mezaniId,
			CommonCreateView:   s.commonCreateView(r),
			ShareTypes:         domain.ShareTypes,
			ShareType:          expenseShare.ShareType,
			Amount:             expenseShare.Amount,
			Share:              expenseShare.Share,
			ExpenseTotalAmount: totalAmount,
		}
		_ = tx.Commit()
		s.render(w, r, "expenseParticipationCreate.tmpl", http.StatusOK, expenseParticipationCreateForm)
	})
}

func (s Server) postExpenseParticipationCreateHandler() http.Handler {
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

		totalAmount, allocatedAmount, err := s.expenseService.GetTotalAllocatedAmounts(expenseId)

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

		expenseParticipationCreateForm := expenseParticipationCreateForm{
			MezaniId:           mezaniId,
			ExpenseId:          expenseId,
			Amount:             amount,
			Share:              share,
			ShareType:          *shareType,
			ShareTypes:         domain.ShareTypes,
			ExpenseTotalAmount: totalAmount,
			CommonCreateView:   s.commonCreateView(r),
		}

		expenseParticipationCreateForm.NotBlank("ShareType", string(*shareType))
		expenseParticipationCreateForm.NotLessThan("Share", share, 0)
		if *shareType == domain.PERCENTAGE {
			expenseParticipationCreateForm.NotGreaterThan("Share", share, 100)
		} else {
			expenseParticipationCreateForm.NotGreaterThan("Share", share, totalAmount-allocatedAmount)
		}

		if !expenseParticipationCreateForm.Valid() {
			s.render(w, r, "expenseParticipationCreate.tmpl", http.StatusBadRequest, expenseParticipationCreateForm)
			return
		}
		now := time.Now()
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
			ShareType: *shareType,
			Share:     share,
			Amount:    amount,
			CreatedAt: now,
		}
		oldAmount, err := s.expenseShareService.Participate(expenseShare)
		if err != nil {
			s.serverError(w, r, err)
			return
		}
		amountDelta := amount
		if oldAmount != nil {
			amountDelta = amountDelta - *oldAmount

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
		http.Redirect(w, r, fmt.Sprintf("/expenses/%d", expenseId), http.StatusSeeOther)
	})
}
