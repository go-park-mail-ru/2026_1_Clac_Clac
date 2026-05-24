package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/usecase"
)

type Manager struct {
	Realtime *usecase.RealtimeService
}

func NewManager(store *Store, conf *config.Config) *Manager {
	return &Manager{
		Realtime: usecase.NewRealtimeService(store.Subscriber),
	}
}
