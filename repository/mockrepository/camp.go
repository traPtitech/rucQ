// Code generated by MockGen. DO NOT EDIT.
// Source: camp.go
//
// Generated by this command:
//
//	mockgen -source=camp.go -destination=mockrepository/camp.go -package=mockrepository
//

// Package mockrepository is a generated GoMock package.
package mockrepository

import (
	context "context"
	reflect "reflect"

	model "github.com/traPtitech/rucQ/model"
	gomock "go.uber.org/mock/gomock"
)

// MockCampRepository is a mock of CampRepository interface.
type MockCampRepository struct {
	ctrl     *gomock.Controller
	recorder *MockCampRepositoryMockRecorder
	isgomock struct{}
}

// MockCampRepositoryMockRecorder is the mock recorder for MockCampRepository.
type MockCampRepositoryMockRecorder struct {
	mock *MockCampRepository
}

// NewMockCampRepository creates a new mock instance.
func NewMockCampRepository(ctrl *gomock.Controller) *MockCampRepository {
	mock := &MockCampRepository{ctrl: ctrl}
	mock.recorder = &MockCampRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCampRepository) EXPECT() *MockCampRepositoryMockRecorder {
	return m.recorder
}

// AddCampParticipant mocks base method.
func (m *MockCampRepository) AddCampParticipant(ctx context.Context, campID uint, user *model.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddCampParticipant", ctx, campID, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddCampParticipant indicates an expected call of AddCampParticipant.
func (mr *MockCampRepositoryMockRecorder) AddCampParticipant(ctx, campID, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCampParticipant", reflect.TypeOf((*MockCampRepository)(nil).AddCampParticipant), ctx, campID, user)
}

// CreateCamp mocks base method.
func (m *MockCampRepository) CreateCamp(camp *model.Camp) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCamp", camp)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateCamp indicates an expected call of CreateCamp.
func (mr *MockCampRepositoryMockRecorder) CreateCamp(camp any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCamp", reflect.TypeOf((*MockCampRepository)(nil).CreateCamp), camp)
}

// DeleteCamp mocks base method.
func (m *MockCampRepository) DeleteCamp(ctx context.Context, campID uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCamp", ctx, campID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCamp indicates an expected call of DeleteCamp.
func (mr *MockCampRepositoryMockRecorder) DeleteCamp(ctx, campID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCamp", reflect.TypeOf((*MockCampRepository)(nil).DeleteCamp), ctx, campID)
}

// GetCampByID mocks base method.
func (m *MockCampRepository) GetCampByID(id uint) (*model.Camp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCampByID", id)
	ret0, _ := ret[0].(*model.Camp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCampByID indicates an expected call of GetCampByID.
func (mr *MockCampRepositoryMockRecorder) GetCampByID(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCampByID", reflect.TypeOf((*MockCampRepository)(nil).GetCampByID), id)
}

// GetCampParticipants mocks base method.
func (m *MockCampRepository) GetCampParticipants(ctx context.Context, campID uint) ([]model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCampParticipants", ctx, campID)
	ret0, _ := ret[0].([]model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCampParticipants indicates an expected call of GetCampParticipants.
func (mr *MockCampRepositoryMockRecorder) GetCampParticipants(ctx, campID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCampParticipants", reflect.TypeOf((*MockCampRepository)(nil).GetCampParticipants), ctx, campID)
}

// GetCamps mocks base method.
func (m *MockCampRepository) GetCamps() ([]model.Camp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCamps")
	ret0, _ := ret[0].([]model.Camp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCamps indicates an expected call of GetCamps.
func (mr *MockCampRepositoryMockRecorder) GetCamps() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCamps", reflect.TypeOf((*MockCampRepository)(nil).GetCamps))
}

// IsCampParticipant mocks base method.
func (m *MockCampRepository) IsCampParticipant(ctx context.Context, campID uint, userID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsCampParticipant", ctx, campID, userID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsCampParticipant indicates an expected call of IsCampParticipant.
func (mr *MockCampRepositoryMockRecorder) IsCampParticipant(ctx, campID, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsCampParticipant", reflect.TypeOf((*MockCampRepository)(nil).IsCampParticipant), ctx, campID, userID)
}

// RemoveCampParticipant mocks base method.
func (m *MockCampRepository) RemoveCampParticipant(ctx context.Context, campID uint, user *model.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveCampParticipant", ctx, campID, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveCampParticipant indicates an expected call of RemoveCampParticipant.
func (mr *MockCampRepositoryMockRecorder) RemoveCampParticipant(ctx, campID, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveCampParticipant", reflect.TypeOf((*MockCampRepository)(nil).RemoveCampParticipant), ctx, campID, user)
}

// UpdateCamp mocks base method.
func (m *MockCampRepository) UpdateCamp(ctx context.Context, campID uint, camp *model.Camp) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCamp", ctx, campID, camp)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCamp indicates an expected call of UpdateCamp.
func (mr *MockCampRepositoryMockRecorder) UpdateCamp(ctx, campID, camp any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCamp", reflect.TypeOf((*MockCampRepository)(nil).UpdateCamp), ctx, campID, camp)
}
