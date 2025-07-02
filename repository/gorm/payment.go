package gorm

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreatePayment(ctx context.Context, payment *model.Payment) error {
	err := gorm.G[*model.Payment](r.db).Create(ctx, &payment)

	if err != nil {
		return err
	}

	return nil
}
