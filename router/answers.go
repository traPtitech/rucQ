package router

import (
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/rucQ/backend/repository"
)

// GetMyAnswer アンケート回答取得
// (GET /users/me/answers)
func (s *Server) GetMyAnswer(e echo.Context, questionID QuestionId, params GetMyAnswersParams) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
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

	var res AnswerBody

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}
	_, err = s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if answer.Content != nil {
		// TODO: model.QuestionTypeMultiple と model.QuestionTypeSingle の実装をapi.gen.goの型定義に合わせて修正する
		// if question.Type == string(QuestionTypeMultiple) {
		//  if err := res.Content.FromAnswerContent1(*answer.Content); err != nil {
		//   e.Logger().Errorf("failed to convert content: %v", err)
		//
		//   return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		//  }
		// } else {
		//  if err := res.Content.FromAnswerContent0((*answer.Content)[0]); err != nil {
		//   e.Logger().Errorf("failed to convert content: %v", err)
		//
		//   return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		//  }
		// }
	}

	// res.UserTraqId = *params.XForwardedUser // TODO: AnswerBody に UserTraqId がないためコメントアウト

	return e.JSON(http.StatusOK, res)
}

// PutAnswer アンケート回答編集
// (PUT /users/me/answers)
func (s *Server) PutAnswer(e echo.Context, questionID QuestionId, params PutAnswerParams) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}
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

	_, err = s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

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

	if err := s.repo.UpdateAnswer(answer); err != nil {
		e.Logger().Errorf("failed to update answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res AnswerBody // TODO: Answer を AnswerBody に修正

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

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

// GetUserAnswer ユーザーのアンケート回答取得
// (GET /users/{traQID}/answers)
func (s *Server) GetUserAnswer(e echo.Context, traqID string, questionID QuestionId, params AdminGetAnswersParams) error { // TODO: GetUserAnswerParams を AdminGetAnswersParams に修正
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}
	operator, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

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

	var res AnswerBody // TODO: Answer を AnswerBody に修正

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	_, err = s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

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
	//  }
	// }

	// res.UserTraqId = string(traqID) // TODO: AnswerBody に UserTraqId がないためコメントアウト

	return e.JSON(http.StatusOK, res)
}

// PutUserAnswer ユーザーのアンケート回答編集
// (PUT /users/{traQID}/answers)
func (s *Server) PutUserAnswer(e echo.Context, traqID string, questionID QuestionId, params AdminPutAnswerParams) error { // TODO: PutUserAnswerParams を AdminPutAnswerParams に修正
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}
	operator, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req AdminPutAnswerJSONRequestBody // TODO: PutUserAnswerJSONRequestBody を AdminPutAnswerJSONRequestBody に修正

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

	_, err = s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

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

	if err := s.repo.UpdateAnswer(answer); err != nil {
		e.Logger().Errorf("failed to update answer: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res AnswerBody // TODO: Answer を AnswerBody に修正

	if err := copier.Copy(&res, &answer); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

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
	//  }
	//
	//  // res.UserTraqId = string(traqID) // TODO: AnswerBody に UserTraqId がないためコメントアウト
	//
	//  return e.JSON(http.StatusOK, res)
	// }

	return e.JSON(http.StatusOK, res)
}

// PostAnswer アンケート回答作成
// (POST /answers)
func (s *Server) PostAnswer(c echo.Context, params PostAnswerParams) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "Not implemented")
}

// AnswerBody defines model for Answer.
// TODO: この型定義は oapi-codegen が生成する型と合わせる必要がある
type AnswerBody struct {
	Content    *string `json:"content,omitempty"` // 仮の型。実際の型に合わせる
	Id         *int    `json:"id,omitempty"`
	QuestionId *int    `json:"questionId,omitempty"`
	UserId     *string `json:"userId,omitempty"`
}

// TODO: 以下の関数は、oapi-codegen が生成する型に合わせて修正する必要がある
// func (a *AnswerBody) FromAnswerContent1(content []string) error {
//  return errors.New("not implemented")
// }
//
// func (a *AnswerBody) FromAnswerContent0(content string) error {
//  return errors.New("not implemented")
// }
