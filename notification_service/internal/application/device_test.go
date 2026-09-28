package application

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

const fcmToken = "dGhpcyBpcyBub3QgcmVhbA:APA91bFakeRegistrationTokenForTestsOnly_abcdefghijklmnopqrstuvwxyz0123456789"

func activeTokens(s *store, userID uint) []string {
	var tokens []string
	for _, d := range s.devices {
		if d.UserID == userID && d.IsActive {
			tokens = append(tokens, d.DeviceToken)
		}
	}
	return tokens
}

func TestRegisterDeviceValidation(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	bad := []port.RegisterDeviceInput{
		{UserID: 0, Token: fcmToken, DeviceType: "android"},
		{UserID: 1, Token: " ", DeviceType: "android"},
		{UserID: 1, Token: fcmToken, DeviceType: "windows"},
		{UserID: 1, Token: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", DeviceType: "ios"},
	}
	for i, in := range bad {
		if _, err := svc.RegisterDevice(ctx, in); !errors.Is(err, common.ErrInvalidRequest) {
			t.Errorf("case %d: got %v", i, err)
		}
	}
}

func TestSameTokenMovesToNewUser(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	a, _ := svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken, DeviceType: "Android"})
	b, err := svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 2, Token: fcmToken, DeviceType: "android"})
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID || len(f.s.devices) != 1 || len(activeTokens(f.s, 1)) != 0 || len(activeTokens(f.s, 2)) != 1 || b.DeviceType != model.DeviceAndroid {
		t.Errorf("token not moved: a=%+v b=%+v devices=%d", a, b, len(f.s.devices))
	}
}

func TestTokenRotationOnSameInstallReplacesOldToken(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "old", DeviceID: "install-1", DeviceType: "ios"})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "other-phone", DeviceID: "install-2", DeviceType: "android"})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "new", DeviceID: "install-1", DeviceType: "ios"})
	got := activeTokens(f.s, 1)
	if len(got) != 2 {
		t.Errorf("active tokens = %v, want the rotated token and the other phone", got)
	}
	for _, tk := range got {
		if tk == fcmToken+"old" {
			t.Error("rotated-away token is still active")
		}
	}
}

func TestSingleDevicePolicy(t *testing.T) {
	f := newFixture()
	multi := f.devices(DevicePolicy{})
	single := f.devices(DevicePolicy{SingleDevicePerUser: true})
	_, _ = multi.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "old-phone", DeviceType: "android"})
	_, _ = multi.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "new-phone", DeviceType: "ios"})
	if len(activeTokens(f.s, 1)) != 2 {
		t.Fatalf("multi-device should keep both, got %v", activeTokens(f.s, 1))
	}
	_, _ = single.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "newest", DeviceType: "ios"})
	if got := activeTokens(f.s, 1); len(got) != 1 || got[0] != fcmToken+"newest" {
		t.Errorf("single-device should keep only the latest login, got %v", got)
	}
}

func TestUnregister(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "a", DeviceType: "android"})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "b", DeviceType: "ios"})

	if n, err := svc.UnregisterDevice(ctx, port.UnregisterDeviceInput{Token: " " + fcmToken + "a "}); err != nil || n != 1 {
		t.Fatalf("logout one device: %d %v", n, err)
	}
	if n, _ := svc.UnregisterDevice(ctx, port.UnregisterDeviceInput{UserID: 1}); n != 1 || len(activeTokens(f.s, 1)) != 0 {
		t.Errorf("logout everywhere: %d, active=%v", n, activeTokens(f.s, 1))
	}
	if _, err := svc.UnregisterDevice(ctx, port.UnregisterDeviceInput{}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Errorf("empty unregister: %v", err)
	}
}

func TestNewLoginReceivesAndLoggedOutDoesNot(t *testing.T) {
	f := newFixture()
	devices := f.devices(DevicePolicy{})
	_, _ = devices.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 5, Token: fcmToken + "old", DeviceType: "android"})
	_, _ = devices.UnregisterDevice(ctx, port.UnregisterDeviceInput{Token: fcmToken + "old"})
	_, _ = devices.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 5, Token: fcmToken + "new", DeviceType: "ios"})

	if err := f.notif.Dispatch(ctx, pushNotification(f, 5), 1, false); err != nil {
		t.Fatal(err)
	}
	if len(f.push.calls) != 1 || len(f.push.calls[0]) != 1 || f.push.calls[0][0] != fcmToken+"new" {
		t.Errorf("sent to %v, want only the new device", f.push.calls)
	}
}

func TestRegisterRejectsTokenFCMDoesNotAccept(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{ValidateOnRegister: true}, fcmToken+"stale")
	if _, err := svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "stale", DeviceType: "android"}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Errorf("got %v", err)
	}
	if len(f.s.devices) != 0 {
		t.Error("rejected token was stored")
	}
	lenient := f.devices(DevicePolicy{}, fcmToken+"stale")
	if _, err := lenient.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "stale", DeviceType: "android"}); err != nil {
		t.Errorf("validation disabled should accept: %v", err)
	}
}

func TestPushByEmailAndPhone(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	if _, err := svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 42, Token: fcmToken + "ios", DeviceType: "ios", Email: " An@Shop.VN ", Phone: "+84 912 345 678"}); err != nil {
		t.Fatal(err)
	}
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 42, Token: fcmToken + "android", DeviceType: "android"})

	for _, n := range []*model.Notification{
		{Type: "push", Email: "an@shop.vn", Title: "t"},
		{Type: "push", Phone: "0912.345.678", Title: "t"},
	} {
		created, err := f.notif.CreateNotification(ctx, n)
		if err != nil {
			t.Fatal(err)
		}
		f.push.calls = nil
		if err := f.notif.Dispatch(ctx, created, 1, false); err != nil {
			t.Fatal(err)
		}
		if len(f.push.calls) != 1 || len(f.push.calls[0]) != 2 {
			t.Errorf("%+v: sent to %v, want both devices of user 42", n, f.push.calls)
		}
	}

	unknown, _ := f.notif.CreateNotification(ctx, &model.Notification{Type: "push", Email: "nobody@shop.vn", Title: "t"})
	if err := f.notif.Dispatch(ctx, unknown, 1, false); !errors.Is(err, common.ErrPermanent) {
		t.Errorf("unknown email should fail permanently, got %v", err)
	}
	if _, err := f.notif.CreateNotification(ctx, &model.Notification{Type: "push", Title: "t"}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Errorf("push without recipient: %v", err)
	}
}

func TestContactMovesToNewOwner(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	_, _ = svc.UpsertContact(ctx, port.ContactInput{UserID: 1, Email: "shared@shop.vn", Phone: "0912345678"})
	c, err := svc.UpsertContact(ctx, port.ContactInput{UserID: 2, Email: "SHARED@shop.vn"})
	if err != nil || *c.Email != "shared@shop.vn" {
		t.Fatalf("%+v %v", c, err)
	}
	if id, _ := f.contacts.UserIDByEmail(ctx, "shared@shop.vn"); id != 2 {
		t.Errorf("email owner = %d, want 2", id)
	}
	if id, _ := f.contacts.UserIDByPhone(ctx, "0912345678"); id != 1 {
		t.Errorf("phone should stay with user 1, got %d", id)
	}
	if _, err := svc.UpsertContact(ctx, port.ContactInput{UserID: 3, Phone: "abc"}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Errorf("bad phone: %v", err)
	}
}

func TestStaleTokensAreDeactivated(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{StaleAfter: 270 * 24 * time.Hour})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "fresh", DeviceType: "android"})
	old := f.s.addDevice(1, fcmToken+"old", true)
	old.LastUsedAt = time.Now().Add(-300 * 24 * time.Hour)

	n, err := svc.DeactivateStaleTokens(ctx)
	if err != nil || n != 1 || f.s.devices[old.ID].IsActive {
		t.Errorf("deactivated %d, err %v, old active=%v", n, err, f.s.devices[old.ID].IsActive)
	}
}

func TestNormalizePhone(t *testing.T) {
	for in, want := range map[string]string{"+84 912 345 678": "0912345678", "84912345678": "0912345678", "0912-345-678": "0912345678", "+14155550100": "+14155550100"} {
		if got, ok := model.NormalizePhone(in); !ok || got != want {
			t.Errorf("%q -> %q", in, got)
		}
	}
	if _, ok := model.NormalizePhone("12ab"); ok {
		t.Error("letters accepted")
	}
}

func TestTokenIsNeverExposedAndCannotBeOverwritten(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	d, _ := svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken, DeviceID: "inst-1", DeviceType: "ios"})

	raw, _ := json.Marshal(d)
	if strings.Contains(string(raw), fcmToken) || strings.Contains(string(raw), "device_token") {
		t.Errorf("device JSON leaks the token: %s", raw)
	}
	if p := d.TokenPreview(); !strings.HasPrefix(p, fcmToken[:6]) || strings.Contains(p, fcmToken[6:20]) {
		t.Errorf("preview = %q", p)
	}

	_, _ = svc.UpdateDevice(ctx, &model.UserDevice{ID: d.ID, DeviceToken: "attacker-token", DeviceType: "android"})
	if got := f.s.devices[d.ID].DeviceToken; got != fcmToken {
		t.Errorf("PUT replaced the stored token with %q", got)
	}
}

func TestUnregisterByDeviceID(t *testing.T) {
	f := newFixture()
	svc := f.devices(DevicePolicy{})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "a", DeviceID: "phone", DeviceType: "android"})
	_, _ = svc.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 1, Token: fcmToken + "b", DeviceID: "tablet", DeviceType: "ios"})

	if _, err := svc.UnregisterDevice(ctx, port.UnregisterDeviceInput{DeviceID: "phone"}); !errors.Is(err, common.ErrInvalidRequest) {
		t.Errorf("device_id without user_id: %v", err)
	}
	if n, err := svc.UnregisterDevice(ctx, port.UnregisterDeviceInput{UserID: 1, DeviceID: "phone"}); err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if got := activeTokens(f.s, 1); len(got) != 1 || got[0] != fcmToken+"b" {
		t.Errorf("active = %v", got)
	}
	list, _ := svc.ListDevices(ctx, 1)
	if len(list) != 2 {
		t.Errorf("list by user = %d devices, want 2 (active and logged out)", len(list))
	}
}
