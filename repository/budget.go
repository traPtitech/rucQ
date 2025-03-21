package repository

import (
	"fmt"

	"github.com/traP-jp/rucQ/backend/model"
)

func (r *Repository) GetBudget(traqID string) (*model.Budget, error) {
	defaultCamp, err := r.GetDefaultCamp()

	if err != nil {
		return nil, err
	}

	user, err := r.GetOrCreateUser(traqID)

	if err != nil {
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}

	var budget model.Budget

	if err := r.db.
		Where(&model.Budget{
			UserID: user.ID,
			CampID: defaultCamp.ID,
		}).
		FirstOrCreate(&budget).
		Error; err != nil {
		return nil, err
	}

	return &budget, nil
}

func (r *Repository) GetBudgets(campID uint) ([]model.Budget, error) {
	var budgets []model.Budget

	if err := r.db.
		Where(&model.Budget{
			CampID: campID,
		}).
		Find(&budgets).
		Error; err != nil {
		return nil, fmt.Errorf("failed to get budgets: %w", err)
	}

	return budgets, nil
}

func (r *Repository) UpdateBudget(budget *model.Budget) error {
	return r.db.Save(budget).Error
}
