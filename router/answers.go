package router

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

// DMの文章の構築に使う
// テンプレート部分が大体120バイトぐらいなので256バイト確保しておく
const defaultBuilderSize = 256

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
			slog.WarnContext(
				e.Request().Context(),
				"question group not found",
				slog.Int("questionGroupId", questionGroupId),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Question group not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get answers",
			slog.String("error", err.Error()),
			slog.Int("questionGroupId", questionGroupId),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

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
			slog.WarnContext(
				e.Request().Context(),
				"question not found",
				slog.Int("questionId", questionID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		if errors.Is(err, model.ErrForbidden) {
			slog.WarnContext(
				e.Request().Context(),
				"question is not public",
				slog.Int("questionId", questionID),
			)

			return echo.NewHTTPError(http.StatusForbidden, "Question is not public")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get public answers",
			slog.String("error", err.Error()),
			slog.Int("questionId", questionID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

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
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	answers, err := converter.Convert[[]model.Answer](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	for i := range answers {
		answers[i].UserID = *params.XForwardedUser
	}

	if err := s.repo.CreateAnswers(e.Request().Context(), &answers); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to create answers",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

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
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	answer, err := converter.Convert[model.Answer](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	answer.UserID = *params.XForwardedUser

	if err := s.repo.UpdateAnswer(e.Request().Context(), uint(answerID), &answer); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update answer",
			slog.String("error", err.Error()),
			slog.Int("answerId", answerID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.AnswerResponse](answer)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

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
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

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
			slog.WarnContext(
				e.Request().Context(),
				"question not found",
				slog.Int("questionId", questionID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get answers",
			slog.String("error", err.Error()),
			slog.Int("questionId", questionID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

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
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

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
			slog.WarnContext(
				e.Request().Context(),
				"question group not found",
				slog.Int("questionGroupId", questionGroupID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Question group not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get answers",
			slog.String("error", err.Error()),
			slog.Int("questionGroupId", questionGroupID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

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
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	targetUser, err := s.repo.GetOrCreateUser(
		e.Request().Context(),
		userID,
	)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create target user",
			slog.String("error", err.Error()),
			slog.String("userId", userID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var req api.AdminPostAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	answer, err := converter.Convert[model.Answer](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 対象ユーザーのIDを設定
	answer.UserID = targetUser.ID

	// 回答を作成
	if err := s.repo.CreateAnswer(e.Request().Context(), &answer); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"question or option not found",
				slog.String("error", err.Error()),
				slog.Int("questionId", int(answer.QuestionID)),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Question or option not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to create answer",
			slog.String("error", err.Error()),
			slog.Int("questionId", int(answer.QuestionID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	go func(newAnswer model.Answer) {
		ctx := context.WithoutCancel(e.Request().Context())
		question, err := s.repo.GetQuestionByID(newAnswer.QuestionID)

		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to get question by ID",
				slog.String("error", err.Error()),
				slog.Int("questionId", int(newAnswer.QuestionID)),
				slog.Int("answerId", int(newAnswer.ID)),
			)

			return
		}

		var messageBuilder strings.Builder

		messageBuilder.Grow(defaultBuilderSize)
		messageBuilder.WriteString("@")
		messageBuilder.WriteString(*params.XForwardedUser)
		messageBuilder.WriteString("がアンケート「")
		messageBuilder.WriteString(question.Title)
		messageBuilder.WriteString("」のあなたの回答を変更しました。\n")

		switch newAnswer.Type {
		case model.FreeTextQuestion:
			if newAnswer.FreeTextContent != nil {
				messageBuilder.WriteString("### 変更前\n")
				messageBuilder.WriteString("未回答\n")
				messageBuilder.WriteString("### 変更後\n")
				messageBuilder.WriteString("```\n")
				messageBuilder.WriteString(*newAnswer.FreeTextContent)
				messageBuilder.WriteString("\n```\n")
			} else {
				slog.ErrorContext(
					ctx,
					"FreeTextContent of answer is nil",
					slog.Int("answerId", int(newAnswer.ID)),
				)

				return
			}

		case model.FreeNumberQuestion:
			if newAnswer.FreeNumberContent != nil {
				messageBuilder.WriteString("### 変更前\n")
				messageBuilder.WriteString("未回答\n")
				messageBuilder.WriteString("### 変更後\n")
				messageBuilder.WriteString("```\n")
				messageBuilder.WriteString(
					strconv.FormatFloat(*newAnswer.FreeNumberContent, 'g', -1, 64),
				)
				messageBuilder.WriteString("\n```\n")
			} else {
				slog.ErrorContext(
					ctx,
					"FreeNumberContent of answer is nil",
					slog.Int("answerId", int(newAnswer.ID)),
				)

				return
			}

		case model.SingleChoiceQuestion:
			var newOptionContent string

			if len(newAnswer.SelectedOptions) == 1 {
				newOptionContent = newAnswer.SelectedOptions[0].Content
			} else {
				slog.ErrorContext(
					ctx,
					"The number of selected options for new answer is not 1",
					slog.Int("answerId", int(newAnswer.ID)),
				)

				return
			}

			messageBuilder.WriteString("### 変更前\n")
			messageBuilder.WriteString("未回答\n")
			messageBuilder.WriteString("### 変更後\n")
			messageBuilder.WriteString(newOptionContent)

		case model.MultipleChoiceQuestion:
			messageBuilder.WriteString("### 変更前\n")
			messageBuilder.WriteString("未回答\n")
			messageBuilder.WriteString("### 変更後\n")

			for _, opt := range newAnswer.SelectedOptions {
				messageBuilder.WriteString("- ")
				messageBuilder.WriteString(opt.Content)
				messageBuilder.WriteString("\n")
			}
		}

		if err := s.traqService.PostDirectMessage(ctx, newAnswer.UserID, messageBuilder.String()); err != nil {
			slog.ErrorContext(
				ctx,
				"failed to send direct message",
				slog.String("error", err.Error()),
				slog.String("userId", newAnswer.UserID),
				slog.Int("answerId", int(newAnswer.ID)),
			)
		}
	}(answer)

	res, err := converter.Convert[api.AnswerResponse](answer)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

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
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	answer, err := converter.Convert[model.Answer](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 変更前の回答を取得
	oldAnswer, err := s.repo.GetAnswerByID(e.Request().Context(), uint(answerID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"answer not found",
				slog.Int("answerId", answerID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Answer not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get answer",
			slog.String("error", err.Error()),
			slog.Int("answerId", answerID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateAnswer(e.Request().Context(), uint(answerID), &answer); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update answer",
			slog.String("error", err.Error()),
			slog.Int("answerId", answerID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 回答者にDMを送信（非同期）
	go func(oldAnswer, newAnswer model.Answer) {
		ctx := context.WithoutCancel(e.Request().Context())
		question, err := s.repo.GetQuestionByID(newAnswer.QuestionID)

		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to get question by ID",
				slog.String("error", err.Error()),
				slog.Int("questionId", int(newAnswer.QuestionID)),
				slog.Int("answerId", answerID),
			)

			return
		}

		var messageBuilder strings.Builder

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
				slog.ErrorContext(
					ctx,
					"FreeTextContent of answer is nil",
					slog.Int("answerId", answerID),
				)

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
				slog.ErrorContext(
					ctx,
					"FreeNumberContent of answer is nil",
					slog.Int("answerId", answerID),
				)

				return
			}

		case model.SingleChoiceQuestion:
			var oldOptionContent, newOptionContent string

			if len(oldAnswer.SelectedOptions) == 1 {
				oldOptionContent = oldAnswer.SelectedOptions[0].Content
			} else {
				slog.ErrorContext(
					ctx,
					"The number of selected options for old answer is not 1",
					slog.Int("answerId", answerID),
				)

				return
			}

			if len(newAnswer.SelectedOptions) == 1 {
				newOptionContent = newAnswer.SelectedOptions[0].Content
			} else {
				slog.ErrorContext(
					ctx,
					"The number of selected options for new answer is not 1",
					slog.Int("answerId", answerID),
				)

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
			slog.ErrorContext(
				ctx,
				"failed to send direct message",
				slog.String("error", err.Error()),
				slog.String("userId", newAnswer.UserID),
				slog.Int("answerId", answerID),
			)
		}
	}(*oldAnswer, answer)

	res, err := converter.Convert[api.AnswerResponse](answer)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}
