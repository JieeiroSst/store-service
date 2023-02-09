package mocks

import (
	reflect "reflect"

	model "github.com/JieeiroSst/authorize-service/model"
	gomock "github.com/golang/mock/gomock"
)

// MockOtps is a mock of Otps interface.
type MockOtps struct {
	ctrl     *gomock.Controller
	recorder *MockOtpsMockRecorder
}

// MockOtpsMockRecorder is the mock recorder for MockOtps.
type MockOtpsMockRecorder struct {
	mock *MockOtps
}

// NewMockOtps creates a new mock instance.
func NewMockOtps(ctrl *gomock.Controller) *MockOtps {
	mock := &MockOtps{ctrl: ctrl}
	mock.recorder = &MockOtpsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOtps) EXPECT() *MockOtpsMockRecorder {
	return m.recorder
}

// Authorize mocks base method.
func (m *MockOtps) Authorize(otp, username string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authorize", otp, username)
	ret0, _ := ret[0].(error)
	return ret0
}

// Authorize indicates an expected call of Authorize.
func (mr *MockOtpsMockRecorder) Authorize(otp, username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authorize", reflect.TypeOf((*MockOtps)(nil).Authorize), otp, username)
}

// CreateOtpByUser mocks base method.
func (m *MockOtps) CreateOtpByUser(username string) (*model.OTP, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOtpByUser", username)
	ret0, _ := ret[0].(*model.OTP)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOtpByUser indicates an expected call of CreateOtpByUser.
func (mr *MockOtpsMockRecorder) CreateOtpByUser(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOtpByUser", reflect.TypeOf((*MockOtps)(nil).CreateOtpByUser), username)
}