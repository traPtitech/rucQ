//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"
	"errors"

	"github.com/traPtitech/rucQ/model"
)

var ErrPaymentNotFound = errors.New("payment not found")

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *model.Payment) error
	GetPayments(ctx context.Context, campID uint) ([]model.Payment, error)
	GetPaymentByUserID(ctx context.Context, userID string) (*model.Payment, error)
	UpdatePayment(ctx context.Context, paymentID uint, payment *model.Payment) error
}
