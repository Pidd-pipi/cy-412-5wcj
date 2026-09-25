package service

import (
	"errors"
	"fmt"
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"log/slog"
)

var (
	ErrRepairNotReporter  = errors.New("only the reporter can accept or return this repair")
	ErrRepairNotAccepting = errors.New("repair is not awaiting acceptance")
	ErrRepairCloseDenied  = errors.New("closing a repair requires reporter confirmation")
)

type RepairService struct {
	repo   *repository.RepairRepository
	users  *repository.UserRepository
	logger *slog.Logger
}

func NewRepairService(r *repository.RepairRepository, u *repository.UserRepository, l *slog.Logger) *RepairService {
	return &RepairService{r, u, l}
}
func (s *RepairService) Create(uid uint, title, desc, typ, images string) (model.Repair, error) {
	v := model.Repair{UserID: uid, Title: title, Description: desc, Type: typ, Images: images, Status: constants.RepairStatusPending}
	if e := s.repo.Create(&v); e != nil {
		return v, fmt.Errorf("Repair[user_id=%d] create failed: %w", uid, e)
	}
	return s.repo.ByID(v.ID)
}
func (s *RepairService) List(status string) ([]model.Repair, error) { return s.repo.List(status) }
func (s *RepairService) Assign(id, handlerID uint, role string) (model.Repair, error) {
	handler, e := s.users.ByID(handlerID)
	if e != nil {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign failed: staff not found, current role=%s: %w", id, role, e)
	}
	if handler.Role != constants.UserRoleStaff && handler.Role != constants.UserRoleAdmin {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign failed: handler %d is not staff/admin, current role=%s", id, handlerID, role)
	}
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	v.HandlerID = &handlerID
	v.Handler = nil
	v.User = model.User{}
	v.Status = constants.RepairStatusAssigned
	if e = s.repo.Update(&v); e != nil {
		return v, fmt.Errorf("Repair[id=%d] assign failed: %w", id, e)
	}
	return s.repo.ByID(id)
}
func (s *RepairService) UpdateStatus(id uint, status string, rating int, role string) (model.Repair, error) {
	if !constants.ValidRepairStatuses[status] {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] status failed: invalid status, current role=%s", id, role)
	}
	if status == constants.RepairStatusClosed {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] status failed: %w, current role=%s", id, ErrRepairCloseDenied, role)
	}
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if status == constants.RepairStatusDone {
		status = constants.RepairStatusAcceptance
	}
	v.Status = status
	if rating > 0 {
		v.Rating = rating
	}
	if e = s.repo.Update(&v); e != nil {
		return v, fmt.Errorf("Repair[id=%d] status failed: %w", id, e)
	}
	return s.repo.ByID(id)
}
func (s *RepairService) Confirm(id, uid uint, role string) (model.Repair, error) {
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if v.UserID != uid {
		return v, fmt.Errorf("Repair[id=%d] confirm failed: %w, current role=%s", id, ErrRepairNotReporter, role)
	}
	if v.Status == constants.RepairStatusClosed {
		return v, nil
	}
	if v.Status != constants.RepairStatusAcceptance {
		return v, fmt.Errorf("Repair[id=%d] confirm failed: %w (status=%s), current role=%s", id, ErrRepairNotAccepting, v.Status, role)
	}
	v.Status = constants.RepairStatusClosed
	if e = s.repo.Update(&v); e != nil {
		return v, fmt.Errorf("Repair[id=%d] confirm failed: %w", id, e)
	}
	return s.repo.ByID(id)
}
func (s *RepairService) Return(id, uid uint, reason, role string) (model.Repair, error) {
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if v.UserID != uid {
		return v, fmt.Errorf("Repair[id=%d] return failed: %w, current role=%s", id, ErrRepairNotReporter, role)
	}
	if v.Status != constants.RepairStatusAcceptance {
		return v, fmt.Errorf("Repair[id=%d] return failed: %w (status=%s), current role=%s", id, ErrRepairNotAccepting, v.Status, role)
	}
	v.Status = constants.RepairStatusProcessing
	v.ReturnReason = reason
	v.ReworkCount++
	if e = s.repo.Update(&v); e != nil {
		return v, fmt.Errorf("Repair[id=%d] return failed: %w", id, e)
	}
	return s.repo.ByID(id)
}
func (s *RepairService) OpenCount() (int64, error) { return s.repo.CountOpen() }
