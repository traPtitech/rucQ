package gorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestCreateCamp(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		camp := model.Camp{
			DisplayID:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          dateStart,
			DateEnd:            dateEnd,
		}
		err := r.CreateCamp(&camp)

		assert.NoError(t, err)
	})
}

func TestGetCamps(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp1 := mustCreateCamp(t, r)
		camp2 := mustCreateCamp(t, r)
		camps, err := r.GetCamps()

		assert.NoError(t, err)
		assert.Len(t, camps, 2)
		assert.Contains(t, camps, camp1)
		assert.Contains(t, camps, camp2)
	})
}

func TestIsCampParticipant(t *testing.T) {
	t.Parallel()

	t.Run("参加者の場合はtrueを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 参加受付を開く
		camp.IsRegistrationOpen = true
		err := r.UpdateCamp(camp.ID, &camp)
		assert.NoError(t, err)

		// ユーザーをキャンプに参加させる
		err = r.AddCampParticipant(t.Context(), camp.ID, &user)
		assert.NoError(t, err)

		// 参加者かどうかを確認
		isParticipant, err := r.IsCampParticipant(t.Context(), camp.ID, user.ID)
		assert.NoError(t, err)
		assert.True(t, isParticipant)
	})

	t.Run("参加者でない場合はfalseを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// ユーザーをキャンプに参加させない
		
		// 参加者かどうかを確認
		isParticipant, err := r.IsCampParticipant(t.Context(), camp.ID, user.ID)
		assert.NoError(t, err)
		assert.False(t, isParticipant)
	})

	t.Run("存在しないキャンプIDを指定した場合はfalseを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)

		// 存在しないキャンプIDを指定
		isParticipant, err := r.IsCampParticipant(t.Context(), 999999, user.ID)
		assert.NoError(t, err)
		assert.False(t, isParticipant)
	})

	t.Run("存在しないユーザーIDを指定した場合はfalseを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)

		// 存在しないユーザーIDを指定
		isParticipant, err := r.IsCampParticipant(t.Context(), camp.ID, "nonexistent-user")
		assert.NoError(t, err)
		assert.False(t, isParticipant)
	})

	t.Run("複数の参加者の中から特定のユーザーを見つけることができる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		user3 := mustCreateUser(t, r)

		// 参加受付を開く
		camp.IsRegistrationOpen = true
		err := r.UpdateCamp(camp.ID, &camp)
		assert.NoError(t, err)

		// user1とuser3を参加者に追加
		err = r.AddCampParticipant(t.Context(), camp.ID, &user1)
		assert.NoError(t, err)
		err = r.AddCampParticipant(t.Context(), camp.ID, &user3)
		assert.NoError(t, err)

		// 参加者の確認
		isParticipant1, err := r.IsCampParticipant(t.Context(), camp.ID, user1.ID)
		assert.NoError(t, err)
		assert.True(t, isParticipant1)

		isParticipant2, err := r.IsCampParticipant(t.Context(), camp.ID, user2.ID)
		assert.NoError(t, err)
		assert.False(t, isParticipant2)

		isParticipant3, err := r.IsCampParticipant(t.Context(), camp.ID, user3.ID)
		assert.NoError(t, err)
		assert.True(t, isParticipant3)
	})
}
