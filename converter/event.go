package converter

import (
	"errors"

	"github.com/jinzhu/copier"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

var eventSchemaToModel = copier.TypeConverter{
	SrcType: api.EventRequest{},
	DstType: model.Event{},
	Fn: func(src any) (any, error) {
		req, ok := src.(api.EventRequest)

		if !ok {
			return nil, errors.New("src is not an api.EventRequest")
		}

		var dst model.Event

		// リクエストのTypeを取得するため、一旦DurationEventRequestに変換する
		durationEventRequest, err := req.AsDurationEventRequest()

		if err != nil {
			return nil, err
		}

		switch model.EventType(durationEventRequest.Type) {
		case model.EventTypeDuration:
			if err := copier.Copy(&dst, &durationEventRequest); err != nil {
				return nil, err
			}

		case model.EventTypeMoment:
			momentEventRequest, err := req.AsMomentEventRequest()

			if err != nil {
				return nil, err
			}

			if err := copier.Copy(&dst, &momentEventRequest); err != nil {
				return nil, err
			}

		case model.EventTypeOfficial:
			officialEventRequest, err := req.AsOfficialEventRequest()

			if err != nil {
				return nil, err
			}

			if err := copier.Copy(&dst, &officialEventRequest); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("unknown event type")
		}

		return dst, nil
	},
}

var eventModelToSchema = copier.TypeConverter{
	SrcType: model.Event{},
	DstType: api.EventResponse{},
	Fn: func(src any) (any, error) {
		eventModel, ok := src.(model.Event)

		if !ok {
			return nil, errors.New("src is not a model.Event")
		}

		var dst api.EventResponse

		switch eventModel.Type {
		case model.EventTypeDuration:
			var durationEvent api.DurationEventResponse

			if err := copier.Copy(&durationEvent, &eventModel); err != nil {
				return nil, err
			}

			if err := dst.FromDurationEventResponse(durationEvent); err != nil {
				return nil, err
			}

		case model.EventTypeMoment:
			var momentEvent api.MomentEventResponse

			if err := copier.Copy(&momentEvent, &eventModel); err != nil {
				return nil, err
			}

			if err := dst.FromMomentEventResponse(momentEvent); err != nil {
				return nil, err
			}

		case model.EventTypeOfficial:
			var officialEvent api.OfficialEventResponse

			if err := copier.Copy(&officialEvent, &eventModel); err != nil {
				return nil, err
			}

			if err := dst.FromOfficialEventResponse(officialEvent); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("unknown event type")
		}

		return dst, nil
	},
}
