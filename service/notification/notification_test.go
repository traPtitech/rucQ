package notification

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository/mockrepository"
	"github.com/traPtitech/rucQ/service/traq/mocktraq"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestNotificationServiceImpl_SendAnswerChangeMessage(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	editorUserID := random.AlphaNumericString(t, 32)
	userID := random.AlphaNumericString(t, 32)
	questionID := uint(random.PositiveInt(t))
	questionTitle := random.AlphaNumericString(t, 20)

	question := &model.Question{
		Model: gorm.Model{ID: questionID},
		Title: questionTitle,
	}

	t.Run("FreeText - 新規回答", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newContent := random.AlphaNumericString(t, 50)
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: &newContent,
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			"未回答\n" +
			"### 変更後\n" +
			newContent + "\n"

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("FreeText - 回答更新", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		oldContent := random.AlphaNumericString(t, 40)
		newContent := random.AlphaNumericString(t, 50)
		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: &oldContent,
		}
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: oldAnswer.ID},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: &newContent,
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			oldContent + "\n" +
			"### 変更後\n" +
			newContent + "\n"

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, oldAnswer, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("FreeNumber - 新規回答", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newNumber := random.Float64(t)
		newAnswer := model.Answer{
			Model:             gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:            userID,
			QuestionID:        questionID,
			Type:              model.FreeNumberQuestion,
			FreeNumberContent: &newNumber,
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			"未回答\n" +
			"### 変更後\n" +
			strconv.FormatFloat(
				newNumber,
				'g',
				-1,
				64,
			) + "\n"

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("FreeNumber - 回答更新", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		oldNumber := random.Float64(t)
		newNumber := random.Float64(t)
		oldAnswer := &model.Answer{
			Model:             gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:            userID,
			QuestionID:        questionID,
			Type:              model.FreeNumberQuestion,
			FreeNumberContent: &oldNumber,
		}
		newAnswer := model.Answer{
			Model:             gorm.Model{ID: oldAnswer.ID},
			UserID:            userID,
			QuestionID:        questionID,
			Type:              model.FreeNumberQuestion,
			FreeNumberContent: &newNumber,
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			strconv.FormatFloat(
				oldNumber,
				'g',
				-1,
				64,
			) + "\n" +
			"### 変更後\n" +
			strconv.FormatFloat(
				newNumber,
				'g',
				-1,
				64,
			) + "\n"

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, oldAnswer, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("SingleChoice - 新規回答", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newOption := model.Option{Content: random.AlphaNumericString(t, 20)}
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.SingleChoiceQuestion,
			SelectedOptions: []model.Option{newOption},
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			"未回答\n" +
			"### 変更後\n" +
			newOption.Content

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("SingleChoice - 回答更新", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		oldOption := model.Option{Content: random.AlphaNumericString(t, 15)}
		newOption := model.Option{Content: random.AlphaNumericString(t, 20)}
		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.SingleChoiceQuestion,
			SelectedOptions: []model.Option{oldOption},
		}
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: oldAnswer.ID},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.SingleChoiceQuestion,
			SelectedOptions: []model.Option{newOption},
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			oldOption.Content + "\n" +
			"### 変更後\n" +
			newOption.Content

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, oldAnswer, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("MultipleChoice - 新規回答（複数選択）", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		optionA := random.AlphaNumericString(t, 15)
		optionB := random.AlphaNumericString(t, 20)
		options := []model.Option{
			{Content: optionA},
			{Content: optionB},
		}
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.MultipleChoiceQuestion,
			SelectedOptions: options,
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			"未回答\n" +
			"### 変更後\n" +
			"- " + optionA + "\n" +
			"- " + optionB + "\n"

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("MultipleChoice - 選択なし", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		oldOptionContent := random.AlphaNumericString(t, 15)
		oldOptions := []model.Option{{Content: oldOptionContent}}
		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.MultipleChoiceQuestion,
			SelectedOptions: oldOptions,
		}
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: oldAnswer.ID},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.MultipleChoiceQuestion,
			SelectedOptions: []model.Option{},
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			"- " + oldOptionContent + "\n" +
			"### 変更後\n" +
			"選択なし\n"

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, oldAnswer, newAnswer)

		assert.NoError(t, err)
	})

	t.Run("質問取得エラー", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newContent := random.AlphaNumericString(t, 30)
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: &newContent,
		}

		expectedError := errors.New("question not found")
		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(nil, expectedError)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.Equal(t, expectedError, err)
	})

	t.Run("FreeText - 新回答のFreeTextContentがnil", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newAnswer := model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: nil,
		}

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.Error(t, err)
		assert.Equal(t, "FreeTextContent of new answer is nil", err.Error())
	})

	t.Run("FreeText - 旧回答のFreeTextContentがnil", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newContent := random.AlphaNumericString(t, 30)
		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: nil,
		}
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: oldAnswer.ID},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: &newContent,
		}

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, oldAnswer, newAnswer)

		assert.Error(t, err)
		assert.Equal(t, "FreeTextContent of old answer is nil", err.Error())
	})

	t.Run("FreeNumber - 新回答のFreeNumberContentがnil", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newAnswer := model.Answer{
			Model:             gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:            userID,
			QuestionID:        questionID,
			Type:              model.FreeNumberQuestion,
			FreeNumberContent: nil,
		}

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.Error(t, err)
		assert.Equal(t, "FreeNumberContent of new answer is nil", err.Error())
	})

	t.Run("FreeNumber - 旧回答のFreeNumberContentがnil", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newNumber := random.Float64(t)
		oldAnswer := &model.Answer{
			Model:             gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:            userID,
			QuestionID:        questionID,
			Type:              model.FreeNumberQuestion,
			FreeNumberContent: nil,
		}
		newAnswer := model.Answer{
			Model:             gorm.Model{ID: oldAnswer.ID},
			UserID:            userID,
			QuestionID:        questionID,
			Type:              model.FreeNumberQuestion,
			FreeNumberContent: &newNumber,
		}

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, oldAnswer, newAnswer)

		assert.Error(t, err)
		assert.Equal(t, "FreeNumberContent of old answer is nil", err.Error())
	})

	t.Run("SingleChoice - 新回答の選択肢数が1でない", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newAnswer := model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.SingleChoiceQuestion,
			SelectedOptions: []model.Option{},
		}

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.Error(t, err)
		assert.Equal(t, "the number of selected options for new answer is not 1", err.Error())
	})

	t.Run("SingleChoice - 旧回答の選択肢数が1でない", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		oldAnswer := &model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.SingleChoiceQuestion,
			SelectedOptions: []model.Option{},
		}
		newAnswerOption := random.AlphaNumericString(t, 15)
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: oldAnswer.ID},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.SingleChoiceQuestion,
			SelectedOptions: []model.Option{{Content: newAnswerOption}},
		}

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, oldAnswer, newAnswer)

		assert.Error(t, err)
		assert.Equal(t, "the number of selected options for old answer is not 1", err.Error())
	})

	t.Run("TraqService エラー", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := mockrepository.NewMockRepository(ctrl)
		traqService := mocktraq.NewMockTraqService(ctrl)
		s := NewNotificationService(repo, traqService)

		newContent := random.AlphaNumericString(t, 30)
		newAnswer := model.Answer{
			Model:           gorm.Model{ID: uint(random.PositiveInt(t))},
			UserID:          userID,
			QuestionID:      questionID,
			Type:            model.FreeTextQuestion,
			FreeTextContent: &newContent,
		}

		expectedMessage := "@" + editorUserID + "がアンケート「" + questionTitle + "」のあなたの回答を変更しました\n" +
			"### 変更前\n" +
			"未回答\n" +
			"### 変更後\n" +
			newContent + "\n"

		repo.MockQuestionRepository.EXPECT().
			GetQuestionByID(questionID).
			Return(question, nil)

		expectedError := errors.New("failed to post message")
		traqService.EXPECT().
			PostDirectMessage(ctx, userID, expectedMessage).
			Return(expectedError)

		err := s.SendAnswerChangeMessage(ctx, editorUserID, nil, newAnswer)

		assert.Equal(t, expectedError, err)
	})
}
