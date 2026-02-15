package converter

import (
	"errors"

	"github.com/jinzhu/copier"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	activityservice "github.com/traPtitech/rucQ/service/activity"
)

var activityResponseToSchema = copier.TypeConverter{
	SrcType: activityservice.ActivityResponse{},
	DstType: api.ActivityResponse{},
	Fn: func(src any) (any, error) {
		activity, ok := src.(activityservice.ActivityResponse)
		if !ok {
			return nil, errors.New("src is not an activityservice.ActivityResponse")
		}

		var dst api.ActivityResponse

		switch activity.Type {
		case model.ActivityTypeRoomCreated:
			err := dst.FromRoomCreatedActivity(api.RoomCreatedActivity{
				Id:   int(activity.ID),
				Type: api.RoomCreated,
				Time: activity.Time,
			})
			if err != nil {
				return nil, err
			}

		case model.ActivityTypePaymentCreated:
			if activity.PaymentCreated == nil {
				return nil, errors.New("PaymentCreated detail is nil")
			}

			err := dst.FromPaymentCreatedActivity(api.PaymentCreatedActivity{
				Id:     int(activity.ID),
				Type:   api.PaymentCreated,
				Time:   activity.Time,
				Amount: activity.PaymentCreated.Amount,
			})
			if err != nil {
				return nil, err
			}

		case model.ActivityTypePaymentAmountChanged:
			if activity.PaymentAmountChanged == nil {
				return nil, errors.New("PaymentAmountChanged detail is nil")
			}

			err := dst.FromPaymentAmountChangedActivity(api.PaymentAmountChangedActivity{
				Id:     int(activity.ID),
				Type:   api.PaymentAmountChanged,
				Time:   activity.Time,
				Amount: activity.PaymentAmountChanged.Amount,
			})
			if err != nil {
				return nil, err
			}

		case model.ActivityTypePaymentPaidChanged:
			if activity.PaymentPaidChanged == nil {
				return nil, errors.New("PaymentPaidChanged detail is nil")
			}

			err := dst.FromPaymentPaidChangedActivity(api.PaymentPaidChangedActivity{
				Id:     int(activity.ID),
				Type:   api.PaymentPaidChanged,
				Time:   activity.Time,
				Amount: activity.PaymentPaidChanged.Amount,
			})
			if err != nil {
				return nil, err
			}

		case model.ActivityTypeRollCallCreated:
			if activity.RollCallCreated == nil {
				return nil, errors.New("RollCallCreated detail is nil")
			}

			err := dst.FromRollCallCreatedActivity(api.RollCallCreatedActivity{
				Id:         int(activity.ID),
				Type:       api.RollCallCreated,
				Time:       activity.Time,
				RollcallId: int(activity.RollCallCreated.RollCallID),
				Name:       activity.RollCallCreated.Name,
				IsSubject:  activity.RollCallCreated.IsSubject,
				Answered:   activity.RollCallCreated.Answered,
			})
			if err != nil {
				return nil, err
			}

		case model.ActivityTypeQuestionCreated:
			if activity.QuestionCreated == nil {
				return nil, errors.New("QuestionCreated detail is nil")
			}

			err := dst.FromQuestionCreatedActivity(api.QuestionCreatedActivity{
				Id:              int(activity.ID),
				Type:            api.QuestionCreated,
				Time:            activity.Time,
				QuestionGroupId: int(activity.QuestionCreated.QuestionGroupID),
				Name:            activity.QuestionCreated.Name,
				Due:             activity.QuestionCreated.Due,
				NeedsResponse:   activity.QuestionCreated.NeedsResponse,
			})
			if err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("unknown activity type: " + string(activity.Type))
		}

		return dst, nil
	},
}
