package converter

import (
	"errors"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	activityservice "github.com/traPtitech/rucQ/service/activity"
)

// ConvertActivityResponse はservice層のActivityResponseをapi.ActivityResponseに変換する。
// copierでは oneOf の変換ができないためカスタム実装。
func ConvertActivityResponse(src activityservice.ActivityResponse) (api.ActivityResponse, error) {
	var dst api.ActivityResponse

	switch src.Type {
	case model.ActivityTypeRoomCreated:
		err := dst.FromRoomCreatedActivity(api.RoomCreatedActivity{
			Id:   int(src.ID),
			Type: api.RoomCreated,
			Time: src.Time,
		})
		if err != nil {
			return dst, err
		}

	case model.ActivityTypePaymentAmountChanged:
		if src.PaymentAmountChanged == nil {
			return dst, errors.New("PaymentAmountChanged detail is nil")
		}

		err := dst.FromPaymentAmountChangedActivity(api.PaymentAmountChangedActivity{
			Id:     int(src.ID),
			Type:   api.PaymentAmountChanged,
			Time:   src.Time,
			Amount: src.PaymentAmountChanged.Amount,
		})
		if err != nil {
			return dst, err
		}

	case model.ActivityTypePaymentPaidChanged:
		if src.PaymentPaidChanged == nil {
			return dst, errors.New("PaymentPaidChanged detail is nil")
		}

		err := dst.FromPaymentPaidChangedActivity(api.PaymentPaidChangedActivity{
			Id:     int(src.ID),
			Type:   api.PaymentPaidChanged,
			Time:   src.Time,
			Amount: src.PaymentPaidChanged.Amount,
		})
		if err != nil {
			return dst, err
		}

	case model.ActivityTypeRollCallCreated:
		if src.RollCallCreated == nil {
			return dst, errors.New("RollCallCreated detail is nil")
		}

		err := dst.FromRollCallCreatedActivity(api.RollCallCreatedActivity{
			Id:         int(src.ID),
			Type:       api.RollCallCreated,
			Time:       src.Time,
			RollcallId: int(src.RollCallCreated.RollCallID),
			Name:       src.RollCallCreated.Name,
			IsSubject:  src.RollCallCreated.IsSubject,
			Answered:   src.RollCallCreated.Answered,
		})
		if err != nil {
			return dst, err
		}

	case model.ActivityTypeQuestionCreated:
		if src.QuestionCreated == nil {
			return dst, errors.New("QuestionCreated detail is nil")
		}

		err := dst.FromQuestionCreatedActivity(api.QuestionCreatedActivity{
			Id:              int(src.ID),
			Type:            api.QuestionCreated,
			Time:            src.Time,
			QuestionGroupId: int(src.QuestionCreated.QuestionGroupID),
			Name:            src.QuestionCreated.Name,
			Due:             src.QuestionCreated.Due,
			NeedsResponse:   src.QuestionCreated.NeedsResponse,
		})
		if err != nil {
			return dst, err
		}

	default:
		return dst, errors.New("unknown activity type: " + string(src.Type))
	}

	return dst, nil
}

// ConvertActivityResponses は複数のActivityResponseを変換する。
func ConvertActivityResponses(
	src []activityservice.ActivityResponse,
) ([]api.ActivityResponse, error) {
	dst := make([]api.ActivityResponse, 0, len(src))

	for _, s := range src {
		d, err := ConvertActivityResponse(s)
		if err != nil {
			return nil, err
		}

		dst = append(dst, d)
	}

	return dst, nil
}
