package activity

import (
	"context"
	"errors"
	"slices"
	"sort"

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
	activity := &model.Activity{
		Type:        model.ActivityTypeRoomCreated,
		CampID:      room.RoomGroupID, // RoomGroup経由でCampIDを取得する必要がある
		ReferenceID: room.ID,
	}
	return s.repo.CreateActivity(ctx, activity)
}

// RecordRoomCreatedWithCampID はCampIDが既知の場合に使う
func (s *activityServiceImpl) RecordRoomCreatedWithCampID(
	ctx context.Context,
	room model.Room,
	campID uint,
) error {
	activity := &model.Activity{
		Type:        model.ActivityTypeRoomCreated,
		CampID:      campID,
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

	var result []ActivityResponse

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

			payment, err := s.repo.GetPaymentByID(ctx, a.ReferenceID)
			if err != nil {
				if errors.Is(err, repository.ErrPaymentNotFound) {
					continue
				}
				return nil, err
			}

			resp := ActivityResponse{
				ID:   a.ID,
				Type: a.Type,
				Time: a.CreatedAt,
			}

			switch a.Type {
			case model.ActivityTypePaymentCreated:
				resp.PaymentCreated = &PaymentCreatedDetail{Amount: payment.Amount}
			case model.ActivityTypePaymentAmountChanged:
				detail := &PaymentChangedDetail{Amount: payment.Amount}
				resp.PaymentAmountChanged = detail
			default:
				detail := &PaymentChangedDetail{Amount: payment.Amount}
				resp.PaymentPaidChanged = detail
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
			needsResponse := false
			for _, q := range qg.Questions {
				if q.IsRequired && !answeredQuestionIDs[q.ID] {
					needsResponse = true
					break
				}
			}

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

	// CreatedAt降順でソート（repositoryで既にソートされているが念のため）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time.After(result[j].Time)
	})

	return result, nil
}
