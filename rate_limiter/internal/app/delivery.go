package app

import (
	limiter "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/handler"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/rate_limiter"
	"google.golang.org/grpc"
)

type Delivery struct {
	Limiter *limiter.Handler
}

func NewDelivery(m *Manager) *Delivery {
	return &Delivery{
		Limiter: limiter.NewHandler(m.Limiter),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pb.RegisterRateLimiterServiceServer(grpcServer, d.Limiter)
}
