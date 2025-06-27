package converter

import "github.com/jinzhu/copier"

func Convert[T any](src any) (T, error) {
	var dst T

	err := copier.CopyWithOption(&dst, src, copier.Option{
		Converters: []copier.TypeConverter{
			campSchemaToModel,
			campModelToSchema,
			eventSchemaToModel,
			eventModelToSchema,
		},
	})

	return dst, err
}
