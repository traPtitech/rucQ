package router

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/model"
	activityservice "github.com/traPtitech/rucQ/service/activity"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestServer_GetActivities(t *testing.T) {
	t.Parallel()

	t.Run("成功", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := &model.User{ID: userID}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(user, nil).
			Times(1)

		roomCreatedID := uint(random.PositiveInt(t))
		roomCreatedTime := random.Time(t)
		paymentCreatedID := uint(random.PositiveInt(t))
		paymentCreatedTime := random.Time(t)
		paymentAmountID := uint(random.PositiveInt(t))
		paymentAmountTime := random.Time(t)
		paymentPaidID := uint(random.PositiveInt(t))
		paymentPaidTime := random.Time(t)
		rollCallCreatedID := uint(random.PositiveInt(t))
		rollCallTime := random.Time(t)
		questionCreatedID := uint(random.PositiveInt(t))
		questionTime := random.Time(t)
		createdAmount := random.PositiveInt(t)
		amountChanged := random.PositiveInt(t)
		amountPaid := random.PositiveInt(t)
		rollCallID := uint(random.PositiveInt(t))
		rollCallName := random.AlphaNumericString(t, 20)
		questionGroupID := uint(random.PositiveInt(t))
		questionGroupName := random.AlphaNumericString(t, 20)

		activities := []activityservice.ActivityResponse{
			{
				ID:   roomCreatedID,
				Type: model.ActivityTypeRoomCreated,
				Time: roomCreatedTime,
			},
			{
				ID:   paymentCreatedID,
				Type: model.ActivityTypePaymentCreated,
				Time: paymentCreatedTime,
				PaymentCreated: &activityservice.PaymentCreatedDetail{
					Amount: createdAmount,
				},
			},
			{
				ID:   paymentAmountID,
				Type: model.ActivityTypePaymentAmountChanged,
				Time: paymentAmountTime,
				PaymentAmountChanged: &activityservice.PaymentChangedDetail{
					Amount: amountChanged,
				},
			},
			{
				ID:   paymentPaidID,
				Type: model.ActivityTypePaymentPaidChanged,
				Time: paymentPaidTime,
				PaymentPaidChanged: &activityservice.PaymentChangedDetail{
					Amount: amountPaid,
				},
			},
			{
				ID:   rollCallCreatedID,
				Type: model.ActivityTypeRollCallCreated,
				Time: rollCallTime,
				RollCallCreated: &activityservice.RollCallCreatedDetail{
					RollCallID: rollCallID,
					Name:       rollCallName,
					IsSubject:  true,
					Answered:   true,
				},
			},
			{
				ID:   questionCreatedID,
				Type: model.ActivityTypeQuestionCreated,
				Time: questionTime,
				QuestionCreated: &activityservice.QuestionCreatedDetail{
					QuestionGroupID: questionGroupID,
					Name:            questionGroupName,
					Due:             questionTime,
					NeedsResponse:   true,
				},
			},
		}

		h.activityService.EXPECT().
			GetActivities(gomock.Any(), campID, userID).
			Return(activities, nil).
			Times(1)

		res := h.expect.GET("/api/camps/{campId}/activities", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array()

		res.Length().IsEqual(6)

		// Check first activity (RoomCreated)
		act1 := res.Value(0).Object()
		act1.Value("id").Number().IsEqual(roomCreatedID)
		act1.Value("type").String().IsEqual("room_created")
		act1.Value("time").String().IsEqual(roomCreatedTime.Format(time.RFC3339Nano))

		// Check second activity (PaymentCreated)
		act2 := res.Value(1).Object()
		act2.Value("id").Number().IsEqual(paymentCreatedID)
		act2.Value("type").String().IsEqual("payment_created")
		act2.Value("time").String().IsEqual(paymentCreatedTime.Format(time.RFC3339Nano))
		act2.Value("amount").Number().IsEqual(createdAmount)

		// Check third activity (PaymentAmountChanged)
		act3 := res.Value(2).Object()
		act3.Value("id").Number().IsEqual(paymentAmountID)
		act3.Value("type").String().IsEqual("payment_amount_changed")
		act3.Value("time").String().IsEqual(paymentAmountTime.Format(time.RFC3339Nano))
		act3.Value("amount").Number().IsEqual(amountChanged)

		// Check fourth activity (PaymentPaidChanged)
		act4 := res.Value(3).Object()
		act4.Value("id").Number().IsEqual(paymentPaidID)
		act4.Value("type").String().IsEqual("payment_paid_changed")
		act4.Value("time").String().IsEqual(paymentPaidTime.Format(time.RFC3339Nano))
		act4.Value("amount").Number().IsEqual(amountPaid)

		// Check fifth activity (RollCallCreated)
		act5 := res.Value(4).Object()
		act5.Value("id").Number().IsEqual(rollCallCreatedID)
		act5.Value("type").String().IsEqual("roll_call_created")
		act5.Value("time").String().IsEqual(rollCallTime.Format(time.RFC3339Nano))
		act5.Value("rollcallId").Number().IsEqual(int(rollCallID))
		act5.Value("name").String().IsEqual(rollCallName)
		act5.Value("isSubject").Boolean().IsTrue()
		act5.Value("answered").Boolean().IsTrue()

		// Check sixth activity (QuestionCreated)
		act6 := res.Value(5).Object()
		act6.Value("id").Number().IsEqual(questionCreatedID)
		act6.Value("type").String().IsEqual("question_created")
		act6.Value("time").String().IsEqual(questionTime.Format(time.RFC3339Nano))
		act6.Value("questionGroupId").Number().IsEqual(int(questionGroupID))
		act6.Value("name").String().IsEqual(questionGroupName)
		act6.Value("due").String().IsEqual(questionTime.Format(time.RFC3339Nano))
		act6.Value("needsResponse").Boolean().IsTrue()
	})

	t.Run("activityServiceでエラーが起こった場合", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := uint(random.PositiveInt(t))
		userID := random.AlphaNumericString(t, 32)
		user := &model.User{ID: userID}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userID).
			Return(user, nil).
			Times(1)

		h.activityService.EXPECT().
			GetActivities(gomock.Any(), campID, userID).
			Return(nil, assert.AnError).
			Times(1)

		h.expect.GET("/api/camps/{campId}/activities", campID).
			WithHeader("X-Forwarded-User", userID).
			Expect().
			Status(http.StatusInternalServerError)
	})
}
