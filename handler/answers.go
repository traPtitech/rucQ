package handler

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/rucQ/backend/repository"
)

func (s *Server) GetMyAnswer(e echo.Context, questionID QuestionId, params GetMyAnswerParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	answer, err := s.repo.GetOrCreateAnswer(&repository.GetAnswerQuery{
		QuestionID: uint(questionID),
		UserID:     user.ID,
	})

	if err != nil {
		e.Logger().Errorf("failed to get or create answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res Answer

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
	question, err := s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if answer.Content != nil {
		if question.Type == string(QuestionTypeMultiple) {
			if err := res.Content.FromAnswerContent1(*answer.Content); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		} else {
			if err := res.Content.FromAnswerContent0((*answer.Content)[0]); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		}
	}

	res.UserTraqId = *params.XForwardedUser

	return e.JSON(http.StatusOK, res)
}

func (s *Server) PutAnswer(e echo.Context, questionID QuestionId, params PutAnswerParams) error {
	var req PutAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	answer, err := s.repo.GetOrCreateAnswer(&repository.GetAnswerQuery{
		QuestionID: uint(questionID),
		UserID:     user.ID,
	})

	if err != nil {
		e.Logger().Errorf("failed to get or create answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	question, err := s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if req.Content != nil {
		if question.Type == string(QuestionTypeMultiple) {
			content, err := req.Content.AsPutAnswerRequestContent1()

			if err != nil {
				return e.JSON(http.StatusBadRequest, err)
			}

			contentStrSlice := []string(content)
			answer.Content = &contentStrSlice
		} else {
			content, err := req.Content.AsPutAnswerRequestContent0()

			if err != nil {
				return e.JSON(http.StatusBadRequest, err)
			}

			contentStrSlice := []string{string(content)}
			answer.Content = &contentStrSlice
		}
	}

	if err := s.repo.UpdateAnswer(answer); err != nil {
		e.Logger().Errorf("failed to update answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res Answer

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if answer.Content != nil {
		if question.Type == string(QuestionTypeMultiple) {
			if err := res.Content.FromAnswerContent1(*answer.Content); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		} else {
			if err := res.Content.FromAnswerContent0((*answer.Content)[0]); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		}
	}

	res.UserTraqId = *params.XForwardedUser

	return e.JSON(http.StatusOK, res)
}

func (s *Server) GetUserAnswer(e echo.Context, traqID TraqId, questionID QuestionId, paramas GetUserAnswerParams) error {
	operator, err := s.repo.GetOrCreateUser(*paramas.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	user, err := s.repo.GetOrCreateUser(traqID)

	if err != nil {
		e.Logger().Errorf("failed to get user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	answer, err := s.repo.GetOrCreateAnswer(&repository.GetAnswerQuery{
		QuestionID: uint(questionID),
		UserID:     user.ID,
	})

	if err != nil {
		e.Logger().Errorf("failed to get answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res Answer

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	question, err := s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if answer.Content != nil {
		if question.Type == string(QuestionTypeMultiple) {
			if err := res.Content.FromAnswerContent1(*answer.Content); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		} else {
			if err := res.Content.FromAnswerContent0((*answer.Content)[0]); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		}
	}

	res.UserTraqId = string(traqID)

	return e.JSON(http.StatusOK, res)
}

func (s *Server) PutUserAnswer(e echo.Context, traqID TraqId, questionID QuestionId, params PutUserAnswerParams) error {
	operator, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req PutUserAnswerJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	user, err := s.repo.GetOrCreateUser(traqID)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	answer, err := s.repo.GetOrCreateAnswer(&repository.GetAnswerQuery{
		QuestionID: uint(questionID),
		UserID:     user.ID,
	})

	if err != nil {
		e.Logger().Errorf("failed to get or create answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	question, err := s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if req.Content != nil {
		if question.Type == string(QuestionTypeMultiple) {
			content, err := req.Content.AsPutAnswerRequestContent1()

			if err != nil {
				return e.JSON(http.StatusBadRequest, err)
			}

			contentStrSlice := []string(content)
			answer.Content = &contentStrSlice
		} else {
			content, err := req.Content.AsPutAnswerRequestContent0()

			if err != nil {
				return e.JSON(http.StatusBadRequest, err)
			}

			contentStrSlice := []string{string(content)}
			answer.Content = &contentStrSlice
		}
	}

	if err := s.repo.UpdateAnswer(answer); err != nil {
		e.Logger().Errorf("failed to update answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res Answer

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if answer.Content != nil {
		if question.Type == string(QuestionTypeMultiple) {
			if err := res.Content.FromAnswerContent1(*answer.Content); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		} else {
			if err := res.Content.FromAnswerContent0((*answer.Content)[0]); err != nil {
				e.Logger().Errorf("failed to convert content: %v", err)

				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		}

		res.UserTraqId = string(traqID)

		return e.JSON(http.StatusOK, res)
	}

	return e.JSON(http.StatusOK, res)
}
