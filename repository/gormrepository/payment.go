package gormrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
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

func (r *Repository) GetPaymentByUserID(
	ctx context.Context,
	userID string,
) (*model.Payment, error) {
	payment, err := gorm.G[model.Payment](r.db).
		Where("user_id = ?", userID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrPaymentNotFound
		}

		return nil, err
	}

	return &payment, nil
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
