package service

import (
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"testing"
)

func newRepairService(t *testing.T) (*RepairService, model.User, model.User) {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&model.User{}, &model.Repair{})
	owner := model.User{Phone: "1", Nickname: "业主", Role: "resident"}
	other := model.User{Phone: "2", Nickname: "邻居", Role: "resident"}
	db.Create(&owner)
	db.Create(&other)
	s := NewRepairService(repository.NewRepairRepository(db), repository.NewUserRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
	return s, owner, other
}

func TestRepairAcceptanceFlowTable(t *testing.T) {
	s, owner, other := newRepairService(t)
	v, e := s.Create(owner.ID, "水管漏水", "厨房水管持续滴水", "水电", "")
	if e != nil {
		t.Fatalf("create: %v", e)
	}
	if v, e = s.UpdateStatus(v.ID, constants.RepairStatusProcessing, 0, "staff"); e != nil {
		t.Fatalf("processing: %v", e)
	}
	if v, e = s.UpdateStatus(v.ID, constants.RepairStatusDone, 0, "staff"); e != nil || v.Status != constants.RepairStatusAcceptance {
		t.Fatalf("done should become acceptance, got %s %v", v.Status, e)
	}
	if _, e = s.Confirm(v.ID, other.ID, "resident"); e == nil {
		t.Fatal("non-reporter confirm should fail")
	}
	if _, e = s.Return(v.ID, other.ID, "没修好", "resident"); e == nil {
		t.Fatal("non-reporter return should fail")
	}
	if v, e = s.Return(v.ID, owner.ID, "师傅没修好，仍然漏水", "resident"); e != nil || v.Status != constants.RepairStatusProcessing || v.ReworkCount != 1 || v.ReturnReason == "" {
		t.Fatalf("return got %+v %v", v, e)
	}
	if _, e = s.Confirm(v.ID, owner.ID, "resident"); e == nil {
		t.Fatal("confirm outside acceptance should fail")
	}
	if v, e = s.UpdateStatus(v.ID, constants.RepairStatusAcceptance, 0, "staff"); e != nil || v.Status != constants.RepairStatusAcceptance {
		t.Fatalf("rework done should become acceptance, got %s %v", v.Status, e)
	}
	if v, e = s.Confirm(v.ID, owner.ID, "resident"); e != nil || v.Status != constants.RepairStatusClosed {
		t.Fatalf("confirm got %+v %v", v, e)
	}
	rework := v.ReworkCount
	if v, e = s.Confirm(v.ID, owner.ID, "resident"); e != nil || v.Status != constants.RepairStatusClosed || v.ReworkCount != rework {
		t.Fatalf("repeat confirm must keep final state, got %+v %v", v, e)
	}
	if _, e = s.Return(v.ID, owner.ID, "又坏了", "resident"); e == nil {
		t.Fatal("return on closed repair should fail")
	}
	if _, e = s.UpdateStatus(v.ID, constants.RepairStatusClosed, 0, "staff"); e == nil {
		t.Fatal("staff endpoint must not close a repair")
	}
}
