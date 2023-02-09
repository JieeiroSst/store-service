package mocks

import (
	reflect "reflect"

	model "github.com/JieeiroSst/authorize-service/model"
	gomock "github.com/golang/mock/gomock"
	otpgo "github.com/jltorresm/otpgo"
)

// MockOTP is a mock of OTP interface.
type MockOTP struct {
	ctrl     *gomock.Controller
	recorder *MockOTPMockRecorder
}

// MockOTPMockRecorder is the mock recorder for MockOTP.
type MockOTPMockRecorder struct {
	mock *MockOTP
}

// NewMockOTP creates a new mock instance.
func NewMockOTP(ctrl *gomock.Controller) *MockOTP {
	mock := &MockOTP{ctrl: ctrl}
	mock.recorder = &MockOTPMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOTP) EXPECT() *MockOTPMockRecorder {
	return m.recorder
}

// Authorize mocks base method.
func (m *MockOTP) Authorize(otp, username string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authorize", otp, username)
	ret0, _ := ret[0].(error)
	return ret0
}

// Authorize indicates an expected call of Authorize.
func (mr *MockOTPMockRecorder) Authorize(otp, username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authorize", reflect.TypeOf((*MockOTP)(nil).Authorize), otp, username)
}

// CreateOtpByUser mocks base method.
func (m *MockOTP) CreateOtpByUser(username string) (*model.OTP, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOtpByUser", username)
	ret0, _ := ret[0].(*model.OTP)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOtpByUser indicates an expected call of CreateOtpByUser.
func (mr *MockOTPMockRecorder) CreateOtpByUser(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOtpByUser", reflect.TypeOf((*MockOTP)(nil).CreateOtpByUser), username)
}

// generate mocks base method.
func (m *MockOTP) generate(username string) otpgo.TOTP {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "generate", username)
	ret0, _ := ret[0].(otpgo.TOTP)
	return ret0
}

// generate indicates an expected call of generate.
func (mr *MockOTPMockRecorder) generate(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "generate", reflect.TypeOf((*MockOTP)(nil).generate), username)
}