package gormrepository

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
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

func TestUpdateCamp(t *testing.T) {
	t.Parallel()

	t.Run("成功", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)

		// 更新するデータを準備
		newName := random.AlphaNumericString(t, 30)
		newGuidebook := random.AlphaNumericString(t, 200)
		newIsDraft := !camp.IsDraft
		newIsPaymentOpen := !camp.IsPaymentOpen
		newIsRegistrationOpen := !camp.IsRegistrationOpen
		newDateStart := random.Time(t)
		newDateEnd := newDateStart.Add(time.Duration(random.PositiveInt(t)))

		updatedCamp := model.Camp{
			DisplayID:          camp.DisplayID, // DisplayIDは更新しない
			Name:               newName,
			Guidebook:          newGuidebook,
			IsDraft:            newIsDraft,
			IsPaymentOpen:      newIsPaymentOpen,
			IsRegistrationOpen: newIsRegistrationOpen,
			DateStart:          newDateStart,
			DateEnd:            newDateEnd,
		}

		err := r.UpdateCamp(t.Context(), camp.ID, &updatedCamp)
		require.NoError(t, err)

		// 更新後のデータを取得して確認
		retrievedCamp, err := r.GetCampByID(t.Context(), camp.ID)
		require.NoError(t, err)

		assert.Equal(t, camp.ID, retrievedCamp.ID)
		assert.Equal(t, camp.DisplayID, retrievedCamp.DisplayID)
		assert.Equal(t, newName, retrievedCamp.Name)
		assert.Equal(t, newGuidebook, retrievedCamp.Guidebook)
		assert.Equal(t, newIsDraft, retrievedCamp.IsDraft)
		assert.Equal(t, newIsPaymentOpen, retrievedCamp.IsPaymentOpen)
		assert.Equal(t, newIsRegistrationOpen, retrievedCamp.IsRegistrationOpen)
		// 時刻の比較は秒単位で行う（MySQLの時刻精度の問題を回避）
		assert.WithinDuration(t, newDateStart, retrievedCamp.DateStart, time.Second)
		assert.WithinDuration(t, newDateEnd, retrievedCamp.DateEnd, time.Second)
	})

	t.Run("部分更新", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)

		// 名前だけを更新するためのマップを使用
		// Select("*")を使用しているため、構造体の場合はゼロ値も更新される
		err := r.db.Model(&model.Camp{}).
			Where("id = ?", camp.ID).
			Update("name", random.AlphaNumericString(t, 25)).Error
		require.NoError(t, err)

		// 更新後のデータを取得して確認
		retrievedCamp, err := r.GetCampByID(t.Context(), camp.ID)
		require.NoError(t, err)

		// 他のフィールドは元のまま
		assert.Equal(t, camp.DisplayID, retrievedCamp.DisplayID)
		assert.Equal(t, camp.Guidebook, retrievedCamp.Guidebook)
		assert.Equal(t, camp.IsDraft, retrievedCamp.IsDraft)
		assert.Equal(t, camp.IsPaymentOpen, retrievedCamp.IsPaymentOpen)
		assert.Equal(t, camp.IsRegistrationOpen, retrievedCamp.IsRegistrationOpen)
		assert.WithinDuration(t, camp.DateStart, retrievedCamp.DateStart, time.Second)
		assert.WithinDuration(t, camp.DateEnd, retrievedCamp.DateEnd, time.Second)
	})

	t.Run("ゼロ値での更新", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)

		// Select("*")を使用しているため、ゼロ値でも更新される
		zeroValueUpdate := model.Camp{
			Name:               "", // 空文字列
			Guidebook:          "", // 空文字列
			IsDraft:            false,
			IsPaymentOpen:      false,
			IsRegistrationOpen: false,
		}

		err := r.UpdateCamp(t.Context(), camp.ID, &zeroValueUpdate)
		require.NoError(t, err)

		// 更新後のデータを取得して確認
		retrievedCamp, err := r.GetCampByID(t.Context(), camp.ID)
		require.NoError(t, err)

		// ゼロ値が設定されていることを確認
		assert.Equal(t, "", retrievedCamp.Name)
		assert.Equal(t, "", retrievedCamp.Guidebook)
		assert.False(t, retrievedCamp.IsDraft)
		assert.False(t, retrievedCamp.IsPaymentOpen)
		assert.False(t, retrievedCamp.IsRegistrationOpen)
	})

	t.Run("存在しない合宿での更新はエラーになる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		nonExistentID := uint(random.PositiveInt(t))

		camp := model.Camp{
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          random.Time(t),
			DateEnd:            random.Time(t),
		}

		err := r.UpdateCamp(t.Context(), nonExistentID, &camp)
		assert.Error(t, err)
		assert.Equal(t, model.ErrNotFound, err)
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
		assert.Equal(t, repository.ErrCampNotFound, err)
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

		// ユーザーを合宿に参加させる
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

func TestAddCampParticipant(t *testing.T) {
	t.Parallel()

	t.Run("成功", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// ユーザーを参加者に追加
		err := r.AddCampParticipant(t.Context(), camp.ID, &user)

		// 参加受付が開いているかどうかに関わらず成功する
		// (repository層では参加受付の状態を確認しないため)
		if assert.NoError(t, err) {
			// 参加者が追加されていることを確認
			isParticipant, err := r.IsCampParticipant(t.Context(), camp.ID, user.ID)

			if assert.NoError(t, err) {
				assert.True(t, isParticipant)
			}

			// 参加者リストからも確認
			participants, err := r.GetCampParticipants(t.Context(), camp.ID)

			if assert.NoError(t, err) {
				if assert.Len(t, participants, 1) {
					assert.Equal(t, user.ID, participants[0].ID)
				}
			}

		}
	})

	t.Run("エラー: 存在しない合宿ID", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)

		// 存在しない合宿IDを使用
		nonExistentCampID := uint(random.PositiveInt(t))

		err := r.AddCampParticipant(t.Context(), nonExistentCampID, &user)
		assert.Error(t, err)
		assert.Equal(t, model.ErrNotFound, err)
	})

	t.Run("成功: 同じユーザーを複数回追加しても重複しない", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 参加受付を開く
		camp.IsRegistrationOpen = true
		err := r.UpdateCamp(t.Context(), camp.ID, &camp)
		require.NoError(t, err)

		// 同じユーザーを2回追加
		err = r.AddCampParticipant(t.Context(), camp.ID, &user)
		assert.NoError(t, err)

		err = r.AddCampParticipant(t.Context(), camp.ID, &user)
		assert.NoError(t, err) // 重複追加でもエラーにならない

		// 参加者リストに1つだけ追加されていることを確認
		participants, err := r.GetCampParticipants(t.Context(), camp.ID)
		assert.NoError(t, err)
		assert.Len(t, participants, 1)
		assert.Equal(t, user.ID, participants[0].ID)
	})

	t.Run("成功: 複数のユーザーを追加", func(t *testing.T) {
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

		// 複数のユーザーを追加
		err = r.AddCampParticipant(t.Context(), camp.ID, &user1)
		assert.NoError(t, err)

		err = r.AddCampParticipant(t.Context(), camp.ID, &user2)
		assert.NoError(t, err)

		err = r.AddCampParticipant(t.Context(), camp.ID, &user3)
		assert.NoError(t, err)

		// 全員が参加者として追加されていることを確認
		participants, err := r.GetCampParticipants(t.Context(), camp.ID)
		assert.NoError(t, err)
		assert.Len(t, participants, 3)

		participantIDs := make([]string, 3)
		for i, p := range participants {
			participantIDs[i] = p.ID
		}

		assert.Contains(t, participantIDs, user1.ID)
		assert.Contains(t, participantIDs, user2.ID)
		assert.Contains(t, participantIDs, user3.ID)
	})
}

func TestRepository_RemoveCampParticipant(t *testing.T) {
	t.Parallel()

	t.Run("参加者を正常に削除できる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 参加者を追加
		err := r.AddCampParticipant(t.Context(), camp.ID, &user)

		require.NoError(t, err)

		// 参加者が追加されていることを確認
		isParticipant, err := r.IsCampParticipant(t.Context(), camp.ID, user.ID)

		require.NoError(t, err)
		require.True(t, isParticipant)

		// 参加者を削除
		err = r.RemoveCampParticipant(t.Context(), camp.ID, &user)

		assert.NoError(t, err)

		// 参加者が削除されていることを確認
		isParticipant, err = r.IsCampParticipant(t.Context(), camp.ID, user.ID)
		require.NoError(t, err)
		assert.False(t, isParticipant)
	})

	t.Run("存在しない合宿の場合はErrCampNotFoundを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		user := mustCreateUser(t, r)

		// 存在しない合宿のIDを指定
		err := r.RemoveCampParticipant(t.Context(), uint(random.PositiveInt(t)), &user)

		if assert.Error(t, err) {
			assert.Equal(t, repository.ErrCampNotFound, err)
		}
	})

	t.Run("参加していないユーザーの場合はErrParticipantNotFoundを返す", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)

		// 参加していないユーザーを削除しようとする
		err := r.RemoveCampParticipant(t.Context(), camp.ID, &user)

		if assert.Error(t, err) {
			assert.Equal(t, repository.ErrParticipantNotFound, err)
		}
	})

	t.Run("複数の参加者がいる場合に特定のユーザーだけを削除できる", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user1 := mustCreateUser(t, r)
		user2 := mustCreateUser(t, r)
		user3 := mustCreateUser(t, r)

		// 複数のユーザーを追加
		err := r.AddCampParticipant(t.Context(), camp.ID, &user1)
		require.NoError(t, err)

		err = r.AddCampParticipant(t.Context(), camp.ID, &user2)
		require.NoError(t, err)

		err = r.AddCampParticipant(t.Context(), camp.ID, &user3)
		require.NoError(t, err)

		// user2だけを削除
		err = r.RemoveCampParticipant(t.Context(), camp.ID, &user2)
		assert.NoError(t, err)

		// user1とuser3は残っていることを確認
		isParticipant1, err := r.IsCampParticipant(t.Context(), camp.ID, user1.ID)
		require.NoError(t, err)
		assert.True(t, isParticipant1)

		isParticipant3, err := r.IsCampParticipant(t.Context(), camp.ID, user3.ID)
		require.NoError(t, err)
		assert.True(t, isParticipant3)

		// user2は削除されていることを確認
		isParticipant2, err := r.IsCampParticipant(t.Context(), camp.ID, user2.ID)
		require.NoError(t, err)
		assert.False(t, isParticipant2)
	})
}
