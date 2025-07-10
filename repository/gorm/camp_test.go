package gorm

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		err := r.UpdateCamp(t.Context(), camp.ID, &camp)
		require.NoError(t, err)

		err = r.AddCampParticipant(t.Context(), camp.ID, &user)
		require.NoError(t, err)

		isParticipant, err := r.IsCampParticipant(t.Context(), camp.ID, user.ID)
		assert.NoError(t, err)
		assert.True(t, isParticipant)
	})

	t.Run("参加者でない場合はfalseを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		isParticipant, err := r.IsCampParticipant(t.Context(), camp.ID, user.ID)
		assert.NoError(t, err)
		assert.False(t, isParticipant)
	})

	t.Run("存在しない合宿のIDを指定した場合はエラーを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)

		// 存在しない合宿のIDを指定
		isParticipant, err := r.IsCampParticipant(t.Context(), uint(random.PositiveInt(t)), user.ID)
		assert.Error(t, err)
		assert.Equal(t, model.ErrNotFound, err)
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
		err := r.UpdateCamp(t.Context(), camp.ID, &camp)
		require.NoError(t, err)

		// user1とuser3を参加者に追加
		err = r.AddCampParticipant(t.Context(), camp.ID, &user1)
		require.NoError(t, err)
		err = r.AddCampParticipant(t.Context(), camp.ID, &user3)
		require.NoError(t, err)

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

	t.Run("userIDの大文字・小文字が違う場合でもtrueを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 参加受付を開く
		camp.IsRegistrationOpen = true
		err := r.UpdateCamp(t.Context(), camp.ID, &camp)
		require.NoError(t, err)

		// ユーザーをキャンプに参加させる
		err = r.AddCampParticipant(t.Context(), camp.ID, &user)
		require.NoError(t, err)

		// 大文字・小文字を変更したIDで確認
		// 例: "abc123" -> "ABC123"
		wrongCaseUserID := strings.ToUpper(user.ID)
		if wrongCaseUserID == user.ID {
			// 全て大文字だった場合は小文字に変更
			wrongCaseUserID = strings.ToLower(user.ID)
		}

		isParticipantWrongCase, err := r.IsCampParticipant(t.Context(), camp.ID, wrongCaseUserID)
		assert.NoError(t, err)
		assert.True(t, isParticipantWrongCase, "大文字・小文字が異なるuserIDでも参加者として認識されるはず")
	})

	t.Run("userIDの部分的な大文字・小文字変更でもtrueを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 参加受付を開く
		camp.IsRegistrationOpen = true
		err := r.UpdateCamp(t.Context(), camp.ID, &camp)
		require.NoError(t, err)

		err = r.AddCampParticipant(t.Context(), camp.ID, &user)
		require.NoError(t, err)

		// userIDの一部分だけ大文字・小文字を変更
		// 英字が含まれる場合のみテスト
		if len(user.ID) > 0 {
			runes := []rune(user.ID)
			var modified bool
			for i, r := range runes {
				if r >= 'a' && r <= 'z' {
					runes[i] = r - 32 // 小文字を大文字に
					modified = true
					break
				} else if r >= 'A' && r <= 'Z' {
					runes[i] = r + 32 // 大文字を小文字に
					modified = true
					break
				}
			}

			if modified {
				partiallyModifiedUserID := string(runes)
				isParticipantPartial, err := r.IsCampParticipant(
					t.Context(),
					camp.ID,
					partiallyModifiedUserID,
				)
				assert.NoError(t, err)
				assert.True(t, isParticipantPartial, "部分的に大文字・小文字が異なるuserIDでも参加者として認識されるはず")
			}
		}
	})
}
