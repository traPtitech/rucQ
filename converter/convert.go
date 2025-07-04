package converter

import "github.com/jinzhu/copier"

func Convert[T any](src any) (T, error) {
	var dst T

	err := copier.CopyWithOption(&dst, src, copier.Option{
		Converters: []copier.TypeConverter{
			answerSchemaToModel,
			answerModelToSchema,
			campSchemaToModel,
			campModelToSchema,
			eventSchemaToModel,
			eventModelToSchema,
			postQuestionGroupSchemaToModel,
			putQuestionGroupSchemaToModel,
			questionGroupModelToSchema,
			postQuestionSchemaToModel,
			putQuestionSchemaToModel,
			questionModelToSchema,
		},
	})

	return dst, err
}
