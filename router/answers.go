package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetMyAnswers(
	e echo.Context,
	questionGroupId api.QuestionGroupId,
	params api.GetMyAnswersParams,
) error {
	answers, err := s.repo.GetAnswersByUserAndQuestionGroup(
		e.Request().Context(),
		*params.XForwardedUser,
		uint(questionGroupId),
	)

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

func (s *Server) PostAnswers(
	e echo.Context,
	_ api.QuestionGroupId,
	params api.PostAnswersParams,
) error {
	var req api.PostAnswersJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
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
func (s *Server) PutAnswer(e echo.Context, _ api.AnswerId, params api.PutAnswerParams) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}
	var req api.PutAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	// user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	// if err != nil {
	// 	e.Logger().Errorf("failed to get or create user: %v", err)

	// 	return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	// }

	// answer, err := s.repo.GetOrCreateAnswer(&repository.GetAnswerQuery{
	// 	QuestionID: uint(answerID),
	// 	UserID:     user.ID,
	// })

	// if err != nil {
	// 	e.Logger().Errorf("failed to get or create answer: %v", err)

	// 	return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	// }

	// _, err = s.repo.GetQuestionByID(uint(answerID))

	// if err != nil {
	// 	e.Logger().Errorf("failed to get question: %v", err)

	// 	return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	// }

	// TODO: リクエストボディの型が変わったため、以下の処理を修正する
	// if req.Content != nil {
	//  if question.Type == string(QuestionTypeMultiple) {
	//   content, err := req.Content.AsPutAnswerRequestContent1()
	//
	//   if err != nil {
	//    return e.JSON(http.StatusBadRequest, err)
	//   }
	//
	//   contentStrSlice := []string(content)
	//   answer.Content = &contentStrSlice
	//  } else {
	//   content, err := req.Content.AsPutAnswerRequestContent0()
	//
	//   if err != nil {
	//    return e.JSON(http.StatusBadRequest, err)
	//   }
	//
	//   contentStrSlice := []string{string(content)}
	//   answer.Content = &contentStrSlice
	//  }
	// }

	// if err := s.repo.UpdateAnswer(answer); err != nil {
	// 	e.Logger().Errorf("failed to update answer: %v", err)

	// 	return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	// }

	var res any

	// if err := copier.Copy(&res, &answer); err != nil {
	// 	e.Logger().Errorf("failed to copy model to response: %v", err)

	// 	return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	// }

	// TODO: リクエストボディの型が変わったため、以下の処理を修正する
	// if answer.Content != nil {
	//  if question.Type == string(QuestionTypeMultiple) {
	//   if err := res.Content.FromAnswerContent1(*answer.Content); err != nil {
	//    e.Logger().Errorf("failed to convert content: %v", err)
	//
	//    return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	//   }
	//  } else {
	//   if err := res.Content.FromAnswerContent0((*answer.Content)[0]); err != nil {
	//    e.Logger().Errorf("failed to convert content: %v", err)
	//
	//    return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	//   }
	// }

	// res.UserTraqId = *params.XForwardedUser // TODO: AnswerBody に UserTraqId がないためコメントアウト

	return e.JSON(http.StatusOK, res)
}
