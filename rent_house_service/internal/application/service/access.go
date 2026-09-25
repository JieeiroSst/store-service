package service

import (
	"context"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/inbound"
	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
)

func stayAccess(ctx context.Context, homestays outbound.HomestayRepository, actor inbound.Principal, guestID, homestayID int64) (allowed, privileged bool) {
	if actor.IsAdmin() {
		return true, true
	}
	if actor.UserID == guestID {
		return true, false
	}
	h, err := homestays.Get(ctx, homestayID)
	if err != nil || h.HostID == 0 || h.HostID != actor.UserID {
		return false, false
	}
	return true, true
}
