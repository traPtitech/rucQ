package converter

import (
	"errors"

	"github.com/jinzhu/copier"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var rollCallSchemaToModel = copier.TypeConverter{
	SrcType: api.RollCallRequest{},
	DstType: model.RollCall{},
	Fn: func(src any) (any, error) {
		req, ok := src.(api.RollCallRequest)
		if !ok {
			return nil, errors.New("src is not an api.RollCallRequest")
		}

		var dst model.RollCall
		if err := copier.Copy(&dst, &req); err != nil {
			return nil, err
		}

		// Convert subjects from []string to []model.User
		subjects := make([]model.User, len(req.Subjects))
		for i, subjectID := range req.Subjects {
			subjects[i] = model.User{ID: subjectID}
		}
		dst.Subjects = subjects

		return dst, nil
	},
}

var rollCallModelToSchema = copier.TypeConverter{
	SrcType: model.RollCall{},
	DstType: api.RollCallResponse{},
	Fn: func(src any) (any, error) {
		rollCall, ok := src.(model.RollCall)
		if !ok {
			return nil, errors.New("src is not a model.RollCall")
		}

		var dst api.RollCallResponse
		if err := copier.Copy(&dst, &rollCall); err != nil {
			return nil, err
		}

		dst.Id = int(rollCall.ID)

		// Convert subjects from []model.User to []string
		subjects := make([]string, len(rollCall.Subjects))
		for i, subject := range rollCall.Subjects {
			subjects[i] = subject.ID
		}
		dst.Subjects = subjects

		return dst, nil
	},
}
