package activity

import (
	"context"
	"errors"
	"slices"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

type activityServiceImpl struct {
	repo repository.Repository
}

func NewActivityService(repo repository.Repository) *activityServiceImpl {
	return &activityServiceImpl{repo: repo}
}

func (s *activityServiceImpl) RecordRoomCreated(
	ctx context.Context,
	room model.Room,
) error {
	roomGroup, err := s.repo.GetRoomGroupByID(ctx, room.RoomGroupID)
	if err != nil {
		return err
	}

	activity := &model.Activity{
		Type:        model.ActivityTypeRoomCreated,
		CampID:      roomGroup.CampID,
		ReferenceID: room.ID,
	}
	return s.repo.CreateActivity(ctx, activity)
}

func (s *activityServiceImpl) RecordPaymentCreated(
	ctx context.Context,
	payment model.Payment,
) error {
	activity := &model.Activity{
		Type:        model.ActivityTypePaymentCreated,
		CampID:      payment.CampID,
		UserID:      &payment.UserID,
		ReferenceID: payment.ID,
		Amount:      &payment.Amount,
	}
	return s.repo.CreateActivity(ctx, activity)
}

func (s *activityServiceImpl) RecordPaymentAmountChanged(
	ctx context.Context,
	payment model.Payment,
) error {
	activity := &model.Activity{
		Type:        model.ActivityTypePaymentAmountChanged,
		CampID:      payment.CampID,
		UserID:      &payment.UserID,
		ReferenceID: payment.ID,
		Amount:      &payment.Amount,
	}
	return s.repo.CreateActivity(ctx, activity)
}

func (s *activityServiceImpl) RecordPaymentPaidChanged(
	ctx context.Context,
	payment model.Payment,
) error {
	activity := &model.Activity{
		Type:        model.ActivityTypePaymentPaidChanged,
		CampID:      payment.CampID,
		UserID:      &payment.UserID,
		ReferenceID: payment.ID,
		Amount:      &payment.AmountPaid,
	}
	return s.repo.CreateActivity(ctx, activity)
}

func (s *activityServiceImpl) RecordRollCallCreated(
	ctx context.Context,
	rollCall model.RollCall,
) error {
	activity := &model.Activity{
		Type:        model.ActivityTypeRollCallCreated,
		CampID:      rollCall.CampID,
		ReferenceID: rollCall.ID,
	}
	return s.repo.CreateActivity(ctx, activity)
}

func (s *activityServiceImpl) RecordQuestionCreated(
	ctx context.Context,
	questionGroup model.QuestionGroup,
) error {
	activity := &model.Activity{
		Type:        model.ActivityTypeQuestionCreated,
		CampID:      questionGroup.CampID,
		ReferenceID: questionGroup.ID,
	}
	return s.repo.CreateActivity(ctx, activity)
}

// 部屋はユーザーごとに1つ、Paymentの変更は支払い金額の設定と入金確認で最低2回は
// 発生するため、たまに返金処理などが起こることも考慮して5件以内には収まると想定
const estimatedUserSpecificActivitiesCount = 5

func (s *activityServiceImpl) GetActivities(
	ctx context.Context,
	campID uint,
	userID string,
) ([]ActivityResponse, error) {
	activities, err := s.repo.GetActivitiesByCampID(ctx, campID)
	if err != nil {
		return nil, err
	}

	if len(activities) == 0 {
		return []ActivityResponse{}, nil
	}

	// ユーザーの部屋を取得（room_created のフィルタリング用）
	userRoom, err := s.repo.GetRoomByUserID(ctx, campID, userID)
	if err != nil && !errors.Is(err, repository.ErrRoomNotFound) {
		return nil, err
	}

	// 点呼情報を取得（roll_call_created の付加情報用）
	rollCalls, err := s.repo.GetRollCalls(ctx, campID)
	if err != nil && !errors.Is(err, repository.ErrCampNotFound) {
		return nil, err
	}

	rollCallMap := make(map[uint]model.RollCall, len(rollCalls))
	for _, rc := range rollCalls {
		rollCallMap[rc.ID] = rc
	}

	// 質問グループ情報を取得（question_created の付加情報用）
	questionGroups, err := s.repo.GetQuestionGroups(ctx, campID)
	if err != nil {
		return nil, err
	}

	questionGroupMap := make(map[uint]model.QuestionGroup, len(questionGroups))
	for _, qg := range questionGroups {
		questionGroupMap[qg.ID] = qg
	}

	// ユーザーの回答を取得（needsResponse 判定用）
	answers, err := s.repo.GetAnswers(ctx, repository.GetAnswersQuery{
		UserID:                &userID,
		IncludePrivateAnswers: true,
	})
	if err != nil {
		return nil, err
	}

	// QuestionID → bool のマップを作る
	answeredQuestionIDs := make(map[uint]bool, len(answers))
	for _, a := range answers {
		answeredQuestionIDs[a.QuestionID] = true
	}

	// 点呼など全体に影響するActivityとユーザー固有のActivityを合わせて要素数を見積もる
	estimatedActivitiesCount := len(
		rollCalls,
	) + len(
		questionGroups,
	) + estimatedUserSpecificActivitiesCount
	result := make([]ActivityResponse, 0, estimatedActivitiesCount)

	for _, a := range activities {
		switch a.Type {
		case model.ActivityTypeRoomCreated:
			if userRoom == nil || userRoom.ID != a.ReferenceID {
				continue
			}

			result = append(result, ActivityResponse{
				ID:          a.ID,
				Type:        a.Type,
				Time:        a.CreatedAt,
				RoomCreated: &RoomCreatedDetail{},
			})

		case model.ActivityTypePaymentCreated,
			model.ActivityTypePaymentAmountChanged,
			model.ActivityTypePaymentPaidChanged:
			if a.UserID == nil || *a.UserID != userID {
				continue
			}

			if a.Amount == nil {
				continue
			}

			resp := ActivityResponse{
				ID:   a.ID,
				Type: a.Type,
				Time: a.CreatedAt,
			}

			switch a.Type {
			case model.ActivityTypePaymentCreated:
				resp.PaymentCreated = &PaymentCreatedDetail{Amount: *a.Amount}
			case model.ActivityTypePaymentAmountChanged:
				resp.PaymentAmountChanged = &PaymentChangedDetail{Amount: *a.Amount}
			default:
				resp.PaymentPaidChanged = &PaymentChangedDetail{Amount: *a.Amount}
			}

			result = append(result, resp)

		case model.ActivityTypeRollCallCreated:
			rc, ok := rollCallMap[a.ReferenceID]
			if !ok {
				continue
			}

			isSubject := slices.ContainsFunc(rc.Subjects, func(u model.User) bool {
				return u.ID == userID
			})

			hasReaction := slices.ContainsFunc(rc.Reactions, func(r model.RollCallReaction) bool {
				return r.UserID == userID
			})
			answered := hasReaction

			result = append(result, ActivityResponse{
				ID:   a.ID,
				Type: a.Type,
				Time: a.CreatedAt,
				RollCallCreated: &RollCallCreatedDetail{
					RollCallID: rc.ID,
					Name:       rc.Name,
					IsSubject:  isSubject,
					Answered:   answered,
				},
			})

		case model.ActivityTypeQuestionCreated:
			qg, ok := questionGroupMap[a.ReferenceID]
			if !ok {
				continue
			}

			// IsRequired な質問で未回答のものがあるか
			needsResponse := slices.ContainsFunc(qg.Questions, func(q model.Question) bool {
				return q.IsRequired && !answeredQuestionIDs[q.ID]
			})

			result = append(result, ActivityResponse{
				ID:   a.ID,
				Type: a.Type,
				Time: a.CreatedAt,
				QuestionCreated: &QuestionCreatedDetail{
					QuestionGroupID: qg.ID,
					Name:            qg.Name,
					Due:             qg.Due,
					NeedsResponse:   needsResponse,
				},
			})
		}
	}

	return result, nil
}
