// Code generated by MockGen. DO NOT EDIT.
// Source: payment.go
//
// Generated by this command:
//
//	mockgen -source=payment.go -destination=mockrepository/payment.go -package=mockrepository
//

// Package mockrepository is a generated GoMock package.
package mockrepository

import (
	context "context"
	reflect "reflect"

	model "github.com/traPtitech/rucQ/model"
	gomock "go.uber.org/mock/gomock"
)

// MockPaymentRepository is a mock of PaymentRepository interface.
type MockPaymentRepository struct {
	ctrl     *gomock.Controller
	recorder *MockPaymentRepositoryMockRecorder
	isgomock struct{}
}

// MockPaymentRepositoryMockRecorder is the mock recorder for MockPaymentRepository.
type MockPaymentRepositoryMockRecorder struct {
	mock *MockPaymentRepository
}

// NewMockPaymentRepository creates a new mock instance.
func NewMockPaymentRepository(ctrl *gomock.Controller) *MockPaymentRepository {
	mock := &MockPaymentRepository{ctrl: ctrl}
	mock.recorder = &MockPaymentRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPaymentRepository) EXPECT() *MockPaymentRepositoryMockRecorder {
	return m.recorder
}

// CreatePayment mocks base method.
func (m *MockPaymentRepository) CreatePayment(ctx context.Context, payment *model.Payment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePayment", ctx, payment)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePayment indicates an expected call of CreatePayment.
func (mr *MockPaymentRepositoryMockRecorder) CreatePayment(ctx, payment any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePayment", reflect.TypeOf((*MockPaymentRepository)(nil).CreatePayment), ctx, payment)
}

// GetPayments mocks base method.
func (m *MockPaymentRepository) GetPayments(ctx context.Context, campID uint) ([]model.Payment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPayments", ctx, campID)
	ret0, _ := ret[0].([]model.Payment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPayments indicates an expected call of GetPayments.
func (mr *MockPaymentRepositoryMockRecorder) GetPayments(ctx, campID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayments", reflect.TypeOf((*MockPaymentRepository)(nil).GetPayments), ctx, campID)
}

// UpdatePayment mocks base method.
func (m *MockPaymentRepository) UpdatePayment(ctx context.Context, paymentID uint, payment *model.Payment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePayment", ctx, paymentID, payment)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePayment indicates an expected call of UpdatePayment.
func (mr *MockPaymentRepositoryMockRecorder) UpdatePayment(ctx, paymentID, payment any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePayment", reflect.TypeOf((*MockPaymentRepository)(nil).UpdatePayment), ctx, paymentID, payment)
}
