package router

import (
	"context"

	"github.com/sesopenko/genericpubsub"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/service"
)

type reactionEvent struct {
	rollCallID uint
	data       api.RollCallReactionEvent
}

type Server struct {
	repo                repository.Repository
	notificationService service.NotificationService
	traqService         service.TraqService
	reactionPubSub      *genericpubsub.PubSub[reactionEvent]
	isDev               bool
}

const maxReactionEventBuffer = 100

func NewServer(
	repo repository.Repository,
	notificationService service.NotificationService,
	traqService service.TraqService,
	isDev bool,
) *Server {
	return &Server{
		repo:                repo,
		notificationService: notificationService,
		traqService:         traqService,
		reactionPubSub: genericpubsub.New[reactionEvent](
			context.Background(),
			maxReactionEventBuffer,
		),
		isDev: isDev,
	}
}
