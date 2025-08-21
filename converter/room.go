package converter

import (
	"errors"

	"github.com/jinzhu/copier"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var roomSchemaToModel = copier.TypeConverter{
	SrcType: api.RoomRequest{},
	DstType: model.Room{},
	Fn: func(src any) (any, error) {
		srcRoom, ok := src.(api.RoomRequest)

		if !ok {
			return nil, errors.New("src is not an api.RoomRequest")
		}

		var dst model.Room

		if err := copier.Copy(&dst, &srcRoom); err != nil {
			return nil, err
		}

		dst.Members = make([]model.User, len(srcRoom.MemberIds))

		for i, memberID := range srcRoom.MemberIds {
			dst.Members[i] = model.User{
				ID: memberID,
			}
		}

		return dst, nil
	},
}
