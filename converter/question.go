package converter

import (
	"encoding/json"
	"errors"

	"github.com/jinzhu/copier"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var questionSchemaToModel = copier.TypeConverter{
	SrcType: api.QuestionRequest{},
	DstType: model.Question{},
	Fn: func(src any) (any, error) {
		req, ok := src.(api.QuestionRequest)

		if !ok {
			return nil, errors.New("src is not an api.QuestionRequest")
		}

		var dst model.Question

		// まず、Typeを判定するため、一旦Marshalしてから型を判定する
		reqBytes, err := req.MarshalJSON()
		if err != nil {
			return nil, err
		}

		var typeChecker struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(reqBytes, &typeChecker); err != nil {
			return nil, err
		}

		switch model.QuestionType(typeChecker.Type) {
		case model.FreeTextQuestion:
			freeTextQuestionRequest, err := req.AsFreeTextQuestionRequest()
			if err != nil {
				return nil, err
			}

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
			singleChoiceQuestionRequest, err := req.AsSingleChoiceQuestionRequest()
			if err != nil {
				return nil, err
			}

			if err := copier.Copy(&dst, &singleChoiceQuestionRequest); err != nil {
				return nil, err
			}

		case model.MultipleChoiceQuestion:
			multipleChoiceQuestionRequest, err := req.AsMultipleChoiceQuestionRequest()
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
