package usecase

import (
	"context"
	"fmt"

	pubsub "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/usecase/dto"
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

func (s *RealtimeService) Listen(ctx context.Context, boardLink uuid.UUID) (dto.BoardUpdateInfo, error) {
	sub, err := s.subscriber.Subscribe(ctx, pubsub.Channel(boardLink.String()))
	if err != nil {
		return dto.BoardUpdateInfo{}, fmt.Errorf("pubsub.Subscriber.Subscribe: %w", err)
	}

	select {
	case event, ok := <-sub.C():
		if !ok {
			if err := sub.Err(); err != nil {
				return dto.BoardUpdateInfo{}, fmt.Errorf("pubsub.Subscription: %w", err)
			}

			return dto.BoardUpdateInfo{}, fmt.Errorf("pubsub.Subscription cannot read data from channel")
		}

		return dto.BoardUpdateInfo{
			Type:    string(event.Type),
			Payload: event.Payload,
		}, nil
	case <-ctx.Done():
		return dto.BoardUpdateInfo{}, common.ErrTimeout
	}
}
