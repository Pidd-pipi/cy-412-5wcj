package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/dto"
	"github.com/smartestate/smartestate/internal/service"
	"strconv"
)

type RepairHandler struct {
	Handler
	svc *service.RepairService
}

func NewRepairHandler(s *service.RepairService, h *Handler) *RepairHandler {
	return &RepairHandler{Handler: *h, svc: s}
}
func (h *RepairHandler) List(c *gin.Context) {
	v, e := h.svc.List(c.Query("status"))
	if e != nil {
		Fail(c, 500, 50001, e.Error())
		return
	}
	OK(c, v)
}
func (h *RepairHandler) Create(c *gin.Context) {
	var r dto.CreateRepairRequest
	if !Bind(c, &r, h.Validate) {
		return
	}
	v, e := h.svc.Create(c.GetUint("userID"), r.Title, r.Description, r.Type, r.Images)
	if e != nil {
		Fail(c, 500, 50001, e.Error())
		return
	}
	OK(c, v)
}
func (h *RepairHandler) Assign(c *gin.Context) {
	var r dto.AssignRepairRequest
	if !Bind(c, &r, h.Validate) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.Assign(uint(id), r.HandlerID, c.GetString("role"))
	if e != nil {
		Fail(c, 400, 40001, e.Error())
		return
	}
	OK(c, v)
}
func (h *RepairHandler) Status(c *gin.Context) {
	var r dto.UpdateRepairStatusRequest
	if !Bind(c, &r, h.Validate) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.UpdateStatus(uint(id), r.Status, r.Rating, c.GetString("role"))
	if e != nil {
		Fail(c, 400, 40001, e.Error())
		return
	}
	OK(c, v)
}
func (h *RepairHandler) Confirm(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.Confirm(uint(id), c.GetUint("userID"))
	if e != nil {
		acceptFail(c, e)
		return
	}
	OK(c, v)
}
func (h *RepairHandler) Reject(c *gin.Context) {
	var r dto.RejectRepairRequest
	if !Bind(c, &r, h.Validate) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.Reject(uint(id), c.GetUint("userID"), r.Reason)
	if e != nil {
		acceptFail(c, e)
		return
	}
	OK(c, v)
}
func acceptFail(c *gin.Context, e error) {
	switch {
	case errors.Is(e, service.ErrRepairNotOwner):
		Fail(c, 403, constants.CodeForbidden, constants.MessageRepairNotOwner)
	case errors.Is(e, service.ErrRepairNotAccepting):
		Fail(c, 400, constants.CodeBadRequest, constants.MessageRepairNotAccepting)
	default:
		Fail(c, 400, 40001, e.Error())
	}
}
