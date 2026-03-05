package notification

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/service/traq"
)

type notificationServiceImpl struct {
	repo        repository.Repository
	traqService traq.TraqService
}

func NewNotificationService(
	repo repository.Repository,
	traqService traq.TraqService,
) *notificationServiceImpl {
	return &notificationServiceImpl{
		repo:        repo,
		traqService: traqService,
	}
}

func (s *notificationServiceImpl) SendAnswerChangeMessage(
	ctx context.Context,
	editorUserID string,
	oldAnswer *model.Answer,
	newAnswer model.Answer,
) error {
	question, err := s.repo.GetQuestionByID(newAnswer.QuestionID)

	if err != nil {
		return err
	}

	// テンプレート部分が大体120バイトぐらいなので256バイト確保しておく
	const defaultBuilderSize = 256

	var messageBuilder strings.Builder

	messageBuilder.Grow(defaultBuilderSize)
	messageBuilder.WriteString("@")
	messageBuilder.WriteString(editorUserID)
	messageBuilder.WriteString("がアンケート「")
	messageBuilder.WriteString(question.Title)
	messageBuilder.WriteString("」のあなたの回答を変更しました\n")

	switch newAnswer.Type {
	case model.FreeTextQuestion:
		if newAnswer.FreeTextContent == nil {
			return errors.New("FreeTextContent of new answer is nil")
		}

		messageBuilder.WriteString("### 変更前\n")

		if oldAnswer == nil {
			messageBuilder.WriteString("未回答\n")
		} else {
			if oldAnswer.FreeTextContent == nil {
				return errors.New("FreeTextContent of old answer is nil")
			}

			messageBuilder.WriteString(*oldAnswer.FreeTextContent)
			messageBuilder.WriteString("\n")
		}

		messageBuilder.WriteString("### 変更後\n")
		messageBuilder.WriteString(*newAnswer.FreeTextContent)
		messageBuilder.WriteString("\n")

	case model.FreeNumberQuestion:
		if newAnswer.FreeNumberContent == nil {
			return errors.New("FreeNumberContent of new answer is nil")
		}

		messageBuilder.WriteString("### 変更前\n")

		if oldAnswer == nil {
			messageBuilder.WriteString("未回答\n")
		} else {
			if oldAnswer.FreeNumberContent == nil {
				return errors.New("FreeNumberContent of old answer is nil")
			}

			messageBuilder.WriteString(
				strconv.FormatFloat(*oldAnswer.FreeNumberContent, 'g', -1, 64),
			)
			messageBuilder.WriteString("\n")
		}

		messageBuilder.WriteString("### 変更後\n")
		messageBuilder.WriteString(
			strconv.FormatFloat(*newAnswer.FreeNumberContent, 'g', -1, 64),
		)
		messageBuilder.WriteString("\n")

	case model.SingleChoiceQuestion:
		if len(newAnswer.SelectedOptions) != 1 {
			return errors.New("the number of selected options for new answer is not 1")
		}

		newOptionContent := newAnswer.SelectedOptions[0].Content

		messageBuilder.WriteString("### 変更前\n")

		if oldAnswer == nil {
			messageBuilder.WriteString("未回答\n")
		} else {
			if len(oldAnswer.SelectedOptions) != 1 {
				return errors.New("the number of selected options for old answer is not 1")
			}

			messageBuilder.WriteString(oldAnswer.SelectedOptions[0].Content)
			messageBuilder.WriteString("\n")
		}

		messageBuilder.WriteString("### 変更後\n")
		messageBuilder.WriteString(newOptionContent)

	case model.MultipleChoiceQuestion:
		messageBuilder.WriteString("### 変更前\n")

		if oldAnswer == nil {
			messageBuilder.WriteString("未回答\n")
		} else {
			if len(oldAnswer.SelectedOptions) == 0 {
				messageBuilder.WriteString("選択なし\n")
			} else {
				for _, opt := range oldAnswer.SelectedOptions {
					messageBuilder.WriteString("- ")
					messageBuilder.WriteString(opt.Content)
					messageBuilder.WriteString("\n")
				}
			}
		}

		messageBuilder.WriteString("### 変更後\n")

		if len(newAnswer.SelectedOptions) == 0 {
			messageBuilder.WriteString("選択なし\n")
		} else {
			for _, opt := range newAnswer.SelectedOptions {
				messageBuilder.WriteString("- ")
				messageBuilder.WriteString(opt.Content)
				messageBuilder.WriteString("\n")
			}
		}
	}

	if err := s.traqService.PostDirectMessage(
		ctx,
		newAnswer.UserID,
		messageBuilder.String(),
	); err != nil {
		return err
	}

	return nil
}
