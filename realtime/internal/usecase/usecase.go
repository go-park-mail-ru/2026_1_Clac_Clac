package usecase

import (
	"context"
	"fmt"

	pubsub "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/google/uuid"
)

type RealtimeService struct {
	subscriber pubsub.Subscriber[common.BoardUpdateEvent]
}

func NewRealtimeService(subscriber pubsub.Subscriber[common.BoardUpdateEvent]) *RealtimeService {
	return &RealtimeService{
		subscriber: subscriber,
	}
}

func (s *RealtimeService) Subscribe(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error) {
	sub, err := s.subscriber.Subscribe(ctx, pubsub.Channel(boardLink.String()))
	if err != nil {
		return nil, fmt.Errorf("subscriber.Subscribe: %w", err)
	}

	return sub, nil
}
