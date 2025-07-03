package converter

import (
	"errors"

	"github.com/jinzhu/copier"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var questionSchemaToModel = copier.TypeConverter{
	SrcType: api.PostQuestionRequest{},
	DstType: model.Question{},
	Fn: func(src any) (any, error) {
		req, ok := src.(api.PostQuestionRequest)

		if !ok {
			return nil, errors.New("src is not an api.PostQuestionRequest")
		}

		var dst model.Question

		// まずTypeを判定するため、一旦FreeTextQuestionRequestに変換する
		freeTextQuestionRequest, err := req.AsFreeTextQuestionRequest()

		if err != nil {
			return nil, err
		}

		switch model.QuestionType(freeTextQuestionRequest.Type) {
		case model.FreeTextQuestion:
			if err := copier.Copy(&dst, &freeTextQuestionRequest); err != nil {
				return nil, err
			}

		case model.FreeNumberQuestion:
			freeNumberQuestionRequest, err := req.AsFreeNumberQuestionRequest()
			if err != nil {
				return nil, err
			}

			if err := copier.Copy(&dst, &freeNumberQuestionRequest); err != nil {
				return nil, err
			}

		case model.SingleChoiceQuestion:
			singleChoiceQuestionRequest, err := req.AsPostSingleChoiceQuestionRequest()
			if err != nil {
				return nil, err
			}

			if err := copier.Copy(&dst, &singleChoiceQuestionRequest); err != nil {
				return nil, err
			}

		case model.MultipleChoiceQuestion:
			multipleChoiceQuestionRequest, err := req.AsPostMultipleChoiceQuestionRequest()
			if err != nil {
				return nil, err
			}

			if err := copier.Copy(&dst, &multipleChoiceQuestionRequest); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("unknown question type")
		}

		return dst, nil
	},
}

var questionModelToSchema = copier.TypeConverter{
	SrcType: model.Question{},
	DstType: api.QuestionResponse{},
	Fn: func(src any) (any, error) {
		questionModel, ok := src.(model.Question)

		if !ok {
			return nil, errors.New("src is not a model.Question")
		}

		var dst api.QuestionResponse

		switch questionModel.Type {
		case model.FreeTextQuestion:
			var freeTextQuestion api.FreeTextQuestionResponse

			if err := copier.Copy(&freeTextQuestion, &questionModel); err != nil {
				return nil, err
			}

			if err := dst.FromFreeTextQuestionResponse(freeTextQuestion); err != nil {
				return nil, err
			}

		case model.FreeNumberQuestion:
			var freeNumberQuestion api.FreeNumberQuestionResponse

			if err := copier.Copy(&freeNumberQuestion, &questionModel); err != nil {
				return nil, err
			}

			if err := dst.FromFreeNumberQuestionResponse(freeNumberQuestion); err != nil {
				return nil, err
			}

		case model.SingleChoiceQuestion:
			var singleChoiceQuestion api.SingleChoiceQuestionResponse

			if err := copier.Copy(&singleChoiceQuestion, &questionModel); err != nil {
				return nil, err
			}

			if err := dst.FromSingleChoiceQuestionResponse(singleChoiceQuestion); err != nil {
				return nil, err
			}

		case model.MultipleChoiceQuestion:
			var multipleChoiceQuestion api.MultipleChoiceQuestionResponse

			if err := copier.Copy(&multipleChoiceQuestion, &questionModel); err != nil {
				return nil, err
			}

			if err := dst.FromMultipleChoiceQuestionResponse(multipleChoiceQuestion); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("unknown question type")
		}

		return dst, nil
	},
}
