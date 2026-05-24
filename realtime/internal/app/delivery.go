package app

import delivery "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/delivery/http"

type Delivery struct {
	Realtime *delivery.RealtimeHandler
}

func NewDelivery(manager *Manager) *Delivery {
	return &Delivery{
		Realtime: delivery.NewRealtimeHandler(manager.Realtime),
	}
}
