package router

import (
	"context"

	"github.com/sesopenko/genericpubsub"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/service/notification"
	"github.com/traPtitech/rucQ/service/traq"
)

type reactionEvent struct {
	rollCallID uint
	data       api.RollCallReactionEvent
}

type Server struct {
	repo                repository.Repository
	notificationService notification.NotificationService
	traqService         traq.TraqService
	reactionPubSub      *genericpubsub.PubSub[reactionEvent]
	isDev               bool
}

const maxReactionEventBuffer = 100

func NewServer(
	ctx context.Context,
	repo repository.Repository,
	notificationService notification.NotificationService,
	traqService traq.TraqService,
	isDev bool,
) *Server {
	return &Server{
		repo:                repo,
		notificationService: notificationService,
		traqService:         traqService,
		reactionPubSub: genericpubsub.New[reactionEvent](
			ctx,
			maxReactionEventBuffer,
		),
		isDev: isDev,
	}
}
