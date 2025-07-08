//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *model.Payment) error
	GetPayments(ctx context.Context) ([]model.Payment, error)
	UpdatePayment(ctx context.Context, paymentID uint, payment *model.Payment) error
}
