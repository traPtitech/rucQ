package gormrepository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestCreateActivity(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp := mustCreateCamp(t, r)
		user := mustCreateUser(t, r)
		userID := user.ID
		referenceID := uint(random.PositiveInt(t))
		activity := model.Activity{
			Type:        model.ActivityTypePaymentAmountChanged,
			CampID:      camp.ID,
			UserID:      &userID,
			ReferenceID: referenceID,
		}

		err := r.CreateActivity(t.Context(), &activity)

		assert.NoError(t, err)
		assert.NotZero(t, activity.ID)
		assert.Equal(t, model.ActivityTypePaymentAmountChanged, activity.Type)
		assert.Equal(t, camp.ID, activity.CampID)
		if assert.NotNil(t, activity.UserID) {
			assert.Equal(t, userID, *activity.UserID)
		}
		assert.Equal(t, referenceID, activity.ReferenceID)
	})
}

func TestGetActivitiesByCampID(t *testing.T) {
	t.Parallel()

	t.Run("Success - OrderByCreatedAtDesc", func(t *testing.T) {
		t.Parallel()

		r := setup(t)
		camp1 := mustCreateCamp(t, r)
		camp2 := mustCreateCamp(t, r)

		activityOld := model.Activity{
			Type:        model.ActivityTypeRoomCreated,
			CampID:      camp1.ID,
			ReferenceID: uint(random.PositiveInt(t)),
		}
		activityNew := model.Activity{
			Type:        model.ActivityTypeRollCallCreated,
			CampID:      camp1.ID,
			ReferenceID: uint(random.PositiveInt(t)),
		}
		activityOtherCamp := model.Activity{
			Type:        model.ActivityTypeQuestionCreated,
			CampID:      camp2.ID,
			ReferenceID: uint(random.PositiveInt(t)),
		}

		require.NoError(t, r.CreateActivity(t.Context(), &activityOld))
		require.NoError(t, r.CreateActivity(t.Context(), &activityNew))
		require.NoError(t, r.CreateActivity(t.Context(), &activityOtherCamp))

		timeOld := random.Time(t)
		timeNew := timeOld.Add(10 * time.Minute)

		require.NoError(t, r.db.Model(&activityOld).Update("created_at", timeOld).Error)
		require.NoError(t, r.db.Model(&activityNew).Update("created_at", timeNew).Error)

		activities, err := r.GetActivitiesByCampID(t.Context(), camp1.ID)

		require.NoError(t, err)
		require.Len(t, activities, 2)
		assert.Equal(t, activityNew.ID, activities[0].ID)
		assert.Equal(t, activityOld.ID, activities[1].ID)
	})
}
