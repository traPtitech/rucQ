package activity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/repository/mockrepository"
	"github.com/traPtitech/rucQ/testutil/random"
)

type activityTestSetup struct {
	service *activityServiceImpl
	repo    *mockrepository.MockRepository
}

func setup(t *testing.T) *activityTestSetup {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := mockrepository.NewMockRepository(ctrl)
	service := NewActivityService(repo)

	return &activityTestSetup{
		service: service,
		repo:    repo,
	}
}

func TestActivityServiceImpl_RecordActivities(t *testing.T) {
	t.Parallel()

	t.Run("RecordRoomCreated", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		roomID := uint(random.PositiveInt(t))
		roomGroupID := uint(random.PositiveInt(t))
		campID := uint(random.PositiveInt(t))
		room := model.Room{
			Model:       gorm.Model{ID: roomID},
			RoomGroupID: roomGroupID,
		}

		s.repo.MockRoomGroupRepository.EXPECT().
			GetRoomGroupByID(ctx, roomGroupID).
			Return(&model.RoomGroup{Model: gorm.Model{ID: roomGroupID}, CampID: campID}, nil)

		s.repo.MockActivityRepository.EXPECT().
			CreateActivity(ctx, gomock.AssignableToTypeOf(&model.Activity{})).
			DoAndReturn(func(_ context.Context, activity *model.Activity) error {
				assert.Equal(t, model.ActivityTypeRoomCreated, activity.Type)
				assert.Equal(t, campID, activity.CampID)
				assert.Equal(t, roomID, activity.ReferenceID)
				assert.Nil(t, activity.UserID)
				return nil
			})

		err := s.service.RecordRoomCreated(ctx, room)

		assert.NoError(t, err)
	})

	t.Run("RecordPaymentAmountChanged", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		userID := random.AlphaNumericString(t, 32)
		campID := uint(random.PositiveInt(t))
		paymentID := uint(random.PositiveInt(t))
		payment := model.Payment{
			Model:  gorm.Model{ID: paymentID},
			UserID: userID,
			CampID: campID,
		}

		s.repo.MockActivityRepository.EXPECT().
			CreateActivity(ctx, gomock.AssignableToTypeOf(&model.Activity{})).
			DoAndReturn(func(_ context.Context, activity *model.Activity) error {
				assert.Equal(t, model.ActivityTypePaymentAmountChanged, activity.Type)
				assert.Equal(t, campID, activity.CampID)
				assert.Equal(t, paymentID, activity.ReferenceID)
				if assert.NotNil(t, activity.UserID) {
					assert.Equal(t, userID, *activity.UserID)
				}
				return nil
			})

		err := s.service.RecordPaymentAmountChanged(ctx, payment)

		assert.NoError(t, err)
	})

	t.Run("RecordPaymentCreated", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		userID := random.AlphaNumericString(t, 32)
		campID := uint(random.PositiveInt(t))
		paymentID := uint(random.PositiveInt(t))
		payment := model.Payment{
			Model:  gorm.Model{ID: paymentID},
			UserID: userID,
			CampID: campID,
		}

		s.repo.MockActivityRepository.EXPECT().
			CreateActivity(ctx, gomock.AssignableToTypeOf(&model.Activity{})).
			DoAndReturn(func(_ context.Context, activity *model.Activity) error {
				assert.Equal(t, model.ActivityTypePaymentCreated, activity.Type)
				assert.Equal(t, campID, activity.CampID)
				assert.Equal(t, paymentID, activity.ReferenceID)
				if assert.NotNil(t, activity.UserID) {
					assert.Equal(t, userID, *activity.UserID)
				}
				return nil
			})

		err := s.service.RecordPaymentCreated(ctx, payment)

		assert.NoError(t, err)
	})

	t.Run("RecordPaymentPaidChanged", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		userID := random.AlphaNumericString(t, 32)
		campID := uint(random.PositiveInt(t))
		paymentID := uint(random.PositiveInt(t))
		payment := model.Payment{
			Model:  gorm.Model{ID: paymentID},
			UserID: userID,
			CampID: campID,
		}

		s.repo.MockActivityRepository.EXPECT().
			CreateActivity(ctx, gomock.AssignableToTypeOf(&model.Activity{})).
			DoAndReturn(func(_ context.Context, activity *model.Activity) error {
				assert.Equal(t, model.ActivityTypePaymentPaidChanged, activity.Type)
				assert.Equal(t, campID, activity.CampID)
				assert.Equal(t, paymentID, activity.ReferenceID)
				if assert.NotNil(t, activity.UserID) {
					assert.Equal(t, userID, *activity.UserID)
				}
				return nil
			})

		err := s.service.RecordPaymentPaidChanged(ctx, payment)

		assert.NoError(t, err)
	})

	t.Run("RecordRollCallCreated", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		campID := uint(random.PositiveInt(t))
		rollCallID := uint(random.PositiveInt(t))
		rollCall := model.RollCall{
			Model:  gorm.Model{ID: rollCallID},
			CampID: campID,
		}

		s.repo.MockActivityRepository.EXPECT().
			CreateActivity(ctx, gomock.AssignableToTypeOf(&model.Activity{})).
			DoAndReturn(func(_ context.Context, activity *model.Activity) error {
				assert.Equal(t, model.ActivityTypeRollCallCreated, activity.Type)
				assert.Equal(t, campID, activity.CampID)
				assert.Equal(t, rollCallID, activity.ReferenceID)
				assert.Nil(t, activity.UserID)
				return nil
			})

		err := s.service.RecordRollCallCreated(ctx, rollCall)

		assert.NoError(t, err)
	})

	t.Run("RecordQuestionCreated", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		campID := uint(random.PositiveInt(t))
		questionGroupID := uint(random.PositiveInt(t))
		questionGroup := model.QuestionGroup{
			Model:  gorm.Model{ID: questionGroupID},
			CampID: campID,
		}

		s.repo.MockActivityRepository.EXPECT().
			CreateActivity(ctx, gomock.AssignableToTypeOf(&model.Activity{})).
			DoAndReturn(func(_ context.Context, activity *model.Activity) error {
				assert.Equal(t, model.ActivityTypeQuestionCreated, activity.Type)
				assert.Equal(t, campID, activity.CampID)
				assert.Equal(t, questionGroupID, activity.ReferenceID)
				assert.Nil(t, activity.UserID)
				return nil
			})

		err := s.service.RecordQuestionCreated(ctx, questionGroup)

		assert.NoError(t, err)
	})
}

func TestActivityServiceImpl_GetActivities(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		baseTime := random.Time(t)
		timePaymentCreated := baseTime.Add(6 * time.Minute)
		timePaymentAmount := baseTime.Add(5 * time.Minute)
		timeRollCall := baseTime.Add(4 * time.Minute)
		timeRoom := baseTime.Add(3 * time.Minute)
		timePaymentPaid := baseTime.Add(2 * time.Minute)
		timeQuestion := baseTime.Add(1 * time.Minute)

		roomID := uint(random.PositiveInt(t))
		userRoom := &model.Room{Model: gorm.Model{ID: roomID}}

		paymentAmountID := uint(random.PositiveInt(t))
		paymentPaidID := uint(random.PositiveInt(t))
		paymentCreatedID := uint(random.PositiveInt(t))
		paymentCreated := &model.Payment{
			Model:  gorm.Model{ID: paymentCreatedID},
			Amount: random.PositiveInt(t),
		}
		paymentAmount := &model.Payment{
			Model:  gorm.Model{ID: paymentAmountID},
			Amount: random.PositiveInt(t),
		}
		paymentPaid := &model.Payment{
			Model:  gorm.Model{ID: paymentPaidID},
			Amount: random.PositiveInt(t),
		}

		rollCallID := uint(random.PositiveInt(t))
		rollCallName := random.AlphaNumericString(t, 20)
		rollCall := model.RollCall{
			Model: gorm.Model{ID: rollCallID},
			Name:  rollCallName,
			Subjects: []model.User{
				{ID: userID},
			},
			Reactions: []model.RollCallReaction{
				{UserID: userID},
			},
		}

		questionGroupID := uint(random.PositiveInt(t))
		questionGroupName := random.AlphaNumericString(t, 20)
		requiredQuestionID := uint(random.PositiveInt(t))
		optionalQuestionID := uint(random.PositiveInt(t))
		questionGroup := model.QuestionGroup{
			Model: gorm.Model{ID: questionGroupID},
			Name:  questionGroupName,
			Due:   timeQuestion,
			Questions: []model.Question{
				{Model: gorm.Model{ID: requiredQuestionID}, IsRequired: true},
				{Model: gorm.Model{ID: optionalQuestionID}},
			},
		}

		answers := []model.Answer{
			{QuestionID: optionalQuestionID, UserID: userID},
		}

		activities := []model.Activity{
			{
				Model:       gorm.Model{ID: 1, CreatedAt: timeQuestion},
				Type:        model.ActivityTypeQuestionCreated,
				CampID:      campID,
				ReferenceID: questionGroupID,
			},
			{
				Model:       gorm.Model{ID: 2, CreatedAt: timeRoom},
				Type:        model.ActivityTypeRoomCreated,
				CampID:      campID,
				ReferenceID: roomID,
			},
			{
				Model:       gorm.Model{ID: 3, CreatedAt: timePaymentCreated},
				Type:        model.ActivityTypePaymentCreated,
				CampID:      campID,
				ReferenceID: paymentCreatedID,
				UserID:      &userID,
			},
			{
				Model:       gorm.Model{ID: 4, CreatedAt: timePaymentAmount},
				Type:        model.ActivityTypePaymentAmountChanged,
				CampID:      campID,
				ReferenceID: paymentAmountID,
				UserID:      &userID,
			},
			{
				Model:       gorm.Model{ID: 5, CreatedAt: timeRollCall},
				Type:        model.ActivityTypeRollCallCreated,
				CampID:      campID,
				ReferenceID: rollCallID,
			},
			{
				Model:       gorm.Model{ID: 6, CreatedAt: timePaymentPaid},
				Type:        model.ActivityTypePaymentPaidChanged,
				CampID:      campID,
				ReferenceID: paymentPaidID,
				UserID:      &userID,
			},
		}

		s.repo.MockActivityRepository.EXPECT().
			GetActivitiesByCampID(ctx, campID).
			Return(activities, nil)

		s.repo.MockRoomRepository.EXPECT().
			GetRoomByUserID(ctx, campID, userID).
			Return(userRoom, nil)

		s.repo.MockRollCallRepository.EXPECT().
			GetRollCalls(ctx, campID).
			Return([]model.RollCall{rollCall}, nil)

		s.repo.MockQuestionGroupRepository.EXPECT().
			GetQuestionGroups(ctx, campID).
			Return([]model.QuestionGroup{questionGroup}, nil)

		s.repo.MockAnswerRepository.EXPECT().
			GetAnswers(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, query repository.GetAnswersQuery) ([]model.Answer, error) {
				if assert.NotNil(t, query.UserID) {
					assert.Equal(t, userID, *query.UserID)
				}
				assert.True(t, query.IncludePrivateAnswers)
				return answers, nil
			})

		s.repo.MockPaymentRepository.EXPECT().
			GetPaymentByID(ctx, paymentCreatedID).
			Return(paymentCreated, nil)

		s.repo.MockPaymentRepository.EXPECT().
			GetPaymentByID(ctx, paymentAmountID).
			Return(paymentAmount, nil)

		s.repo.MockPaymentRepository.EXPECT().
			GetPaymentByID(ctx, paymentPaidID).
			Return(paymentPaid, nil)

		responses, err := s.service.GetActivities(ctx, campID, userID)

		require.NoError(t, err)
		require.Len(t, responses, 6)

		assert.Equal(t, model.ActivityTypePaymentCreated, responses[0].Type)
		assert.Equal(t, timePaymentCreated, responses[0].Time)
		if assert.NotNil(t, responses[0].PaymentCreated) {
			assert.Equal(t, paymentCreated.Amount, responses[0].PaymentCreated.Amount)
		}

		assert.Equal(t, model.ActivityTypePaymentAmountChanged, responses[1].Type)
		assert.Equal(t, timePaymentAmount, responses[1].Time)
		if assert.NotNil(t, responses[1].PaymentAmountChanged) {
			assert.Equal(t, paymentAmount.Amount, responses[1].PaymentAmountChanged.Amount)
		}

		assert.Equal(t, model.ActivityTypeRollCallCreated, responses[2].Type)
		assert.Equal(t, timeRollCall, responses[2].Time)
		if assert.NotNil(t, responses[2].RollCallCreated) {
			assert.Equal(t, rollCallID, responses[2].RollCallCreated.RollCallID)
			assert.Equal(t, rollCallName, responses[2].RollCallCreated.Name)
			assert.True(t, responses[2].RollCallCreated.IsSubject)
			assert.True(t, responses[2].RollCallCreated.Answered)
		}

		assert.Equal(t, model.ActivityTypeRoomCreated, responses[3].Type)
		assert.Equal(t, timeRoom, responses[3].Time)
		assert.NotNil(t, responses[3].RoomCreated)

		assert.Equal(t, model.ActivityTypePaymentPaidChanged, responses[4].Type)
		assert.Equal(t, timePaymentPaid, responses[4].Time)
		if assert.NotNil(t, responses[4].PaymentPaidChanged) {
			assert.Equal(t, paymentPaid.Amount, responses[4].PaymentPaidChanged.Amount)
		}

		assert.Equal(t, model.ActivityTypeQuestionCreated, responses[5].Type)
		assert.Equal(t, timeQuestion, responses[5].Time)
		if assert.NotNil(t, responses[5].QuestionCreated) {
			assert.Equal(t, questionGroupID, responses[5].QuestionCreated.QuestionGroupID)
			assert.Equal(t, questionGroupName, responses[5].QuestionCreated.Name)
			assert.Equal(t, timeQuestion, responses[5].QuestionCreated.Due)
			assert.True(t, responses[5].QuestionCreated.NeedsResponse)
		}
	})

	t.Run("Error (GetActivitiesByCampID)", func(t *testing.T) {
		t.Parallel()

		s := setup(t)
		ctx := t.Context()
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)

		s.repo.MockActivityRepository.EXPECT().
			GetActivitiesByCampID(ctx, campID).
			Return(nil, errors.New("db error"))

		responses, err := s.service.GetActivities(ctx, campID, userID)

		assert.Error(t, err)
		assert.Nil(t, responses)
	})
}
