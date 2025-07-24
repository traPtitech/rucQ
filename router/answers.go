package router

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (s *Server) GetMyAnswers(
	e echo.Context,
	questionGroupId api.QuestionGroupId,
	params api.GetMyAnswersParams,
) error {
	uintQuestionGroupID := uint(questionGroupId)
	query := repository.GetAnswersQuery{
		UserID:                params.XForwardedUser,
		QuestionGroupID:       &uintQuestionGroupID,
		IncludePrivateAnswers: true, // 自分の回答は非公開でも取得
	}

	answers, err := s.repo.GetAnswers(e.Request().Context(), query)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("question group %d not found", questionGroupId)

			return echo.NewHTTPError(http.StatusNotFound, "Question group not found")
		}

		e.Logger().Errorf("failed to get answers: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) GetAnswers(e echo.Context, questionID api.QuestionId) error {
	uintQuestionID := uint(questionID)
	query := repository.GetAnswersQuery{
		QuestionID:            &uintQuestionID,
		IncludePrivateAnswers: false, // 公開回答のみ
	}

	answers, err := s.repo.GetAnswers(e.Request().Context(), query)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("question %d not found", questionID)

			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		if errors.Is(err, model.ErrForbidden) {
			e.Logger().Warnf("question %d is not public", questionID)

			return echo.NewHTTPError(http.StatusForbidden, "Question is not public")
		}

		e.Logger().Errorf("failed to get public answers: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) PostAnswers(
	e echo.Context,
	_ api.QuestionGroupId,
	params api.PostAnswersParams,
) error {
	var req api.PostAnswersJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	answers, err := converter.Convert[[]model.Answer](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	for i := range answers {
		answers[i].UserID = *params.XForwardedUser
	}

	if err := s.repo.CreateAnswers(e.Request().Context(), &answers); err != nil {
		e.Logger().Errorf("failed to create answers: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

// PutAnswer アンケート回答編集
func (s *Server) PutAnswer(
	e echo.Context,
	answerID api.AnswerId,
	params api.PutAnswerParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	var req api.PutAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	answer, err := converter.Convert[model.Answer](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	answer.UserID = *params.XForwardedUser

	if err := s.repo.UpdateAnswer(e.Request().Context(), uint(answerID), &answer); err != nil {
		e.Logger().Errorf("failed to update answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.AnswerResponse](answer)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

// AdminGetAnswers 回答の一覧を取得（管理者用）
func (s *Server) AdminGetAnswers(
	e echo.Context,
	questionID api.QuestionId,
	params api.AdminGetAnswersParams,
) error {
	user, err := s.repo.GetOrCreateUser(
		e.Request().Context(),
		*params.XForwardedUser,
	)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	uintQuestionID := uint(questionID)
	query := repository.GetAnswersQuery{
		QuestionID:            &uintQuestionID,
		IncludePrivateAnswers: true, // 管理者は非公開回答も取得可能
	}

	answers, err := s.repo.GetAnswers(e.Request().Context(), query)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("question %d not found", questionID)

			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		e.Logger().Errorf("failed to get answers: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminGetAnswersForQuestionGroup(
	e echo.Context,
	questionGroupID api.QuestionGroupId,
	params api.AdminGetAnswersForQuestionGroupParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	user, err := s.repo.GetOrCreateUser(
		e.Request().Context(),
		*params.XForwardedUser,
	)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// userIdパラメータが指定されている場合は特定ユーザーの回答を取得
	// 指定されていない場合は全ユーザーの回答を取得
	uintQuestionGroupID := uint(questionGroupID)
	query := repository.GetAnswersQuery{
		QuestionGroupID:       &uintQuestionGroupID,
		IncludePrivateAnswers: true, // 管理者は非公開回答も取得可能
		UserID:                params.UserId,
	}

	answers, err := s.repo.GetAnswers(e.Request().Context(), query)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("question group %d not found", questionGroupID)

			return echo.NewHTTPError(http.StatusNotFound, "Question group not found")
		}

		e.Logger().Errorf("failed to get answers: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostAnswer(
	e echo.Context,
	userID api.UserId,
	params api.AdminPostAnswerParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	// 管理者権限の確認
	user, err := s.repo.GetOrCreateUser(
		e.Request().Context(),
		*params.XForwardedUser,
	)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	targetUser, err := s.repo.GetOrCreateUser(
		e.Request().Context(),
		string(userID),
	)

	if err != nil {
		e.Logger().Errorf("failed to get or create target user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var req api.AdminPostAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	answer, err := converter.Convert[model.Answer](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 対象ユーザーのIDを設定
	answer.UserID = targetUser.ID

	// 回答を作成
	if err := s.repo.CreateAnswer(e.Request().Context(), &answer); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("question or option not found: %v", err)

			return echo.NewHTTPError(http.StatusNotFound, "Question or option not found")
		}

		e.Logger().Errorf("failed to create answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.AnswerResponse](answer)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutAnswer(
	e echo.Context,
	answerID api.AnswerId,
	params api.AdminPutAnswerParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	user, err := s.repo.GetOrCreateUser(
		e.Request().Context(),
		*params.XForwardedUser,
	)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	answer, err := converter.Convert[model.Answer](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 変更前の回答を取得
	oldAnswer, err := s.repo.GetAnswerByID(e.Request().Context(), uint(answerID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("answer %d not found", answerID)

			return echo.NewHTTPError(http.StatusNotFound, "Answer not found")
		}

		e.Logger().Errorf("failed to get answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateAnswer(e.Request().Context(), uint(answerID), &answer); err != nil {
		e.Logger().Errorf("failed to update answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 回答者にDMを送信（非同期）
	go func(oldAnswer, newAnswer model.Answer) {
		ctx := context.Background()
		question, err := s.repo.GetQuestionByID(newAnswer.QuestionID)

		if err != nil {
			e.Logger().Errorf(
				"failed to get question %d for DM notification: %v",
				newAnswer.QuestionID,
				err,
			)

			return
		}

		var messageBuilder strings.Builder
		// テンプレート部分が大体120バイトぐらいなので256バイト確保しておく
		const defaultBuilderSize = 256

		messageBuilder.Grow(defaultBuilderSize)
		messageBuilder.WriteString("@")
		messageBuilder.WriteString(*params.XForwardedUser)
		messageBuilder.WriteString("がアンケート「")
		messageBuilder.WriteString(question.Title)
		messageBuilder.WriteString("」のあなたの回答を変更しました。\n")

		switch newAnswer.Type {
		case model.FreeTextQuestion:
			if oldAnswer.FreeTextContent != nil && newAnswer.FreeTextContent != nil {
				messageBuilder.WriteString("### 変更前\n```\n")
				messageBuilder.WriteString(*oldAnswer.FreeTextContent)
				messageBuilder.WriteString("\n```")
				messageBuilder.WriteString("\n### 変更後\n```\n")
				messageBuilder.WriteString(*newAnswer.FreeTextContent)
				messageBuilder.WriteString("\n```")
			} else {
				e.Logger().Errorf("FreeTextContent of answer %d is nil", answerID)

				return
			}

		case model.FreeNumberQuestion:
			if oldAnswer.FreeNumberContent != nil && newAnswer.FreeNumberContent != nil {
				messageBuilder.WriteString("### 変更前\n```\n")
				messageBuilder.WriteString(
					strconv.FormatFloat(*oldAnswer.FreeNumberContent, 'g', -1, 64),
				)
				messageBuilder.WriteString("\n```")
				messageBuilder.WriteString("\n### 変更後\n```\n")
				messageBuilder.WriteString(
					strconv.FormatFloat(*newAnswer.FreeNumberContent, 'g', -1, 64),
				)
				messageBuilder.WriteString("\n```")
			} else {
				e.Logger().Errorf("FreeNumberContent of answer %d is nil", answerID)

				return
			}

		case model.SingleChoiceQuestion:
			var oldOptionContent, newOptionContent string

			if len(oldAnswer.SelectedOptions) == 1 {
				oldOptionContent = oldAnswer.SelectedOptions[0].Content
			} else {
				e.Logger().Errorf("The number of selected options for old answer %d is not 1", answerID)

				return
			}

			if len(newAnswer.SelectedOptions) == 1 {
				newOptionContent = newAnswer.SelectedOptions[0].Content
			} else {
				e.Logger().Errorf("The number of selected options for new answer %d is not 1", answerID)

				return
			}

			messageBuilder.WriteString("### 変更前\n- ")
			messageBuilder.WriteString(oldOptionContent)
			messageBuilder.WriteString("\n")
			messageBuilder.WriteString("### 変更後\n- ")
			messageBuilder.WriteString(newOptionContent)

		case model.MultipleChoiceQuestion:
			messageBuilder.WriteString("### 変更前\n")

			for _, opt := range oldAnswer.SelectedOptions {
				messageBuilder.WriteString("- ")
				messageBuilder.WriteString(opt.Content)
				messageBuilder.WriteString("\n")
			}
			messageBuilder.WriteString("### 変更後\n")

			for _, opt := range newAnswer.SelectedOptions {
				messageBuilder.WriteString("- ")
				messageBuilder.WriteString(opt.Content)
				messageBuilder.WriteString("\n")
			}
		}

		if err := s.traqService.PostDirectMessage(ctx, newAnswer.UserID, messageBuilder.String()); err != nil {
			e.Logger().Errorf("failed to send direct message to %s: %v", newAnswer.UserID, err)
		}
	}(*oldAnswer, answer)

	res, err := converter.Convert[api.AnswerResponse](answer)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}
