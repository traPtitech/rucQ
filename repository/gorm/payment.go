package gorm

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreatePayment(ctx context.Context, payment *model.Payment) error {
	err := gorm.G[model.Payment](r.db).Create(ctx, payment)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetPayments(ctx context.Context, campID uint) ([]model.Payment, error) {
	payments, err := gorm.G[model.Payment](r.db).
		Where("camp_id = ?", campID).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *Repository) UpdatePayment(
	ctx context.Context,
	paymentID uint,
	payment *model.Payment,
) error {
	payment.ID = paymentID

	_, err := gorm.G[*model.Payment](r.db).Where("id = ?", paymentID).Updates(ctx, payment)
	if err != nil {
		return err
	}

	return nil
}
