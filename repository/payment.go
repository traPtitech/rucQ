//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *model.Payment) error
	GetPayments(ctx context.Context, campID uint) ([]model.Payment, error)
	UpdatePayment(ctx context.Context, paymentID uint, payment *model.Payment) error
}
