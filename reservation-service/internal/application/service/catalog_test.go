package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

func newCatalog(h domain.Hotel) (*CatalogService, *fakeHotels) {
	repo := &fakeHotels{hotel: h, rtypes: map[int64]domain.RoomType{
		3: {ID: 3, HotelID: 1, Name: "Deluxe", Capacity: 2, Active: true},
		9: {ID: 9, HotelID: 2, Name: "Someone else's"},
	}}
	s := NewCatalogService(repo, nil)
	s.now = func() time.Time { return time.Date(2027, 1, 10, 9, 0, 0, 0, time.UTC) }
	return s, repo
}

func i(n int) *int { return &n }

func TestOnlyTheManagerEditsTheCatalogue(t *testing.T) {
	ctx := context.Background()
	svc, repo := newCatalog(activeHotel())
	in := inbound.RoomTypeInput{Name: "Suite", Capacity: 3}
	if _, err := svc.CreateRoomType(ctx, otherP, 1, in); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other manager: %v", err)
	}
	if _, err := svc.CreateRoomType(ctx, guestP, 1, in); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("guest: %v", err)
	}
	if rt, err := svc.CreateRoomType(ctx, ownerP, 1, in); err != nil || !rt.Active || len(repo.created) != 1 {
		t.Fatalf("manager: %+v %v", rt, err)
	}
	if _, err := svc.CreateRoomType(ctx, adminP, 1, in); err != nil {
		t.Fatalf("admin: %v", err)
	}
	// An unpublished hotel is invisible to strangers.
	h := activeHotel()
	h.Status = domain.HotelPending
	svc, _ = newCatalog(h)
	if _, err := svc.CreateRoomType(ctx, otherP, 1, in); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("pending, stranger: %v", err)
	}
	if _, err := svc.RoomTypes(ctx, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("public read of a pending hotel: %v", err)
	}
}

func TestRoomTypeValidation(t *testing.T) {
	svc, _ := newCatalog(activeHotel())
	for name, in := range map[string]inbound.RoomTypeInput{
		"no name":       {Capacity: 2},
		"zero capacity": {Name: "x"},
		"huge capacity": {Name: "x", Capacity: 11},
		"negative size": {Name: "x", Capacity: 2, SizeM2: -1},
	} {
		if _, err := svc.CreateRoomType(context.Background(), ownerP, 1, in); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestInventoryRules(t *testing.T) {
	ctx := context.Background()
	svc, repo := newCatalog(activeHotel())
	rate := "5000000"
	ok := inbound.InventoryCommand{HotelID: 1, RoomTypeID: 3, From: day(12), To: day(20), Total: i(10), Rate: &rate}
	if err := svc.SetInventory(ctx, ownerP, ok); err != nil || repo.inv == nil || *repo.inv.Total != 10 || *repo.inv.Rate != "5000000" {
		t.Fatalf("set: %v %+v", err, repo.inv)
	}
	bad := func(mod func(*inbound.InventoryCommand)) error {
		c := ok
		mod(&c)
		return svc.SetInventory(ctx, ownerP, c)
	}
	neg := "-5"
	off := ""
	for name, err := range map[string]error{
		"empty range":         bad(func(c *inbound.InventoryCommand) { c.To = c.From }),
		"past nights":         bad(func(c *inbound.InventoryCommand) { c.From, c.To = day(1), day(3) }),
		"a hundred years":     bad(func(c *inbound.InventoryCommand) { c.To = c.From.AddDate(0, 0, 400) }),
		"negative total":      bad(func(c *inbound.InventoryCommand) { c.Total = i(-1) }),
		"negative rate":       bad(func(c *inbound.InventoryCommand) { c.Rate = &neg }),
		"nothing to change":   bad(func(c *inbound.InventoryCommand) { c.Total, c.Rate = nil, nil }),
		"room type elsewhere": bad(func(c *inbound.InventoryCommand) { c.RoomTypeID = 9 }),
	} {
		if err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	// "" takes nights off sale, and omitting the total is allowed when a rate is given.
	if err := bad(func(c *inbound.InventoryCommand) { c.Total, c.Rate = nil, &off }); err != nil || *repo.inv.Rate != "" || repo.inv.Total != nil {
		t.Fatalf("rate off: %v %+v", err, repo.inv)
	}
	if err := svc.SetInventory(ctx, otherP, ok); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other manager: %v", err)
	}
}

func TestServicesValidation(t *testing.T) {
	ctx := context.Background()
	svc, _ := newCatalog(activeHotel())
	for name, in := range map[string]inbound.ServiceInput{
		"no name":   {Price: "1", Unit: domain.UnitStay},
		"bad price": {Name: "x", Price: "-1", Unit: domain.UnitStay},
		"bad unit":  {Name: "x", Price: "1", Unit: "week"},
	} {
		if _, err := svc.CreateService(ctx, ownerP, 1, in); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := svc.CreateService(ctx, ownerP, 1, inbound.ServiceInput{Name: "Spa", Price: "900000", Unit: domain.UnitStay}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateService(ctx, ownerP, 1, 404, inbound.ServiceInput{Name: "x", Price: "1", Unit: domain.UnitStay}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown service: %v", err)
	}
}
