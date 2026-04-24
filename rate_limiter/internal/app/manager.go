package app

import (
	limiter "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/service"
)

type Manager struct {
	Limiter *limiter.Service
}

func NewManager(s *Store) *Manager {
	return &Manager{
		Limiter: limiter.NewService(s.Limiter),
	}
}
