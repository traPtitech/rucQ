package gormrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) GetRooms() ([]model.Room, error) {
	var rooms []model.Room

	if err := r.db.Preload("Members").Find(&rooms).Error; err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *Repository) GetRoomByID(ctx context.Context, roomID uint) (*model.Room, error) {
	room, err := gorm.G[model.Room](r.db).
		Preload("Members", nil).
		Where("id = ?", roomID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRoomNotFound
		}

		return nil, err
	}

	return &room, nil
}

func (r *Repository) GetRoomByUserID(
	ctx context.Context,
	campID uint,
	userID string,
) (*model.Room, error) {
	var room model.Room

	if err := r.db.
		WithContext(ctx).
		Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Joins("JOIN room_groups ON room_groups.id = rooms.room_group_id").
		Where("room_members.user_id = ?", userID).
		Where("room_groups.camp_id = ?", campID).
		Preload("Members").
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRoomNotFound
		}

		return nil, err
	}

	return &room, nil
}

func (r *Repository) CreateRoom(ctx context.Context, room *model.Room) error {
	if err := r.db.
		WithContext(ctx).
		Omit("Members.*"). // 関係は更新するがユーザーの新規作成はされないようにする
		Create(room).
		Error; err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return repository.ErrUserOrRoomGroupNotFound
		}

		return err
	}

	return nil
}

func (r *Repository) UpdateRoom(ctx context.Context, roomID uint, room *model.Room) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var campID uint
		err := tx.Table("rooms").
			Select("room_groups.camp_id").
			Joins("JOIN room_groups ON room_groups.id = rooms.room_group_id").
			Where("rooms.id = ?", roomID).
			Row().Scan(&campID)

		if err != nil {
			// そもそも部屋が存在しない場合
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrRoomNotFound
			}
			return err
		}

		// 重複チェック
		if len(room.Members) > 0 {
			// 登録予定のメンバーIDのリストを作成
			var newUserIDs []string
			for _, m := range room.Members {
				newUserIDs = append(newUserIDs, m.ID)
			}

			var count int64
			// 合宿全体の中で、自分以外の部屋に所属しているメンバーがいないか数える
			err = tx.Table("room_members").
				Joins("JOIN rooms ON rooms.id = room_members.room_id").
				Joins("JOIN room_groups ON room_groups.id = rooms.room_group_id").
				Where("room_groups.camp_id = ?", campID).       // 同じ合宿内
				Where("rooms.id <> ?", roomID).                 // 自分以外の部屋
				Where("room_members.user_id IN ?", newUserIDs). // 登録予定の誰か
				Count(&count).Error

			if err != nil {
				return err
			}

			// 1人でも見つかれば重複エラーとする
			if count > 0 {
				return repository.ErrUserAlreadyAssigned
			}
		}

		room.ID = roomID

		rowsAffected, err := gorm.G[*model.Room](tx).Omit("Members").Updates(ctx, room)

		if err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				return repository.ErrRoomGroupNotFound
			}

			return err
		}

		if rowsAffected == 0 {
			return repository.ErrRoomNotFound
		}

		if err := tx.WithContext(ctx).
			Model(room).
			Omit("Members.*"). // ユーザーの新規作成はされないようにする
			Association("Members").
			Replace(room.Members); err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				return repository.ErrUserNotFound
			}

			return err
		}

		return nil
	})
}

func (r *Repository) DeleteRoom(ctx context.Context, roomID uint) error {
	rowsAffected, err := gorm.G[model.Room](r.db).
		Where("id = ?", roomID).
		Delete(ctx)

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrRoomNotFound
	}

	return nil
}
