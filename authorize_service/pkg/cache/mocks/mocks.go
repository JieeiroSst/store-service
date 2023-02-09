package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockCacheHelper is a mock of CacheHelper interface.
type MockCacheHelper struct {
	ctrl     *gomock.Controller
	recorder *MockCacheHelperMockRecorder
}

// MockCacheHelperMockRecorder is the mock recorder for MockCacheHelper.
type MockCacheHelperMockRecorder struct {
	mock *MockCacheHelper
}

// NewMockCacheHelper creates a new mock instance.
func NewMockCacheHelper(ctrl *gomock.Controller) *MockCacheHelper {
	mock := &MockCacheHelper{ctrl: ctrl}
	mock.recorder = &MockCacheHelperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheHelper) EXPECT() *MockCacheHelperMockRecorder {
	return m.recorder
}

// GetInterfac mocks base method.
func (m *MockCacheHelper) GetInterfac(ctx context.Context, key string, value interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInterfac", ctx, key, value)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInterfac indicates an expected call of GetInterfac.
func (mr *MockCacheHelperMockRecorder) GetInterfac(ctx, key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInterfac", reflect.TypeOf((*MockCacheHelper)(nil).GetInterfac), ctx, key, value)
}

// Set mocks base method.
func (m *MockCacheHelper) Set(ctx context.Context, key string, value interface{}, exppiration time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, key, value, exppiration)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCacheHelperMockRecorder) Set(ctx, key, value, exppiration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheHelper)(nil).Set), ctx, key, value, exppiration)
}