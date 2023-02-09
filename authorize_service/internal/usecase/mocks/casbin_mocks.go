package mocks

import (
	reflect "reflect"

	model "github.com/JieeiroSst/authorize-service/model"
	gomock "github.com/golang/mock/gomock"
)

// MockCasbins is a mock of Casbins interface.
type MockCasbins struct {
	ctrl     *gomock.Controller
	recorder *MockCasbinsMockRecorder
}

// MockCasbinsMockRecorder is the mock recorder for MockCasbins.
type MockCasbinsMockRecorder struct {
	mock *MockCasbins
}

// NewMockCasbins creates a new mock instance.
func NewMockCasbins(ctrl *gomock.Controller) *MockCasbins {
	mock := &MockCasbins{ctrl: ctrl}
	mock.recorder = &MockCasbinsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCasbins) EXPECT() *MockCasbinsMockRecorder {
	return m.recorder
}

// CasbinRuleAll mocks base method.
func (m *MockCasbins) CasbinRuleAll() ([]model.CasbinRule, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CasbinRuleAll")
	ret0, _ := ret[0].([]model.CasbinRule)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CasbinRuleAll indicates an expected call of CasbinRuleAll.
func (mr *MockCasbinsMockRecorder) CasbinRuleAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CasbinRuleAll", reflect.TypeOf((*MockCasbins)(nil).CasbinRuleAll))
}

// CasbinRuleById mocks base method.
func (m *MockCasbins) CasbinRuleById(id int) (*model.CasbinRule, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CasbinRuleById", id)
	ret0, _ := ret[0].(*model.CasbinRule)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CasbinRuleById indicates an expected call of CasbinRuleById.
func (mr *MockCasbinsMockRecorder) CasbinRuleById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CasbinRuleById", reflect.TypeOf((*MockCasbins)(nil).CasbinRuleById), id)
}

// CreateCasbinRule mocks base method.
func (m *MockCasbins) CreateCasbinRule(casbin model.CasbinRule) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCasbinRule", casbin)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateCasbinRule indicates an expected call of CreateCasbinRule.
func (mr *MockCasbinsMockRecorder) CreateCasbinRule(casbin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCasbinRule", reflect.TypeOf((*MockCasbins)(nil).CreateCasbinRule), casbin)
}

// DeleteCasbinRule mocks base method.
func (m *MockCasbins) DeleteCasbinRule(id int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCasbinRule", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCasbinRule indicates an expected call of DeleteCasbinRule.
func (mr *MockCasbinsMockRecorder) DeleteCasbinRule(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCasbinRule", reflect.TypeOf((*MockCasbins)(nil).DeleteCasbinRule), id)
}

// EnforceCasbin mocks base method.
func (m *MockCasbins) EnforceCasbin(auth model.CasbinAuth) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnforceCasbin", auth)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnforceCasbin indicates an expected call of EnforceCasbin.
func (mr *MockCasbinsMockRecorder) EnforceCasbin(auth interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnforceCasbin", reflect.TypeOf((*MockCasbins)(nil).EnforceCasbin), auth)
}

// UpdateCasbinMethod mocks base method.
func (m *MockCasbins) UpdateCasbinMethod(id int, method string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCasbinMethod", id, method)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCasbinMethod indicates an expected call of UpdateCasbinMethod.
func (mr *MockCasbinsMockRecorder) UpdateCasbinMethod(id, method interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCasbinMethod", reflect.TypeOf((*MockCasbins)(nil).UpdateCasbinMethod), id, method)
}

// UpdateCasbinRuleEndpoint mocks base method.
func (m *MockCasbins) UpdateCasbinRuleEndpoint(id int, endpoint string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCasbinRuleEndpoint", id, endpoint)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCasbinRuleEndpoint indicates an expected call of UpdateCasbinRuleEndpoint.
func (mr *MockCasbinsMockRecorder) UpdateCasbinRuleEndpoint(id, endpoint interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCasbinRuleEndpoint", reflect.TypeOf((*MockCasbins)(nil).UpdateCasbinRuleEndpoint), id, endpoint)
}

// UpdateCasbinRuleName mocks base method.
func (m *MockCasbins) UpdateCasbinRuleName(id int, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCasbinRuleName", id, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCasbinRuleName indicates an expected call of UpdateCasbinRuleName.
func (mr *MockCasbinsMockRecorder) UpdateCasbinRuleName(id, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCasbinRuleName", reflect.TypeOf((*MockCasbins)(nil).UpdateCasbinRuleName), id, name)
}

// UpdateCasbinRulePtype mocks base method.
func (m *MockCasbins) UpdateCasbinRulePtype(id int, ptype string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCasbinRulePtype", id, ptype)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCasbinRulePtype indicates an expected call of UpdateCasbinRulePtype.
func (mr *MockCasbinsMockRecorder) UpdateCasbinRulePtype(id, ptype interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCasbinRulePtype", reflect.TypeOf((*MockCasbins)(nil).UpdateCasbinRulePtype), id, ptype)
}