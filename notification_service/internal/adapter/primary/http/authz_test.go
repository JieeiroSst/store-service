package http

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

type fakeDevices struct {
	port.UserDeviceUsecase
	devices    map[uint]model.UserDevice
	registered port.RegisterDeviceInput
	listedFor  uint
}

func (f *fakeDevices) RegisterDevice(_ context.Context, in port.RegisterDeviceInput) (*model.UserDevice, error) {
	f.registered = in
	return &model.UserDevice{ID: 9, UserID: in.UserID}, nil
}
func (f *fakeDevices) GetDevice(_ context.Context, id uint) (*model.UserDevice, error) {
	d, ok := f.devices[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &d, nil
}
func (f *fakeDevices) ListDevices(_ context.Context, userID uint) ([]model.UserDevice, error) {
	f.listedFor = userID
	return nil, nil
}
func (f *fakeDevices) DeleteDevice(context.Context, uint) error { return nil }
func (f *fakeDevices) GetContact(_ context.Context, userID uint) (*model.UserContact, error) {
	return &model.UserContact{UserID: userID}, nil
}

func TestNotificationAuthorization(t *testing.T) {
	fakeUserService(t)
	devices := &fakeDevices{devices: map[uint]model.UserDevice{1: {ID: 1, UserID: 7}, 2: {ID: 2, UserID: 8}}}
	router := NewRouter(&Handler{device: devices})
	user := testToken("7", "user")
	admin := testToken("1", "admin")

	cases := []struct {
		name string
		call authCall
		want int
	}{
		{"anonymous send blocked", authCall{method: "POST", path: "/api/v1/notifications", body: `{}`}, 401},
		{"user cannot send", authCall{method: "POST", path: "/api/v1/notifications/push", body: `{}`, bearer: user}, 403},
		{"user cannot read campaigns", authCall{method: "GET", path: "/api/v1/campaigns", bearer: user}, 403},
		{"bad internal token", authCall{method: "GET", path: "/api/v1/campaigns", internal: "nope"}, 401},
		{"own contact", authCall{method: "GET", path: "/api/v1/users/7/contact", bearer: user}, 200},
		{"foreign contact", authCall{method: "GET", path: "/api/v1/users/8/contact", bearer: user}, 403},
		{"admin foreign contact", authCall{method: "GET", path: "/api/v1/users/8/contact", bearer: admin}, 200},
		{"internal contact", authCall{method: "GET", path: "/api/v1/users/8/contact", internal: testInternalToken}, 200},
		{"own device", authCall{method: "GET", path: "/api/v1/devices/1", bearer: user}, 200},
		{"foreign device", authCall{method: "GET", path: "/api/v1/devices/2", bearer: user}, 403},
		{"foreign device delete", authCall{method: "DELETE", path: "/api/v1/devices/2", bearer: user}, 403},
		{"foreign device list", authCall{method: "GET", path: "/api/v1/devices?user_id=8", bearer: user}, 403},
		{"own device list", authCall{method: "GET", path: "/api/v1/devices?user_id=7", bearer: user}, 200},
	}
	for _, c := range cases {
		if got, body := doAuth(router, c.call); got != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, got, c.want, body)
		}
	}

	if code, _ := doAuth(router, authCall{method: "POST", path: "/api/v1/devices", body: `{"user_id":8,"device_token":"t"}`, bearer: user}); code != 200 || devices.registered.UserID != 7 {
		t.Fatalf("register forced owner: code=%d user=%d", code, devices.registered.UserID)
	}
}
