package router

import (
	"errors"
	"net/http"

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
	query := repository.GetAnswersQuery{
		UserID:                params.XForwardedUser,
		QuestionGroupID:       &[]uint{uint(questionGroupId)}[0],
		IncludePrivateAnswers: true, // 自分の回答は非公開でも取得
	}

	answers, err := s.repo.GetAnswers(e.Request().Context(), query)

	if err != nil {
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
	query := repository.GetAnswersQuery{
		QuestionID:            &[]uint{uint(questionID)}[0],
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

	query := repository.GetAnswersQuery{
		QuestionID:            &[]uint{uint(questionID)}[0],
		IncludePrivateAnswers: true, // 管理者は非公開回答も取得可能
	}

	answers, err := s.repo.GetAnswers(e.Request().Context(), query)

	if err != nil {
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
	query := repository.GetAnswersQuery{
		QuestionGroupID:       &[]uint{uint(questionGroupID)}[0],
		IncludePrivateAnswers: true, // 管理者は非公開回答も取得可能
	}

	if params.UserId != nil {
		query.UserID = params.UserId
	}

	answers, err := s.repo.GetAnswers(e.Request().Context(), query)

	if err != nil {
		if params.UserId != nil {
			e.Logger().Errorf("failed to get answers for user %s: %v", *params.UserId, err)
		} else {
			e.Logger().Errorf("failed to get answers for question group %d: %v", questionGroupID, err)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.AnswerResponse](answers)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
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
