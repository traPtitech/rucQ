package converter

import (
	"errors"

	"github.com/jinzhu/copier"
	"github.com/oapi-codegen/runtime/types"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var campSchemaToModel = copier.TypeConverter{
	SrcType: api.CampRequest{},
	DstType: model.Camp{},
	Fn: func(src any) (any, error) {
		srcCamp, ok := src.(api.CampRequest)

		if !ok {
			return nil, errors.New("src is not an api.CampRequest")
		}

		var dst model.Camp

		if err := copier.Copy(&dst, &srcCamp); err != nil {
			return nil, err
		}

		dst.DateStart = srcCamp.DateStart.Time
		dst.DateEnd = srcCamp.DateEnd.Time

		return dst, nil
	},
}

var campModelToSchema = copier.TypeConverter{
	SrcType: model.Camp{},
	DstType: api.CampResponse{},
	Fn: func(src any) (any, error) {
		srcCamp, ok := src.(model.Camp)

		if !ok {
			return nil, errors.New("src is not a model.Camp")
		}

		var dst api.CampResponse

		if err := copier.Copy(&dst, &srcCamp); err != nil {
			return nil, err
		}

		dst.DateStart = types.Date{Time: srcCamp.DateStart}
		dst.DateEnd = types.Date{Time: srcCamp.DateEnd}

		return dst, nil
	},
}
