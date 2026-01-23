package converter

import (
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var answerSchemaToModel = copier.TypeConverter{
	SrcType: api.AnswerRequest{},
	DstType: model.Answer{},
	Fn: func(src any) (any, error) {
		req, ok := src.(api.AnswerRequest)

		if !ok {
			return nil, errors.New("src is not an api.AnswerRequest")
		}

		var dst model.Answer

		if freeTextAnswerRequest, err := req.AsFreeTextAnswerRequest(); err == nil &&
			freeTextAnswerRequest.Type == api.FreeTextAnswerRequestTypeFreeText {
			if err := copier.Copy(&dst, &freeTextAnswerRequest); err != nil {
				return nil, err
			}

			dst.FreeTextContent = &freeTextAnswerRequest.Content
		} else if freeNumberAnswerRequest, err := req.AsFreeNumberAnswerRequest(); err == nil &&
			freeNumberAnswerRequest.Type == api.FreeNumberAnswerRequestTypeFreeNumber {
			if err := copier.Copy(&dst, &freeNumberAnswerRequest); err != nil {
				return nil, err
			}

			content := float64(freeNumberAnswerRequest.Content)
			dst.FreeNumberContent = &content
		} else if singleChoiceAnswerRequest, err := req.AsSingleChoiceAnswerRequest(); err == nil &&
			singleChoiceAnswerRequest.Type == api.SingleChoiceAnswerRequestTypeSingle {
			if err := copier.Copy(&dst, &singleChoiceAnswerRequest); err != nil {
				return nil, err
			}
			dst.SelectedOptions = []model.Option{
				{
					Model: gorm.Model{
						ID: uint(singleChoiceAnswerRequest.OptionId),
					},
				},
			}
		} else if multipleChoiceAnswerRequest, err := req.AsMultipleChoiceAnswerRequest(); err == nil && multipleChoiceAnswerRequest.Type == api.MultipleChoiceAnswerRequestTypeMultiple {
			if err := copier.Copy(&dst, &multipleChoiceAnswerRequest); err != nil {
				return nil, err
			}

			dst.SelectedOptions = make([]model.Option, len(multipleChoiceAnswerRequest.OptionIds))
			for i, optionId := range multipleChoiceAnswerRequest.OptionIds {
				dst.SelectedOptions[i] = model.Option{Model: gorm.Model{ID: uint(optionId)}}
			}
		} else {
			return nil, errors.New("unknown answer type")
		}

		return dst, nil
	},
}

var answerModelToSchema = copier.TypeConverter{
	SrcType: model.Answer{},
	DstType: api.AnswerResponse{},
	Fn: func(src any) (any, error) {
		answerModel, ok := src.(model.Answer)

		if !ok {
			return nil, errors.New("src is not a model.Answer")
		}

		var dst api.AnswerResponse

		switch answerModel.Type {
		case model.FreeTextQuestion:
			var freeTextAnswer api.FreeTextAnswerResponse

			if err := copier.Copy(&freeTextAnswer, &answerModel); err != nil {
				return nil, err
			}

			if answerModel.FreeTextContent == nil {
				return nil, errors.New("FreeTextContent is nil")
			}

			freeTextAnswer.Content = *answerModel.FreeTextContent

			if err := dst.FromFreeTextAnswerResponse(freeTextAnswer); err != nil {
				return nil, err
			}

		case model.FreeNumberQuestion:
			var freeNumberAnswer api.FreeNumberAnswerResponse

			if err := copier.Copy(&freeNumberAnswer, &answerModel); err != nil {
				return nil, err
			}

			if answerModel.FreeNumberContent == nil {
				return nil, errors.New("FreeNumberContent is nil")
			}

			freeNumberAnswer.Content = float32(*answerModel.FreeNumberContent)

			if err := dst.FromFreeNumberAnswerResponse(freeNumberAnswer); err != nil {
				return nil, err
			}

		case model.SingleChoiceQuestion:
			var singleChoiceAnswer api.SingleChoiceAnswerResponse

			if err := copier.Copy(&singleChoiceAnswer, &answerModel); err != nil {
				return nil, err
			}

			if len(answerModel.SelectedOptions) != 1 {
				return nil, errors.New("SelectedOptions length is not 1")
			}

			var selectedOption api.OptionResponse

			if err := copier.Copy(&selectedOption, &answerModel.SelectedOptions[0]); err != nil {
				return nil, err
			}

			singleChoiceAnswer.SelectedOption = selectedOption

			if err := dst.FromSingleChoiceAnswerResponse(singleChoiceAnswer); err != nil {
				return nil, err
			}

		case model.MultipleChoiceQuestion:
			var multipleChoiceAnswer api.MultipleChoiceAnswerResponse

			if err := copier.Copy(&multipleChoiceAnswer, &answerModel); err != nil {
				return nil, err
			}

			var selectedOptions []api.OptionResponse

			if err := copier.Copy(&selectedOptions, &answerModel.SelectedOptions); err != nil {
				return nil, err
			}

			multipleChoiceAnswer.SelectedOptions = selectedOptions

			if err := dst.FromMultipleChoiceAnswerResponse(multipleChoiceAnswer); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("unknown answer type")
		}

		return dst, nil
	},
}
