package converter

import (
	"errors"

	"github.com/jinzhu/copier"
	"github.com/oapi-codegen/runtime/types"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var questionGroupSchemaToModel = copier.TypeConverter{
	SrcType: api.PostQuestionGroupRequest{},
	DstType: model.QuestionGroup{},
	Fn: func(src any) (any, error) {
		req, ok := src.(api.PostQuestionGroupRequest)

		if !ok {
			return nil, errors.New("src is not an api.PostQuestionGroupRequest")
		}

		var dst model.QuestionGroup

		if err := copier.Copy(&dst, &req); err != nil {
			return nil, err
		}

		// Questions配列を手動で変換
		dst.Questions = make([]model.Question, len(req.Questions))
		for i, questionReq := range req.Questions {
			convertedQuestion, err := questionSchemaToModel.Fn(questionReq)
			if err != nil {
				return nil, err
			}
			question, ok := convertedQuestion.(model.Question)
			if !ok {
				return nil, errors.New("failed to convert question")
			}
			dst.Questions[i] = question
		}

		dst.Due = req.Due.Time // types.Dateからtime.Timeへの変換

		return dst, nil
	},
}

var questionGroupModelToSchema = copier.TypeConverter{
	SrcType: model.QuestionGroup{},
	DstType: api.QuestionGroupResponse{},
	Fn: func(src any) (any, error) {
		questionGroupModel, ok := src.(model.QuestionGroup)

		if !ok {
			return nil, errors.New("src is not a model.QuestionGroup")
		}

		var dst api.QuestionGroupResponse

		if err := copier.Copy(&dst, &questionGroupModel); err != nil {
			return nil, err
		}

		// Questions配列を手動で変換
		dst.Questions = make([]api.QuestionResponse, len(questionGroupModel.Questions))
		for i, questionModel := range questionGroupModel.Questions {
			convertedQuestion, err := questionModelToSchema.Fn(questionModel)
			if err != nil {
				return nil, err
			}
			question, ok := convertedQuestion.(api.QuestionResponse)
			if !ok {
				return nil, errors.New("failed to convert question")
			}
			dst.Questions[i] = question
		}

		dst.Due = types.Date{Time: questionGroupModel.Due} // time.Timeからtypes.Dateへの変換

		return dst, nil
	},
}
